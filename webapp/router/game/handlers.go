package game

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ninesl/zombie-chickens/webapp/middleware"
	"github.com/ninesl/zombie-chickens/webapp/router/endpoints"
	"github.com/ninesl/zombie-chickens/webapp/state"
	"github.com/ninesl/zombie-chickens/webapp/ui/components"
	"github.com/ninesl/zombie-chickens/webapp/ui/pages"
)

// HandleGamePage serves the main game page
func HandleGamePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := state.GetSession()

		// If game not started, redirect to lobby
		if !session.IsStarted() {
			http.Redirect(w, r, endpoints.HomePage, http.StatusSeeOther)
			return
		}

		// Check player is in game
		sessionID := middleware.GetSessionID(r.Context())
		_, playerIdx := session.GetPlayerBySession(sessionID)
		if playerIdx == -1 {
			http.Redirect(w, r, endpoints.HomePage, http.StatusSeeOther)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		pages.GamePage().Render(r.Context(), w)
	}
}

// HandleGameConnect handles SSE connections for game updates
func HandleGameConnect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := state.GetSession()

		// Check game has started
		if !session.IsStarted() {
			http.Error(w, "game not started", http.StatusBadRequest)
			return
		}

		// Get session ID from context (set by middleware)
		sessionID := middleware.GetSessionID(r.Context())
		if sessionID == "" {
			http.Error(w, "no session", http.StatusUnauthorized)
			return
		}

		// Verify player is in game
		_, playerIdx := session.GetPlayerBySession(sessionID)
		if playerIdx == -1 {
			http.Error(w, "not in game", http.StatusForbidden)
			return
		}

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		// Create client
		client := &state.Client{
			SessionID: sessionID,
			PlayerIdx: playerIdx,
			Send:      make(chan []byte, 10),
			Done:      make(chan struct{}),
		}

		session.AddGameClient(client)
		log.Printf("[SSE] Client %d connected (session: %s...)", playerIdx, sessionID[:8])
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[SSE] PANIC in SSE handler: %v", r)
			}
			session.RemoveGameClient(client)
			// Close Done channel if not already closed
			select {
			case <-client.Done:
				// Already closed
			default:
				close(client.Done)
			}
			log.Printf("[SSE] Client %d disconnected", playerIdx)
		}()

		// Send initial state
		initialData := renderGameBoard(session)
		sseMsg := state.FormatSSE(endpoints.SSEEventGame, initialData)
		if _, err := w.Write(sseMsg); err != nil {
			log.Printf("[SSE] Error writing initial state: %v", err)
			return
		}
		flusher.Flush()
		log.Printf("[SSE] Sent initial state to client %d", playerIdx)

		// Keepalive ticker - send comment every 15 seconds to prevent timeout
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		// Stream updates
		ctx := r.Context()
		for {
			select {
			case data := <-client.Send:
				if _, err := w.Write(data); err != nil {
					log.Printf("[SSE] Error writing to client %d: %v", playerIdx, err)
					return
				}
				flusher.Flush()
				log.Printf("[SSE] Sent update to client %d (%d bytes)", playerIdx, len(data))
			case <-ticker.C:
				// SSE comment as keepalive (ignored by client but keeps connection alive)
				if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
					log.Printf("[SSE] Keepalive failed for client %d: %v", playerIdx, err)
					return
				}
				flusher.Flush()
			case <-ctx.Done():
				log.Printf("[SSE] Context done for client %d: %v", playerIdx, ctx.Err())
				return
			case <-client.Done:
				log.Printf("[SSE] Client %d done channel closed", playerIdx)
				return
			}
		}
	}
}

// HandleGameInput handles player input submission
func HandleGameInput() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := state.GetSession()

		if !session.IsStarted() {
			http.Error(w, "game not started", http.StatusBadRequest)
			return
		}

		sessionID := middleware.GetSessionID(r.Context())
		if sessionID == "" {
			http.Error(w, "no session", http.StatusUnauthorized)
			return
		}

		// Parse choice
		choiceStr := r.FormValue(endpoints.FieldChoice)
		choice, err := strconv.Atoi(choiceStr)
		if err != nil {
			http.Error(w, "invalid choice", http.StatusBadRequest)
			return
		}

		// Submit input
		log.Printf("[INPUT] Player submitting choice: %d", choice)
		err = session.SubmitInput(sessionID, choice)
		if err != nil {
			log.Printf("[INPUT] Error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("[INPUT] Choice accepted, broadcasting...")

		// Broadcast updated state to all clients
		ctx := context.Background()
		session.BroadcastGame(ctx, func() []byte {
			log.Printf("[INPUT] Rendering game board for broadcast...")
			data := renderGameBoard(session)
			log.Printf("[INPUT] Rendered %d bytes", len(data))
			return state.FormatSSE(endpoints.SSEEventGame, data)
		})
		log.Printf("[INPUT] Broadcast complete")

		// Return success (client will get update via SSE)
		w.WriteHeader(http.StatusOK)
	}
}

// renderGameBoard renders the game board HTML
func renderGameBoard(session *state.GameSession) []byte {
	return components.RenderGameBoard(session)
}
