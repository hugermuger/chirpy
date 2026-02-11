package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hugermuger/chirpy/internal/auth"
	"github.com/hugermuger/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) addChirp(w http.ResponseWriter, r *http.Request) {
	type setChirp struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get Bearer Token", err)
		return
	}

	testID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate Token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	chirpIn := setChirp{}
	err = decoder.Decode(&chirpIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(chirpIn.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirpIn.Body = profaneFilter(chirpIn.Body)

	params := database.CreateChirpParams{
		Body:   chirpIn.Body,
		UserID: testID,
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't set chirp in database", err)
		return
	}

	jsonChirp := jsonChirp(chirp)

	respondWithJSON(w, http.StatusCreated, jsonChirp)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	sort := r.URL.Query().Get("sort")
	chirps := []database.Chirp{}

	if s != "" {
		chirp, err := cfg.dbQueries.GetChirpsByUserID(r.Context(), uuid.MustParse(s))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
			return
		}
		chirps = chirp
	} else {
		chirp, err := cfg.dbQueries.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
			return
		}
		chirps = chirp
	}

	jsonChirps := []Chirp{}

	if sort == "desc" {
		slices.SortFunc(chirps, func(a, b database.Chirp) int { return b.CreatedAt.Compare(a.CreatedAt) })
	}

	for _, chirp := range chirps {
		jsonChirps = append(jsonChirps, jsonChirp(chirp))
	}

	respondWithJSON(w, http.StatusOK, jsonChirps)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), uuid.MustParse(id))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	jsonChirp := jsonChirp(chirp)

	respondWithJSON(w, http.StatusOK, jsonChirp)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get Bearer Token", err)
		return
	}

	testID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate Token", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), uuid.MustParse(id))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find chirp", err)
		return
	}

	if chirp.UserID == testID {
		err = cfg.dbQueries.DeleteChirp(r.Context(), uuid.MustParse(id))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	respondWithError(w, http.StatusForbidden, "Not the owner of the chirp", err)
}

func jsonChirp(chirp database.Chirp) Chirp {
	return Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
}

func profaneFilter(s string) string {
	words := strings.Split(s, " ")
	expWords := []string{}
	for _, word := range words {
		if strings.ToLower(word) == "kerfuffle" ||
			strings.ToLower(word) == "sharbert" ||
			strings.ToLower(word) == "fornax" {
			word = "****"
		}
		expWords = append(expWords, word)
	}

	return strings.Join(expWords, " ")
}
