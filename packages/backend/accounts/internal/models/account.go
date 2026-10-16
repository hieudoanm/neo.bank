package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	BalanceCents int64
	CurrencyCode string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type BalanceResponse struct {
	AccountID    string `json:"accountId"`
	BalanceCents int64  `json:"balanceCents"`
	CurrencyCode string `json:"currency"`
}
