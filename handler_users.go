package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hugermuger/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) addUser(w http.ResponseWriter, r *http.Request) {
	type setUser struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	mail := setUser{}
	err := decoder.Decode(&mail)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), mail.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't set user in database", err)
		return
	}

	jsonUser := jsonUser(user)

	respondWithJSON(w, http.StatusCreated, jsonUser)
}

func jsonUser(user database.User) User {
	return User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
}
