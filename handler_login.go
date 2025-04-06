package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mambo-dev/adventrak-backend/internal/auth"
)

func (cfg apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)

	params := &Params{}

	err := decoder.Decode(params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode sent parameters", err, false)
		return
	}

	user, err := cfg.db.GetUser(context.Background(), params.Username)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid credentials", err, false)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.PasswordHash)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid Credentials", err, false)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err, false)
		return
	}

	respondWithJSON(w, http.StatusAccepted, UserAuthResponse{
		ID:           user.ID,
		Username:     user.Username,
		AccessToken:  accessToken,
		RefreshToken: accessToken,
	})

}
