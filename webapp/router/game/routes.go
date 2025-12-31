package game

import (
	"github.com/go-chi/chi/v5"
	"github.com/ninesl/zombie-chickens/webapp/router/endpoints"
)

// Routes registers game-related routes
func Routes(r chi.Router) {
	r.Get(endpoints.GamePage, HandleGamePage())
	r.Get(endpoints.GameConnect, HandleGameConnect())
	r.Post(endpoints.GameInput, HandleGameInput())
}
