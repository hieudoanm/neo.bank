package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeDebit  = "debit"
	TypeCredit = "credit"

	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusReversed  = "reversed"
)

// Transaction is a row in the transaction_keyspace transactions table: an
// append-only record of one side of a balance movement, projected from the saga.
type Transaction struct {
	AccountID             uuid.UUID
	CreatedAt             time.Time
	TransactionID         uuid.UUID
	TransferID            *uuid.UUID
	Type                  string
	AmountCents           int64
	CounterpartyAccountID uuid.UUID
	Status                string
}

type TransactionResponse struct {
	TransactionID         string  `json:"transactionId"`
	TransferID            *string `json:"transferId"`
	Type                  string  `json:"type"`
	AmountCents           int64   `json:"amountCents"`
	CounterpartyAccountID string  `json:"counterpartyAccountId"`
	Status                *string `json:"status"`
	CreatedAt             string  `json:"createdAt"`
}

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	NextCursor   *string               `json:"nextCursor"`
	HasMore      bool                  `json:"hasMore"`
}

func NewTransactionResponse(t *Transaction) TransactionResponse {
	resp := TransactionResponse{
		TransactionID:         t.TransactionID.String(),
		Type:                  t.Type,
		AmountCents:           t.AmountCents,
		CounterpartyAccountID: t.CounterpartyAccountID.String(),
		CreatedAt:             t.CreatedAt.UTC().Format(time.RFC3339),
	}
	if t.TransferID != nil {
		v := t.TransferID.String()
		resp.TransferID = &v
	}
	if t.Status != "" {
		v := t.Status
		resp.Status = &v
	}
	return resp
}
