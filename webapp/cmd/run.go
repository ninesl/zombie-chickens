package cmd

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ninesl/zombie-chickens/webapp/config"
)

// Run starts the server with the given config and router.
func Run(c *config.Config, r *chi.Mux) error {
	server := &http.Server{
		Handler:      r,
		Addr:         c.Server.Addr,
		ReadTimeout:  c.Server.ReadTimeout,
		WriteTimeout: c.Server.WriteTimeout,
		IdleTimeout:  c.Server.IdleTimeout,
	}

	log.Printf("server is running on http://%s\n", c.Server.Addr)

	return server.ListenAndServe()
}
