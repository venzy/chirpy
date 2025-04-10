package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(response http.ResponseWriter, request *http.Request) {
	type requestParams struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding createUser params: %s\n", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Basic validation
	if _, err := mail.ParseAddress(params.Email); err != nil {
		respondWithError(response,
			http.StatusBadRequest,
			fmt.Sprintf("Bad email address: %s", params.Email))
	} else {
		// Create in DB, then return representation of new row as response
		newRow, err := cfg.db.CreateUser(request.Context(), params.Email)
		if err != nil {
			respondWithError(response,
				http.StatusInternalServerError,
				fmt.Sprintf("Problem creating user: %s", err))
		}
		newUser := User{
			ID: newRow.ID,
			CreatedAt: newRow.CreatedAt,
			UpdatedAt: newRow.UpdatedAt,
			Email: newRow.Email,
		}
		respondWithJSON(response, http.StatusCreated, newUser)
	}
}

