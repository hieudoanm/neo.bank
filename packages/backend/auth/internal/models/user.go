package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	AccountID    uuid.UUID
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	TokenHash string
	UserID    uuid.UUID
	AccountID uuid.UUID
	Email     string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
