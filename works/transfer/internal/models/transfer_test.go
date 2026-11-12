package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTransferResponse(t *testing.T) {
	id := uuid.New()
	src := uuid.New()
	dst := uuid.New()
	created := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	updated := created.Add(time.Minute)
	tr := &Transfer{
		ID:              id,
		SourceAccountID: src,
		DestAccountID:   dst,
		AmountCents:     500,
		Status:          StatusPending,
		IdempotencyKey:  "key",
		CorrelationID:   uuid.NewString(),
		CreatedAt:       created,
		UpdatedAt:       updated,
		FailureReason:   "",
	}
	resp := NewTransferResponse(tr)
	if resp.TransferId != id.String() || resp.SourceAccountId != src.String() || resp.DestinationAccountId != dst.String() {
		t.Fatalf("unexpected ids: %+v", resp)
	}
	if resp.Status != StatusPending || resp.AmountCents != 500 {
		t.Fatalf("unexpected fields: %+v", resp)
	}
	if resp.CreatedAt != "2026-03-15T10:00:00Z" || resp.UpdatedAt != "2026-03-15T10:01:00Z" {
		t.Fatalf("unexpected timestamps: %+v", resp)
	}
}
