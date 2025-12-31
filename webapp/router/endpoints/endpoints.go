package endpoints

// Page routes
const (
	HomePage  = "/"
	LobbyPage = "/lobby"
	GamePage  = "/game"
)

// SSE stream routes
const (
	LobbyConnect = "/lobby/connect" // SSE stream for lobby updates
	GameConnect  = "/game/connect"  // SSE stream for game state updates
)

// Action routes
const (
	LobbyJoin  = "/lobby/join"  // POST - join lobby with name
	LobbyStart = "/lobby/start" // POST - start game (first player only)
	GameInput  = "/game/input"  // POST - submit player choice
)

// SSE Event names
const (
	SSEEventLobby = "lobby"     // Lobby player list update
	SSEEventGame  = "gamestate" // Full game board update
)

// Form field names
const (
	FieldPlayerName = "player_name"
	FieldChoice     = "choice"
)

// Element IDs for HTMX targeting
const (
	IDLobbyContent = "lobby-content"
	IDGameBoard    = "game-board"
	IDPlayerList   = "player-list"
)
