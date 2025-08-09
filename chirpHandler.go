package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"email"`
	UserID    string    `json:"user_id"`
}

// check length of body and censor bad words
func (cfg *apiConfig) CreateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type request struct { // request body
		Body   string `json:"body"`
		UserID string `json:"user_id"`
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
