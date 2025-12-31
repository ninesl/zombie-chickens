package components

import (
	"bytes"
	"context"

	"github.com/ninesl/zombie-chickens/webapp/state"
)

// RenderGameBoard renders the game board to bytes
func RenderGameBoard(session *state.GameSession) []byte {
	props := BoardProps{
		Game:         session.Game(),
		PendingInput: session.PendingInput(),
		GameOver:     session.IsGameOver(),
	}

	var buf bytes.Buffer
	GameBoard(props).Render(context.Background(), &buf)
	return buf.Bytes()
}

// RenderLobbyContent renders the lobby content to bytes
func RenderLobbyContent(session *state.GameSession, sessionID string, joined bool) []byte {
	var buf bytes.Buffer
	// Note: We need to import pages package, but that creates a cycle
	// So we'll handle this differently in the handlers
	return buf.Bytes()
}
