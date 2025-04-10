package main

import (
	"context"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handleReset(response http.ResponseWriter, _ *http.Request) {
	if cfg.platform != Dev {
		response.WriteHeader(http.StatusForbidden)
	}

	// Reset metrics
	cfg.fileserverHits.Store(0)

	// Reset database
	err := cfg.db.DeleteUsers(context.Background())
	if err != nil {
		respondWithError(response,
			http.StatusInternalServerError,
			fmt.Sprintf("Problem deleting users: %s", err))
	} else {
		response.WriteHeader(http.StatusOK)
	}
}