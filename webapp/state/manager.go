package state

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/ninesl/zombie-chickens/zcgame"
)

// PlayerInfo holds lobby/session info for a player
type PlayerInfo struct {
	SessionID string
	Name      string
	Index     int // -1 until game starts
}

// Client represents a connected SSE client
type Client struct {
	SessionID string
	PlayerIdx int           // -1 for lobby clients
	Send      chan []byte   // Channel to send SSE data
	Done      chan struct{} // Closed when client disconnects
}

// GameSession holds the game state and connected clients
type GameSession struct {
	mu           sync.RWMutex
	game         zcgame.GameView
	pendingInput *zcgame.PlayerInputNeeded
	players      []PlayerInfo
	started      bool
	gameOver     bool

	// SSE clients
	lobbyClients []*Client
	gameClients  []*Client
}

// Global session - single game for now
var (
	session     *GameSession
	sessionOnce sync.Once
)

// GetSession returns the global game session, creating it if needed
func GetSession() *GameSession {
	sessionOnce.Do(func() {
		session = &GameSession{
			players:      make([]PlayerInfo, 0, 4),
			lobbyClients: make([]*Client, 0),
			gameClients:  make([]*Client, 0),
		}
	})
	return session
}

// ResetSession resets the global session (for testing/new game)
func ResetSession() {
	session = &GameSession{
		players:      make([]PlayerInfo, 0, 4),
		lobbyClients: make([]*Client, 0),
		gameClients:  make([]*Client, 0),
	}
}

// AddPlayer adds a player to the lobby
func (gs *GameSession) AddPlayer(sessionID, name string) (int, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.started {
		return -1, ErrGameAlreadyStarted
	}

	if len(gs.players) >= 4 {
		return -1, ErrGameFull
	}

	// Check if session already joined
	for i, p := range gs.players {
		if p.SessionID == sessionID {
			return i, nil // Already joined
		}
	}

	// Make name unique by appending a number if needed
	uniqueName := gs.makeUniqueName(name)

	idx := len(gs.players)
	gs.players = append(gs.players, PlayerInfo{
		SessionID: sessionID,
		Name:      uniqueName,
		Index:     idx,
	})

	return idx, nil
}

// makeUniqueName ensures the name is unique among current players
func (gs *GameSession) makeUniqueName(name string) string {
	// Check if name already exists
	exists := func(n string) bool {
		for _, p := range gs.players {
			if p.Name == n {
				return true
			}
		}
		return false
	}

	if !exists(name) {
		return name
	}

	// Append numbers until unique
	for i := 2; i <= 10; i++ {
		candidate := fmt.Sprintf("%s%d", name, i)
		if !exists(candidate) {
			return candidate
		}
	}

	// Fallback: use session ID prefix
	return fmt.Sprintf("%s_%s", name, gs.players[len(gs.players)].SessionID[:4])
}

// GetPlayerBySession returns player info and their current game index by session ID.
// The index is looked up dynamically from the game state by matching the player's name,
// so it remains correct even after other players are eliminated and indices shift.
// Returns (nil, -1) if session not found, or (playerInfo, -1) if player was eliminated.
func (gs *GameSession) GetPlayerBySession(sessionID string) (*PlayerInfo, int) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// Find the player info by session ID
	var playerInfo *PlayerInfo
	for i := range gs.players {
		if gs.players[i].SessionID == sessionID {
			playerInfo = &gs.players[i]
			break
		}
	}
	if playerInfo == nil {
		return nil, -1
	}

	// If game hasn't started, return the original index
	if !gs.started || gs.game.PlayerCount() == 0 {
		return playerInfo, playerInfo.Index
	}

	// Look up current index by name in the game state
	idx := gs.game.PlayerIdxByName(playerInfo.Name)
	return playerInfo, idx
}

// Players returns a copy of the player list
func (gs *GameSession) Players() []PlayerInfo {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	result := make([]PlayerInfo, len(gs.players))
	copy(result, gs.players)
	return result
}

// PlayerCount returns number of players
func (gs *GameSession) PlayerCount() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return len(gs.players)
}

// IsStarted returns whether the game has started
func (gs *GameSession) IsStarted() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.started
}

// IsGameOver returns whether the game has ended
func (gs *GameSession) IsGameOver() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.gameOver
}

// IsFirstPlayer checks if sessionID is the first player (can start game)
func (gs *GameSession) IsFirstPlayer(sessionID string) bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if len(gs.players) == 0 {
		return false
	}
	return gs.players[0].SessionID == sessionID
}

// StartGame starts the game with current players
func (gs *GameSession) StartGame() error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.started {
		return ErrGameAlreadyStarted
	}

	if len(gs.players) < 1 {
		return ErrNotEnoughPlayers
	}

	// Collect player names
	names := make([]string, len(gs.players))
	for i, p := range gs.players {
		names[i] = p.Name
	}

	// Create the game
	game, err := zcgame.CreateNewGame(names...)
	if err != nil {
		return err
	}

	gs.game = game
	gs.started = true

	// Start the game - get first input needed
	_, inputNeeded := gs.game.ContinueDay()
	gs.pendingInput = inputNeeded

	return nil
}

// Game returns the game view (nil if not started)
func (gs *GameSession) Game() zcgame.GameView {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.game
}

// PendingInput returns what input is currently needed
func (gs *GameSession) PendingInput() *zcgame.PlayerInputNeeded {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.pendingInput
}

// SubmitInput processes player input and advances game state
func (gs *GameSession) SubmitInput(sessionID string, choice int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if !gs.started {
		return ErrGameNotStarted
	}

	if gs.gameOver {
		return ErrGameOver
	}

	// Validate it's this player's turn to provide input
	// Use ActiveInputPlayerIdx which handles event discards where a different player
	// than CurrentPlayerIdx needs to provide input
	activeIdx := gs.game.ActiveInputPlayerIdx()
	playerInfo, playerIdx := gs.getPlayerBySessionLocked(sessionID)
	if playerInfo == nil {
		return ErrPlayerNotFound
	}
	if playerIdx != activeIdx {
		return ErrNotYourTurn
	}

	// Validate choice
	if gs.pendingInput == nil {
		return ErrNoInputNeeded
	}
	if !isValidChoice(choice, gs.pendingInput.ValidChoices) {
		return ErrInvalidChoice
	}

	// Process input
	gameContinues, inputNeeded := gs.game.ContinueAfterInput(choice)
	gs.pendingInput = inputNeeded

	// If no input needed but game continues, advance to next phase
	// This handles transitions like: day -> night, player1 -> player2, etc.
	for gameContinues && inputNeeded == nil {
		gameContinues, inputNeeded = gs.game.ContinueDay()
		gs.pendingInput = inputNeeded
	}

	// Game is only over if gameContinues is false AND no more input is needed
	// (false, nil) = game over
	if !gameContinues && inputNeeded == nil {
		gs.gameOver = true
	}

	return nil
}

// getPlayerBySessionLocked returns player info (must hold lock)
func (gs *GameSession) getPlayerBySessionLocked(sessionID string) (*PlayerInfo, int) {
	for i, p := range gs.players {
		if p.SessionID == sessionID {
			return &gs.players[i], i
		}
	}
	return nil, -1
}

func isValidChoice(choice int, valid []int) bool {
	for _, v := range valid {
		if choice == v {
			return true
		}
	}
	return false
}

// --- SSE Client Management ---

// AddLobbyClient registers an SSE client for lobby updates
// Removes any existing client with the same session ID first to prevent duplicates
func (gs *GameSession) AddLobbyClient(client *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Remove existing client with same session ID
	for i := len(gs.lobbyClients) - 1; i >= 0; i-- {
		if gs.lobbyClients[i].SessionID == client.SessionID {
			// Signal old client to close
			select {
			case <-gs.lobbyClients[i].Done:
				// Already closed
			default:
				close(gs.lobbyClients[i].Done)
			}
			gs.lobbyClients = append(gs.lobbyClients[:i], gs.lobbyClients[i+1:]...)
		}
	}

	gs.lobbyClients = append(gs.lobbyClients, client)
}

// RemoveLobbyClient removes an SSE client
func (gs *GameSession) RemoveLobbyClient(client *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	for i, c := range gs.lobbyClients {
		if c == client {
			gs.lobbyClients = append(gs.lobbyClients[:i], gs.lobbyClients[i+1:]...)
			return
		}
	}
}

// AddGameClient registers an SSE client for game updates
// Removes any existing client with the same session ID first to prevent duplicates
func (gs *GameSession) AddGameClient(client *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Remove existing client with same session ID
	for i := len(gs.gameClients) - 1; i >= 0; i-- {
		if gs.gameClients[i].SessionID == client.SessionID {
			// Signal old client to close
			select {
			case <-gs.gameClients[i].Done:
				// Already closed
			default:
				close(gs.gameClients[i].Done)
			}
			gs.gameClients = append(gs.gameClients[:i], gs.gameClients[i+1:]...)
		}
	}

	gs.gameClients = append(gs.gameClients, client)
}

// RemoveGameClient removes an SSE client
func (gs *GameSession) RemoveGameClient(client *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	for i, c := range gs.gameClients {
		if c == client {
			gs.gameClients = append(gs.gameClients[:i], gs.gameClients[i+1:]...)
			return
		}
	}
}

// BroadcastLobby sends lobby update to all connected lobby clients
func (gs *GameSession) BroadcastLobby(ctx context.Context, render func() []byte) {
	gs.mu.RLock()
	clients := make([]*Client, len(gs.lobbyClients))
	copy(clients, gs.lobbyClients)
	gs.mu.RUnlock()

	data := render()
	for _, client := range clients {
		select {
		case client.Send <- data:
		case <-client.Done:
		case <-ctx.Done():
			return
		default:
			// Client buffer full, skip
		}
	}
}

// BroadcastGame sends game state update to all connected game clients
func (gs *GameSession) BroadcastGame(ctx context.Context, render func() []byte) {
	gs.mu.RLock()
	clients := make([]*Client, len(gs.gameClients))
	copy(clients, gs.gameClients)
	gs.mu.RUnlock()

	data := render()
	for _, client := range clients {
		select {
		case client.Send <- data:
		case <-client.Done:
		case <-ctx.Done():
			return
		default:
			// Client buffer full, skip
		}
	}
}

// FormatSSE formats data as an SSE message
func FormatSSE(event string, data []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("event: ")
	buf.WriteString(event)
	buf.WriteString("\n")

	// Split data by newlines and prefix each with "data: "
	lines := bytes.Split(data, []byte("\n"))
	for i, line := range lines {
		buf.WriteString("data: ")
		buf.Write(line)
		if i < len(lines)-1 {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("\n\n")
	return buf.Bytes()
}
