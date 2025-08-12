package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ayynny/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

// check length of body and censor bad words
func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	requestInstance := request{}
	err := decoder.Decode(&requestInstance)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	// Parse UUID
	uID, err := uuid.Parse(requestInstance.UserID)
	if err != nil {
		log.Printf("Error parsing userID: %v", err)
		w.WriteHeader(400)
		return
	}

	// Validate length
	if len(requestInstance.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "Chirp is too long"})
		return
	}

	// Clean bad words BEFORE saving to database
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	cleanedBody := cleanBadWords(requestInstance.Body, badWords)

	// Create chirp in database
	chirpParams := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: uID,
	}

	createdChirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		log.Printf("Cannot create chirp: %v", err)
		w.WriteHeader(500)
		return
	}

	// Return proper chirp response
	response := Chirp{
		ID:        createdChirp.ID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
		Body:      createdChirp.Body,
		UserID:    createdChirp.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(response)
}

func cleanBadWords(body string, badWords []string) string {
	splitOrg := strings.Split(body, " ")
	for i := range splitOrg {
		lowered := strings.ToLower(splitOrg[i])
		for _, word := range badWords {
			if lowered == word {
				splitOrg[i] = "****"
			}
		}
	}
	return strings.Join(splitOrg, " ")
}
