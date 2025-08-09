package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Simple health check handler
func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Delete all users in the database
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	dbPlatform := os.Getenv("PLATFORM")
	if dbPlatform == "" {
		log.Fatal("PLATFORM must be set")
	}

	if dbPlatform != "dev" {
		w.WriteHeader(403)
		return
	}

	err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("Couldn't delete all users: %s", err)
		w.WriteHeader(400)
		return
	}
	log.Printf("Successfully deleted users")
	w.WriteHeader(200)

}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	htmlTemplate := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	sprintfOutput := fmt.Sprintf(htmlTemplate, cfg.fileserverHits.Load())
	w.Write([]byte(sprintfOutput))
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	requestInstance := request{}
	err := decoder.Decode(&requestInstance)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	createdUser, err := cfg.db.CreateUser(r.Context(), requestInstance.Email)
	if err != nil {
		log.Printf("Cannot create user: %s", err)
	}
	params := User{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	dat, err := json.Marshal(params)
	if err != nil {
		log.Fatalf("Cannot marshal user into JSON: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}
