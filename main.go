package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/ayynny/chirpy/internal/database"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux() // 1. create a new "router" (ServeMux) which decides what code (handler) to run for various paths

	s := http.Server{ // 2. set up http.Server struct
		Handler: mux,
		Addr:    ":8080",
	}

	fileServer := http.FileServer(http.Dir(".")) // returns a http.Handler

	mux.Handle("/app/", http.StripPrefix("/app", fileServer)) //register handler before server starts serving requests

	mux.HandleFunc("/healthz", myHandler)

	err := s.ListenAndServe()
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	type responseErr struct { // response body if there's an error
		Err string `json:"error"`
	}

	type bodyClean struct {
		BodyToClean string `json:"cleaned_body"`
	}

	responseErrInstance := responseErr{}
	bodyCleanInstance := bodyClean{}

	w.Header().Set("Content-Type", "application/json")

	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	// check for length of request's body
	if len(requestInstance.Body) > 140 {
		responseErrInstance.Err = "Chirp is too long"
		w.WriteHeader(400)
		dat, err := json.Marshal(responseErrInstance) // json.Marshal converts Go data structures (like structs) into compact JSON byte arrays
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write(dat) // write out the error, in json format
		return
	} else {
		w.WriteHeader(200)
		bodyCleanInstance.BodyToClean = requestInstance.Body

		splitOrg := strings.Split(bodyCleanInstance.BodyToClean, " ")
		for i := range splitOrg {
			lowered := strings.ToLower(splitOrg[i])
			for _, word := range badWords {
				if lowered == word {
					splitOrg[i] = "****"
				}
			}
		}

		if err != nil {
			log.Printf("Cannot create chirp: %v", err)
		}

		bodyCleanInstance.BodyToClean = strings.Join(splitOrg, " ")
		dat, err := json.Marshal(bodyCleanInstance)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write(dat)
		return
	}

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
