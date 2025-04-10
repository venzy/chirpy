package main

import (
	"encoding/json"
	"log"
	"net/http"
)

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