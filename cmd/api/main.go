package main

import (
	"log"

	"github.com/RakibulBh/homeserver-vitals/internal/env"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	environment := env.GetString("ENV", "development")
	if environment != "development" {
		log.Printf("Running in %s environment", environment)
	} else {
		log.Printf("Running in development environment")
	}

	// Load configuration
	cfg := config{
		addr: ":" + env.GetString("PORT", "2000"),
		env:  environment,
	}

	app := &application{
		config: cfg,
	}

	// Prepare server
	log.Printf("Setting up HTTP server on %s", cfg.addr)
	mux := app.serve()

	log.Fatal(app.run(mux))

	// Start listening for requests
	log.Printf("Starting HTTP server, listening on %s", cfg.addr)
}
