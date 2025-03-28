package main

import (
	"fmt"
	"io"
	"net/http"
)

func (cfg *apiConfig) handleMetrics(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	// Body
	io.WriteString(response, fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handleMetricsReset(response http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Store(0)
	response.WriteHeader(http.StatusOK)
}