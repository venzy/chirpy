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

func (cfg *apiConfig) handleCreateChirp(response http.ResponseWriter, request *http.Request, userID uuid.UUID) {
	type requestParams struct {
		Body string `json:"body"`
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
	_, err = cfg.db.GetUserByID(request.Context(), userID)
	if err != nil {
		msg := fmt.Sprintf("chirps: Could not get user with ID '%s': %s", userID, err)
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
		UserID: userID,
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

func (cfg *apiConfig) handleGetChirps(response http.ResponseWriter, request *http.Request) {
	chirpRows, err := cfg.db.GetChirps(request.Context())
	if err != nil {
		msg := fmt.Sprintf("chirps: Problem retrieving all chirps: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusInternalServerError, msg)
		return
	}
	chirps := []Chirp{}
	for _, chirp := range chirpRows {
		chirps = append(chirps, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}

	respondWithJSON(response, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleGetChirpByID(response http.ResponseWriter, request *http.Request) {
	// Parse request params
	chirpID, err := uuid.Parse(request.PathValue("chirpID"))
	if err != nil {
		msg := fmt.Sprintf("chirps: Problem parsing chirpID from request: %s", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	// DB fetch - exercise calls for indiscriminate 404 response on any problem
	row, err := cfg.db.GetChirpByID(request.Context(), chirpID)
	if err != nil {
		msg := fmt.Sprintf("chirps: Problem retrieving chirp with id '%s': %s", chirpID, err)
		log.Println(msg)
		respondWithError(response, http.StatusNotFound, msg)
		return
	}

	// Response
	respondWithJSON(response, http.StatusOK, Chirp{
		ID: row.ID,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Body: row.Body,
		UserID: row.UserID,
	})
}