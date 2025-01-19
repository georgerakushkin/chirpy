package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/georgerakushkin/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// apiConfig holds the application configuration and state
type apiConfig struct {
	fileserverHits atomic.Int32      // Counter for file server hits
	db             *database.Queries // Database connection and queries
	platform       string            // Platform environment (dev/prod)
}

func main() {
	const filepathRoot = "." // Root directory for serving files
	const port = "8080"      // Server port number

	// Load environment variables from .env file
	godotenv.Load()

	// Get database URL from environment
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	// Get platform environment from environment variables
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	// Initialize database connection
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	// Initialize API configuration
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
	}

	// Create new HTTP multiplexer
	mux := http.NewServeMux()

	// Set up file server with metrics middleware
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)

	// API endpoints
	mux.HandleFunc("GET /api/healthz", handlerReadiness)              // Health check endpoint
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate) // Validate chirp content
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)      // Create new user

	// Admin endpoints
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)    // Reset database (dev only)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics) // View metrics

	// Initialize HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
