package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/venzy/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	maxChirpLength int
	db *database.Queries
}

func (cfg *apiConfig) withMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Problem opening database: %v\n", err)
	}

	dbQueries := database.New(db)
	cfg := &apiConfig{
		maxChirpLength: 140,
		db: dbQueries,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.withMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handleReady)
	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handleMetricsReset)
	mux.HandleFunc("POST /api/validate_chirp", cfg.handleValidate)

	server := &http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}
