package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	maxChirpLength int
}

func (cfg *apiConfig) withMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := &apiConfig{maxChirpLength: 140}
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.withMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handleReady)
	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handleMetricsReset)
	mux.HandleFunc("POST /api/validate_chirp", cfg.handleValidate)

	server := &http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}
