package models

import (
	"time"

	"github.com/google/uuid"
)

type Transfer struct {
	ID              uuid.UUID
	IdempotencyKey  uuid.UUID
	SourceAccountID uuid.UUID
	DestAccountID   uuid.UUID
	AmountCents     int64
	CurrencyCode    string
	Status          string
	FailureReason   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateTransferRequest struct {
	DestinationAccountId string `json:"destinationAccountId"`
	AmountCents          int64  `json:"amountCents"`
}

type CreateTransferResponse struct {
	TransferId    string `json:"transferId"`
	Status        string `json:"status"`
	SourceAccount string `json:"sourceAccount"`
	DestAccount   string `json:"destinationAccount"`
	AmountCents   int64  `json:"amountCents"`
	Currency      string `json:"currency"`
}
