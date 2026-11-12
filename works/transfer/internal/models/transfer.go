package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusReversed   = "reversed"
)

type Transfer struct {
	ID              uuid.UUID
	SourceAccountID uuid.UUID
	DestAccountID   uuid.UUID
	AmountCents     int64
	Status          string
	IdempotencyKey  string
	CorrelationID   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FailureReason   string
}

type CreateTransferRequest struct {
	DestinationAccountId string `json:"destinationAccountId"`
	AmountCents          int64  `json:"amountCents"`
}

type TransferResponse struct {
	TransferId           string `json:"transferId"`
	Status               string `json:"status"`
	CreatedAt            string `json:"createdAt"`
	AmountCents          int64  `json:"amountCents"`
	SourceAccountId      string `json:"sourceAccountId"`
	DestinationAccountId string `json:"destinationAccountId"`
	UpdatedAt            string `json:"updatedAt"`
	FailureReason        string `json:"failureReason,omitempty"`
}

func NewTransferResponse(t *Transfer) *TransferResponse {
	return &TransferResponse{
		TransferId:           t.ID.String(),
		Status:               t.Status,
		CreatedAt:            t.CreatedAt.UTC().Format(time.RFC3339),
		AmountCents:          t.AmountCents,
		SourceAccountId:      t.SourceAccountID.String(),
		DestinationAccountId: t.DestAccountID.String(),
		UpdatedAt:            t.UpdatedAt.UTC().Format(time.RFC3339),
		FailureReason:        t.FailureReason,
	}
}
