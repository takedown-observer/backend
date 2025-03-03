package router

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/takedown-observer/backend/api"
)

// New creates and configures a new router
func New(handler *api.Handler) http.Handler {
	router := mux.NewRouter()

	// API endpoints
	router.HandleFunc("/api/report", handler.ReportHandler).Methods("POST")
	router.HandleFunc("/api/accounts", handler.GetAccountsHandler).Methods("GET")
	router.HandleFunc("/api/download", handler.DownloadCSVHandler).Methods("GET")

	// Serve static files
	staticFiles := http.FileServer(http.Dir("static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFiles))

	// SPA route handler
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First check if it's an API route
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// List of valid routes
		validRoutes := map[string]bool{
			"/":             true,
			"/dashboard":    true,
			"/about":        true,
			"/related-work": true,
		}

		// Check if it's a valid route
		if validRoutes[r.URL.Path] {
			http.ServeFile(w, r, "static/index.html")
			return
		}

		// Return 404 for invalid routes
		http.NotFound(w, r)
	})

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"https://twitter.com",
			"https://x.com",
			"http://localhost:8080",
		},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	return c.Handler(router)
}
