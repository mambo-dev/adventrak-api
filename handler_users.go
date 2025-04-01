package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
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

}
