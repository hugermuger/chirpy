package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/hugermuger/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	cfg := apiConfig{}

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	cfg.dbQueries = database.New(db)
	cfg.platform = os.Getenv("PLATFORM")
	cfg.secret = os.Getenv("SECRET")

	mux := http.NewServeMux()

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", readinessEndpoint)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsRead)
	mux.HandleFunc("POST /admin/reset", cfg.metricsReset)
	mux.HandleFunc("POST /api/chirps", cfg.addChirp)
	mux.HandleFunc("GET /api/chirps", cfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirp)
	mux.HandleFunc("POST /api/users", cfg.addUser)
	mux.HandleFunc("POST /api/login", cfg.loginUser)
	mux.HandleFunc("POST /api/refresh", cfg.refreshToken)
	mux.HandleFunc("POST /api/revoke", cfg.revokeToken)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving from %v on port: %v\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
