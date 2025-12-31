package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ninesl/zombie-chickens/webapp/middleware"
	"github.com/ninesl/zombie-chickens/webapp/router/game"
	"github.com/ninesl/zombie-chickens/webapp/router/lobby"
)

// GetRoutes configures all routes on the mux
func GetRoutes(m *chi.Mux) {
	// Session middleware for all routes
	m.Use(middleware.Session)

	// Static assets from webapp/assets
	fs := http.FileServer(http.Dir("webapp/assets"))
	m.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	// Lobby routes (including home page)
	lobby.Routes(m)

	// Game routes
	game.Routes(m)

	m.NotFound(handle404())
}

func handle404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found"))
	}
}
