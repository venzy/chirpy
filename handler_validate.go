package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (cfg *apiConfig) handleValidate(response http.ResponseWriter, request *http.Request) {
	type requestParams struct {
		Body string `json:"body"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	type successResponse struct {
		Valid bool `json:"valid"`
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
		responseBody, err := json.Marshal(errorResponse{
			Error: fmt.Sprintf("chirp too long, must be less than or equal to %d chars", cfg.maxChirpLength),
		})
		if err != nil {
			log.Printf("Error encoding error response: %s\n", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		response.WriteHeader(http.StatusBadRequest)
		response.Write(responseBody)
	} else {
		responseBody, err := json.Marshal(successResponse{Valid: true})
		if err != nil {
			log.Printf("Error encoding success response: %s\n", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		response.WriteHeader(http.StatusOK)
		response.Write(responseBody)
	}
}