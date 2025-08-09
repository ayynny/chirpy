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
	"time"

	"github.com/ayynny/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"email"`
	UserID    string    `json:"user_id`
}

// check length of body and censor bad words
func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	type request struct { // request body
		Body   string `json:"body"`
		UserID string `json:"user_id`
	}

	decoder := json.NewDecoder(r.Body)      // create a new decoder to read the request json data
	requestInstance := request{}            // create instance of the request body
	err := decoder.Decode(&requestInstance) // pass &requestInstance to modify original struct, not the copy (requestInstance).
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
