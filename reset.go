package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) metricsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write(fmt.Appendf([]byte{}, "Hits reset to: %v", cfg.fileserverHits.Load()))
}
