package main

import (
	"time"

	"github.com/google/uuid"
)

type UserAuthResponse struct {
	ID           uuid.UUID
	Username     string
	AccessToken  string
	RefreshToken string
	CreatedAt    time.Time
}
