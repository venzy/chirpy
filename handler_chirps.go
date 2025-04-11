package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/venzy/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleCreateChirp(response http.ResponseWriter, request *http.Request) {
	type requestParams struct {
		Body string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("chirps: Error decoding createChirp params: %s\n", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Confirm user exists
	parsedUserID, err := uuid.Parse(params.UserID)
	if err != nil {
		msg := fmt.Sprintf("chirps: Error parsing supplied UserID: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}
	_, err = cfg.db.GetUserByID(request.Context(), parsedUserID)
	if err != nil {
		msg := fmt.Sprintf("chirps: Could not get user with ID '%s': %s", params.UserID, err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Validate intended message
	if len(params.Body) > cfg.maxChirpLength {
		msg := fmt.Sprintf("chirps: chirp too long, must be less than or equal to %d chars", cfg.maxChirpLength)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// Clean up message
	cleanedBody := cleanBody(params.Body)

	// Create in DB
	newChirpRow, err := cfg.db.CreateChirp(request.Context(), database.CreateChirpParams{
		Body: cleanedBody,
		UserID: parsedUserID,
	})
	if err != nil {
		msg := fmt.Sprintf("chirps: Problem creating chirp: %s", err)
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}

	// Respond with success
	newChirp := Chirp{
		ID: newChirpRow.ID,
		CreatedAt: newChirpRow.CreatedAt,
		UpdatedAt: newChirpRow.UpdatedAt,
		Body: newChirpRow.Body,
		UserID: newChirpRow.UserID,
	}
	respondWithJSON(response, http.StatusCreated, newChirp)
}

func cleanBody(body string) string {
	splitBody := strings.Split(body, " ")
	newSplitBody := []string{}
	for _, word := range splitBody {
		lower := strings.ToLower(word)
		if lower == "kerfuffle" || lower == "sharbert" || lower == "fornax" {
			newSplitBody = append(newSplitBody, "****")
		} else {
			newSplitBody = append(newSplitBody, word)
		}
	}
	return strings.Join(newSplitBody, " ")
}