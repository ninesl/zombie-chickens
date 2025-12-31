package lobby

import (
	"github.com/go-chi/chi/v5"
	"github.com/ninesl/zombie-chickens/webapp/router/endpoints"
)

// Routes registers lobby-related routes
func Routes(r chi.Router) {
	r.Get(endpoints.HomePage, HandleLobbyPage())
	r.Get(endpoints.LobbyConnect, HandleLobbyConnect())
	r.Post(endpoints.LobbyJoin, HandleLobbyJoin())
	r.Post(endpoints.LobbyStart, HandleLobbyStart())
}
