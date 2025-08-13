package main

import (
	"encoding/json"
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

type UserRequest struct {
	Email string `json:"email"`
}

/*
create user handler by creating a new decoder on the response body.
make an instance of the request struct and decode on the instance's address
create a new user by calling the database's CreateUser function
*/
func (cfg *apiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	requestInstance := UserRequest{}
	err := decoder.Decode(&requestInstance)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	createdUser, err := cfg.db.CreateUser(r.Context(), requestInstance.Email)
	if err != nil {
		log.Printf("Cannot create user: %s", err)
		w.WriteHeader(500)
		return
	}
	params := User{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	dat, err := json.Marshal(params)
	if err != nil {
		log.Printf("Cannot marshal user into JSON: %s", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

// Delete all users in the database
func (cfg *apiConfig) DeleteAllUsersHandler(w http.ResponseWriter, r *http.Request) {
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
