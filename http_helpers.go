package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/venzy/chirpy/internal/auth"
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

func (cfg *apiConfig) withAuthenticatedUser(handlerWithUser func(http.ResponseWriter, *http.Request, uuid.UUID)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		handlerWithUser(w, r, userID)
	})
}