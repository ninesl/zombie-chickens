package lobby

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/ninesl/zombie-chickens/webapp/middleware"
	"github.com/ninesl/zombie-chickens/webapp/router/endpoints"
	"github.com/ninesl/zombie-chickens/webapp/state"
	"github.com/ninesl/zombie-chickens/webapp/ui/pages"
)

// HandleLobbyPage serves the main lobby page
func HandleLobbyPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := state.GetSession()
		sessionID := middleware.GetSessionID(r.Context())

		// Check if already in game and game started - redirect to game
		if session.IsStarted() {
			_, playerIdx := session.GetPlayerBySession(sessionID)
			if playerIdx >= 0 {
				http.Redirect(w, r, endpoints.GamePage, http.StatusSeeOther)
				return
			}
		}

		_, playerIdx := session.GetPlayerBySession(sessionID)
		joined := playerIdx >= 0

		w.Header().Set("Content-Type", "text/html")
		pages.LobbyPage(session, sessionID, joined).Render(r.Context(), w)
	}
}

// HandleLobbyConnect handles SSE connections for lobby updates
func HandleLobbyConnect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := state.GetSession()
		sessionID := middleware.GetSessionID(r.Context())

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
			PlayerIdx: -1,
			Send:      make(chan []byte, 10),
			Done:      make(chan struct{}),
		}

		session.AddLobbyClient(client)
		defer func() {
			session.RemoveLobbyClient(client)
			// Close Done channel if not already closed
			select {
			case <-client.Done:
				// Already closed
			default:
				close(client.Done)
			}
		}()

		// Send initial state
		_, playerIdx := session.GetPlayerBySession(sessionID)
		joined := playerIdx >= 0
		initialData := renderLobbyContent(session, sessionID, joined)
		sseMsg := state.FormatSSE(endpoints.SSEEventLobby, initialData)
		w.Write(sseMsg)
		flusher.Flush()

		// Keepalive ticker - send comment every 15 seconds to prevent timeout
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		// Stream updates
		ctx := r.Context()
		for {
			select {
			case data := <-client.Send:
				w.Write(data)
				flusher.Flush()
			case <-ticker.C:
				// SSE comment as keepalive (ignored by client but keeps connection alive)
				w.Write([]byte(": keepalive\n\n"))
				flusher.Flush()
			case <-ctx.Done():
				return
			case <-client.Done:
				return
			}
		}
	}
}

// HandleLobbyJoin handles player joining the lobby
func HandleLobbyJoin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := state.GetSession()
		sessionID := middleware.GetSessionID(r.Context())

		name := r.FormValue(endpoints.FieldPlayerName)
		if name == "" {
			name = "Guest"
		}

		// Add player
		_, err := session.AddPlayer(sessionID, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Broadcast update to all lobby clients
		ctx := context.Background()
		session.BroadcastLobby(ctx, func() []byte {
			// Each client gets personalized view based on their session
			// For simplicity, we send a generic update and let clients know to refresh
			data := renderLobbyContent(session, sessionID, true)
			return state.FormatSSE(endpoints.SSEEventLobby, data)
		})

		// Return the updated lobby content for this player
		w.Header().Set("Content-Type", "text/html")
		pages.LobbyContent(session, sessionID, true).Render(r.Context(), w)
	}
}

// HandleLobbyStart handles starting the game
func HandleLobbyStart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := state.GetSession()
		sessionID := middleware.GetSessionID(r.Context())

		// Only first player can start
		if !session.IsFirstPlayer(sessionID) {
			http.Error(w, "only first player can start", http.StatusForbidden)
			return
		}

		// Start the game
		err := session.StartGame()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Broadcast to lobby clients that game started
		ctx := context.Background()
		session.BroadcastLobby(ctx, func() []byte {
			data := renderLobbyContent(session, sessionID, true)
			return state.FormatSSE(endpoints.SSEEventLobby, data)
		})

		// Return redirect info
		w.Header().Set("Content-Type", "text/html")
		pages.LobbyContent(session, sessionID, true).Render(r.Context(), w)
	}
}

func renderLobbyContent(session *state.GameSession, sessionID string, joined bool) []byte {
	var buf bytes.Buffer
	pages.LobbyContent(session, sessionID, joined).Render(context.Background(), &buf)
	return buf.Bytes()
}
