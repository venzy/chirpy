package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handleValidate(response http.ResponseWriter, request *http.Request) {
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

func respondWithError(response http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	responseBody, err := json.Marshal(errorResponse{
		Error: msg,
	})
	if err != nil {
		log.Printf("Error encoding error response: %s\n", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(code)
	response.Write(responseBody)
}

func respondWithJSON(response http.ResponseWriter, code int, payload interface{}) {
	responseBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding success response: %s\n", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(code)
	response.Write(responseBody)
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