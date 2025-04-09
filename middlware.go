package main

import (
	"context"
	"net/http"

	"github.com/mambo-dev/adventrak-backend/internal/auth"
)

type key string

const UserIDKey key = "userID"

func (cfg apiConfig) authMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)

		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Failed to get bearer token", err, false)
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.jwtSecret)

		if err != nil {
			respondWithError(w, http.StatusForbidden, "Invalid Token. You are already logged out", err, false)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDKey, userID)))

	})
}

func (cfg apiConfig) UseAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg.authMiddleware(handler).ServeHTTP(w, r)
	}
}
