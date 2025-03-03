package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/takedown-observer/backend/api"
	"github.com/takedown-observer/backend/db"
	"github.com/takedown-observer/backend/router"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("SQLITE_DB_PATH")
	if dbPath == "" {
		dbPath = "takedowns.db" // Default to local file if env var not set
	}

	// Ensure the database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize database
	log.Printf("Initializing database at: %s", dbPath)
	database, err := db.New(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create API handler
	handler := api.NewHandler(database)

	// Set up router
	r := router.New(handler)

	// Start server
	log.Printf("Server starting on :80...")
	if err := http.ListenAndServe(":80", r); err != nil {
		log.Fatal(err)
	}
}
