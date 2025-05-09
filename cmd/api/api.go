package main

import (
	"net/http"
	"time"

	"github.com/RakibulBh/homeserver-vitals/internal/env"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

type application struct {
	config config
}

type config struct {
	addr string
	env  string
}

func (app *application) serve() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("FRONTEND_URL", "http://localhost:3000")},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Healthcheck
	r.Get("/health", app.healthCheck)

	// initiate SSE
	r.Get("/sse", app.initiateSSE)

	// Get Vitals
	r.Get("/vitals", app.printVitals)

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := http.Server{
		Addr:              app.config.addr,
		Handler:           mux,
		WriteTimeout:      80 * time.Second,
		ReadTimeout:       80 * time.Second,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 50 * time.Second,
	}

	return srv.ListenAndServe()
}
