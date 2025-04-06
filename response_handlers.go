package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type MultipleErrorResponse struct {
	Error map[string]string `json:"errors"`
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error, multiple bool) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", err)
	}

	log.Print(err)

	type errorResponse struct {
		Error string `json:"error"`
	}

	if multiple {
		errors := generateValidationError(err)
		respondWithJSON(w, code, MultipleErrorResponse{
			Error: errors,
		})
		return
	}

	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err, false)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func generateValidationError(err error) map[string]string {
	errors := make(map[string]string, 0)

	validationErrors := err.(validator.ValidationErrors)
	for _, validationError := range validationErrors {
		errors[validationError.Field()] = fmt.Sprintf("Validation failed on the %v tag", validationError.Tag())
	}

	return errors
}
