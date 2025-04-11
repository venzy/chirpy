package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/venzy/chirpy/internal/auth"
	"github.com/venzy/chirpy/internal/database"
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
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("users: Error decoding createUser params: %s\n", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Basic validation
	if _, err := mail.ParseAddress(params.Email); err != nil {
		msg := fmt.Sprintf("users: Bad email address: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	if len(params.Password) < 1 {
		msg := fmt.Sprintf("users: Password len > 1 required")
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Create in DB, then return representation of new row as response
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		msg := fmt.Sprintf("users: Problem hashing supplied password")
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}

	newRow, err := cfg.db.CreateUser(request.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		msg := fmt.Sprintf("users: Problem creating user: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}

	newUser := User{
		ID: newRow.ID,
		CreatedAt: newRow.CreatedAt,
		UpdatedAt: newRow.UpdatedAt,
		Email: newRow.Email,
	}
	respondWithJSON(response, http.StatusCreated, newUser)
}

func (cfg *apiConfig) handleLogin(response http.ResponseWriter, request *http.Request) {
	type requestParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("users: Error decoding login params: %s\n", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Basic validation
	if _, err := mail.ParseAddress(params.Email); err != nil {
		msg := fmt.Sprintf("users: Bad email address: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Get user
	user, err := cfg.db.GetUserByEmail(request.Context(), params.Email)
	if err != nil {
		msg := fmt.Sprintf("users: No such user '%s' or error: %s", params.Email, err)
		log.Println(msg)
		// Don't leak which was wrong (email or password)
		respondWithError(response, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// Auth
	if err = auth.CheckPasswordHash(user.HashedPassword, params.Password); err != nil {
		msg := fmt.Sprintf("users: Password hash mismatch or error: %s", err)
		log.Println(msg)
		// Don't leak which was wrong (email or password)
		respondWithError(response, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// Later, replace/supplement with auth token
	loggedInUser := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	respondWithJSON(response, http.StatusOK, loggedInUser)
}
	