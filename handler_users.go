package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hugermuger/chirpy/internal/auth"
	"github.com/hugermuger/chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type setUser struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (cfg *apiConfig) addUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	inputuser := setUser{}
	err := decoder.Decode(&inputuser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	password, err := auth.HashPassword(inputuser.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	params := database.CreateUserParams{
		Email:          inputuser.Email,
		HashedPassword: password,
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't set user in database", err)
		return
	}

	jsonUser := jsonUser(user)

	respondWithJSON(w, http.StatusCreated, jsonUser)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	inputuser := setUser{}
	err := decoder.Decode(&inputuser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.dbQueries.GetUser(r.Context(), inputuser.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	correct, err := auth.CheckPasswordHash(inputuser.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't check password", err)
		return
	}

	if !correct {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	jsonUser := jsonUser(user)

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make JWT", err)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()

	params := database.CreateTokenParams{
		ID:        refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}

	cfg.dbQueries.CreateToken(r.Context(), params)

	jsonUser.RefreshToken = refreshToken
	jsonUser.Token = token
	respondWithJSON(w, http.StatusOK, jsonUser)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing Token", err)
		return
	}

	UserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate Token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	inputuser := setUser{}
	err = decoder.Decode(&inputuser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	password, err := auth.HashPassword(inputuser.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	params := database.UpdateUserParams{
		ID:             UserID,
		Email:          inputuser.Email,
		HashedPassword: password,
	}

	user, err := cfg.dbQueries.UpdateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	jsonUser := jsonUser(user)

	respondWithJSON(w, http.StatusOK, jsonUser)
}

func (cfg *apiConfig) setUserRed(w http.ResponseWriter, r *http.Request) {
	type Polka struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	polka := Polka{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&polka)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if polka.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No ApiKey", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Wrong ApiKey", err)
		return
	}

	params := database.SetUserRedParams{
		IsChirpyRed: true,
		ID:          uuid.MustParse(polka.Data.UserID),
	}

	err = cfg.dbQueries.SetUserRed(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func jsonUser(user database.User) User {
	return User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		Token:       "",
		IsChirpyRed: user.IsChirpyRed,
	}
}
