package webapp

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ninesl/zombie-chickens/webapp/cmd"
	"github.com/ninesl/zombie-chickens/webapp/config"
	"github.com/ninesl/zombie-chickens/webapp/router"
)

// RunServer starts the web server for Zombie Chickens
func RunServer() {
	// Create router
	r := chi.NewRouter()

	// Create config
	c := config.NewConfig()

	// Set middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Get router with all routes
	router.GetRoutes(r)

	// Run server
	if err := cmd.Run(c, r); err != nil {
		log.Fatal(err)
	}
}
