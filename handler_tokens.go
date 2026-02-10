package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hugermuger/chirpy/internal/auth"
	"github.com/hugermuger/chirpy/internal/database"
)

type Token struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > 0 {
		respondWithError(w, http.StatusUnsupportedMediaType, "Request body not allowed", fmt.Errorf("Request body not allowed"))
		return
	}

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get Bearer Token", err)
		return
	}

	dbtoken, err := cfg.dbQueries.GetToken(r.Context(), refresh_token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	} else if dbtoken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token revoked", err)
		return
	} else if dbtoken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Token expired", err)
		return
	}

	token, err := auth.MakeJWT(dbtoken.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make JWT", err)
		return
	}

	jsonToken := Token{
		Token: token,
	}

	respondWithJSON(w, http.StatusOK, jsonToken)
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > 0 {
		respondWithError(w, http.StatusUnsupportedMediaType, "Request body not allowed", fmt.Errorf("Request body not allowed"))
		return
	}

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get Bearer Token", err)
		return
	}

	params := database.MarkTokenRevokedParams{
		UpdatedAt: time.Now(),
		ID:        refresh_token,
	}

	err = cfg.dbQueries.MarkTokenRevoked(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find Token in DB", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
