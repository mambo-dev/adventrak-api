package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mambo-dev/adventrak-backend/internal/auth"
	"github.com/mambo-dev/adventrak-backend/internal/database"
	"github.com/mambo-dev/adventrak-backend/internal/mailer"
)

func (cfg *apiConfig) handlerSignup(w http.ResponseWriter, r *http.Request) {
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

	user, err := cfg.db.GetUser(context.Background(), database.GetUserParams{
		Username: params.Username,
	})

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password", err, false)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.PasswordHash)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password", err, false)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Minute*10)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err, false)
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
		respondWithError(w, http.StatusInternalServerError, "Failed to save refresh token", err, false)
		return
	}

	respondWithJSON(w, http.StatusAccepted, UserAuthResponse{
		ID:           user.ID,
		Username:     user.Username,
		AccessToken:  accessToken,
		RefreshToken: token.Token,
		CreatedAt:    user.CreatedAt,
	})

}

func (cfg apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		RefreshToken string `json:"refreshToken"`
	}

	decoder := json.NewDecoder(r.Body)

	params := &Params{}

	err := decoder.Decode(params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode sent parameters", err, false)
		return
	}

	if params.RefreshToken == "" || len(params.RefreshToken) <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid Refresh Token", err, false)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(context.Background(), params.RefreshToken)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid Refresh Token", err, false)
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Refresh Token was revoked at:%v", refreshToken.RevokedAt.Time), err, false)
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh Token is expired login again", err, false)
		return
	}

	if !refreshToken.UserID.Valid {
		respondWithError(w, http.StatusInternalServerError, "Failed to register user to token during login", err, false)
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID.UUID, cfg.jwtSecret, time.Minute*10)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create access token", err, false)
		return
	}

	newRefreshToken, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to make refresh token", err, false)
		return
	}

	token, err := cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 730),
		UserID:    refreshToken.UserID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token", err, false)
		return
	}

	type RefreshResponse struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}

	respondWithJSON(w, http.StatusAccepted, RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: token.Token,
	})

}

func (cfg apiConfig) handlerVerifyEmail(w http.ResponseWriter, r *http.Request) {

	err := rateLimit(w, r, "general")

	if err != nil {
		respondWithError(w, http.StatusForbidden, "Too many requests. Please slow down.", err, false)
		return
	}
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to get bearer token", err, false)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to validate JWT", err, false)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), database.GetUserParams{
		ID: userID,
	})

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find user possibly deleted", err, false)
		return
	}

	err = mailer.SendEmail(mailer.EmailDetails{
		FromEmail:   mailer.SystemEmails["system"].Email,
		FromName:    mailer.SystemEmails["system"].Name,
		ToEmail:     user.Email,
		ToName:      user.Username,
		Subject:     "Verify your Email",
		HtmlContent: "<strong>Your email is not verified</strong>",
	}, cfg.sendGridApiKey)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to send verification email", err, false)
		return
	}

	respondWithJSON(w, http.StatusOK, nil)

}
