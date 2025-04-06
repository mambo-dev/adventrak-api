package main

import (
	"time"

	"github.com/google/uuid"
)

type UserAuthResponse struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	CreatedAt    time.Time `json:"createdAt"`
}
