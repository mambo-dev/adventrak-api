package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/auth"
	"github.com/mambo-dev/adventrak-backend/internal/database"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,gte=8"`
		Username string `json:"username" validate:"required,min=5,max=20"`
	}

	err := rateLimit(w, r, "signup")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err = decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode parameters", err, false)
		return
	}

	validate := validator.New()

	err = validate.Struct(params)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to validate user input", err, true)
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create password hash", err, false)
		return
	}

	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Username:     params.Username,
		PasswordHash: passwordHash,
		Email:        params.Email,
	})

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to create user", err, false)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to make refresh token", err, false)
		return
	}

	token, err := cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 730),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err, false)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err, false)
		return
	}

	respondWithJSON(w, http.StatusCreated, UserAuthResponse{
		Username:     user.Username,
		CreatedAt:    user.CreatedAt,
		AccessToken:  accessToken,
		ID:           user.ID,
		RefreshToken: token.Token,
	})

}
