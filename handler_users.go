package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mambo-dev/adventrak-backend/internal/auth"
	"github.com/mambo-dev/adventrak-backend/internal/database"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,gte=8"`
		Username string `json:"username" validate:"required,min=5,max=20"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err := decoder.Decode(&params)

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
		Username: params.Username,
		Password: passwordHash,
		Email:    params.Email,
	})

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to create user", err, false)
		return
	}

	type ReturnUser struct {
		Email     string
		Username  string
		CreatedAt time.Time
	}

	respondWithJSON(w, http.StatusCreated, ReturnUser{
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})

}
