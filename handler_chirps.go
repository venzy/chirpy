package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handleCreateChirp(response http.ResponseWriter, request *http.Request) {
	type requestParams struct {
		Body string `json:"body"`
	}

	type successResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding validate params: %s\n", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(params.Body) > cfg.maxChirpLength {
		respondWithError(response,
			http.StatusBadRequest,
			fmt.Sprintf("chirp too long, must be less than or equal to %d chars", cfg.maxChirpLength))
	} else {
		respondWithJSON(response,
			http.StatusOK,
			successResponse{CleanedBody: cleanBody(params.Body)})
	}
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