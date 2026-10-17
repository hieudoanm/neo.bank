package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTransactionResponse(t *testing.T) {
	id := uuid.New()
	cp := uuid.New()
	tid := uuid.New()
	created := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	tx := &Transaction{
		AccountID:             uuid.New(),
		CreatedAt:             created,
		TransactionID:         id,
		TransferID:            &tid,
		Type:                  TypeDebit,
		AmountCents:           250,
		CounterpartyAccountID: cp,
		Status:                StatusCompleted,
	}
	resp := NewTransactionResponse(tx)
	if resp.TransactionID != id.String() || resp.TransferID == nil || *resp.TransferID != tid.String() {
		t.Fatalf("unexpected ids: %+v", resp)
	}
	if resp.Type != TypeDebit || resp.AmountCents != 250 || resp.CounterpartyAccountID != cp.String() {
		t.Fatalf("unexpected fields: %+v", resp)
	}
	if resp.Status == nil || *resp.Status != StatusCompleted {
		t.Fatalf("unexpected status: %+v", resp.Status)
	}
	if resp.CreatedAt != "2026-03-15T10:00:00Z" {
		t.Fatalf("unexpected createdAt: %s", resp.CreatedAt)
	}
}

func TestNewTransactionResponseNilOptionals(t *testing.T) {
	tx := &Transaction{TransactionID: uuid.New(), Type: TypeCredit, AmountCents: 1}
	resp := NewTransactionResponse(tx)
	if resp.TransferID != nil {
		t.Fatal("expected nil transferId when unset")
	}
	if resp.Status != nil {
		t.Fatal("expected nil status when empty")
	}
}
