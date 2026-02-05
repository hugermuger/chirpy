package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	cfg := apiConfig{}

	mux := http.NewServeMux()

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", readinessEndpoint)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsRead)
	mux.HandleFunc("POST /admin/reset", cfg.metricsReset)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving from %v on port: %v\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
