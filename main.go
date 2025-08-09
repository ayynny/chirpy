package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ayynny/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(dbConn)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       "",
	}

	mux := http.NewServeMux()

	// Create file server handler
	fileServer := http.FileServer(http.Dir("."))
	fileServerStripped := http.StripPrefix("/app", fileServer)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerStripped))

	// Register other endpoints
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("GET /api/healthz", myHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)

	// Start server
	s := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	fmt.Println("Server listening on http://localhost:8080")
	if err := s.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}
}
