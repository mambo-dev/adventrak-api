package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/auth"
)

func (cfg apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Password string `json:"password" validate:"required,gte=8"`
		Username string `json:"username" validate:"required,min=5,max=20"`
	}

	err := rateLimit(w, r, "login")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	decoder := json.NewDecoder(r.Body)

	params := &Params{}

	err = decoder.Decode(params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode sent parameters", err, false)
		return
	}

	validate := validator.New()

	err = validate.Struct(params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to validate user input", err, true)
		return
	}

	user, err := cfg.db.GetUser(context.Background(), params.Username)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password", err, false)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.PasswordHash)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password", err, false)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err, false)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(context.Background(), uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err, false)
		return
	}

	respondWithJSON(w, http.StatusAccepted, UserAuthResponse{
		ID:           user.ID,
		Username:     user.Username,
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		CreatedAt:    user.CreatedAt,
	})

}
