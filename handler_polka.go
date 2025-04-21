package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// Handlers for Polka payment processing webhooks

// NOTE: Ideally negative responses should include a retry-after header. Not including for now.
func (cfg *apiConfig) handlePolkaWebhook(response http.ResponseWriter, request *http.Request) {
	// Parse request params
	var params struct {
		Event string `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		}
	}

	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("polka: Error decoding webhook params: %s\n", err)
		log.Println(msg)
		respondWithError(response, http.StatusBadRequest, msg)
		return
	}

	switch params.Event {
	case "user.upgraded":
		cfg.handleUserUpgrade(response, request, params.Data.UserID)
	default:
		msg := fmt.Sprintf("polka: Unknown event type: %s\n", params.Event)
		log.Println(msg)
		respondWithError(response, http.StatusNoContent, msg)
		return
	}
}