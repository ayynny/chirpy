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

type ChirpRequest struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

/*
	POST /api/chirps

If a chirp is valid, meaning cleaned bad words and within character limit,
it is saved in the database with an id, created_at, updated_at, body, and user_id that references
the id of the user who created the chirp.
*/
func (cfg *apiConfig) CreateChirpHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirpRequestInstance := ChirpRequest{}
	err := decoder.Decode(&chirpRequestInstance)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	// Parse UUID
	uID, err := uuid.Parse(chirpRequestInstance.UserID)
	if err != nil {
		log.Printf("Error parsing userID: %v", err)
		w.WriteHeader(400)
		return
	}

	// Validate length
	if len(chirpRequestInstance.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "Chirp is too long"})
		return
	}

	// Clean bad words BEFORE saving to database
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	cleanedBody := cleanBadWords(chirpRequestInstance.Body, badWords)

	// Create chirp in database
	chirpParams := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: uID,
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		log.Printf("Cannot create chirp: %v", err)
		w.WriteHeader(500)
		return
	}

	// Return proper chirp response
	response := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
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

/*
	GET /api/chirps

Returns all chirps in the database in the same structure as
the POST /api/chirps endpoint but as an array. Ordered by create_at in ascending order.
200 status code for success
*/
func (cfg *apiConfig) GetAllChirpsHandler(w http.ResponseWriter, r *http.Request) {

	allChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Cannot get all chirps: %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")

	var arr []Chirp

	for _, curr := range allChirps {
		chirp := Chirp{
			ID:        curr.ID,
			CreatedAt: curr.CreatedAt,
			UpdatedAt: curr.UpdatedAt,
			Body:      curr.Body,
			UserID:    curr.UserID,
		}
		arr = append(arr, chirp)
	}
	json.NewEncoder(w).Encode(arr)

}

// func (q *Queries) GetAChirp(ctx context.Context, id uuid.UUID) (Chirp, error) {
func (cfg *apiConfig) GetAChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	uuID, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("Error parsing UUID from chirpID: %v", err)
		w.WriteHeader(404)
		return
	}

	chirp, err := cfg.db.GetAChirp(r.Context(), uuID)
	log.Printf("chirp: %v", chirp)
	if err != nil {
		log.Printf("Error getting the chirp: %v", err)
		w.WriteHeader(404)
		return
	}

	response := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	json.NewEncoder(w).Encode(response)
}
