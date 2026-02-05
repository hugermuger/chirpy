package main

import (
	"encoding/json"
	"net/http"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	message := chirp{}
	err := decoder.Decode(&message)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if len(message.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Valid: true,
	})
}
