package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/models"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/repositories"
)

type fakeReader struct {
	rows   []models.Transaction
	params []repositories.ListParams
	err    error
}

func (f *fakeReader) ListByAccount(_ context.Context, _ uuid.UUID, p repositories.ListParams) ([]models.Transaction, error) {
	f.params = append(f.params, p)
	return f.rows, f.err
}

func tx(createdAt time.Time) models.Transaction {
	return models.Transaction{
		AccountID:             uuid.New(),
		CreatedAt:             createdAt,
		TransactionID:         uuid.New(),
		Type:                  models.TypeDebit,
		AmountCents:           100,
		CounterpartyAccountID: uuid.New(),
		Status:                models.StatusCompleted,
	}
}

func TestListTransactionsFirstPage(t *testing.T) {
	accountID := uuid.New()
	base := time.Now().UTC().Truncate(time.Second)
	reader := &fakeReader{rows: []models.Transaction{tx(base), tx(base.Add(-time.Second))}}
	svc := NewTransactionService(reader)

	resp, err := svc.ListTransactions(context.Background(), accountID, accountID.String(), "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasMore {
		t.Fatalf("expected hasMore=false")
	}
	if resp.NextCursor != nil {
		t.Fatalf("expected nil nextCursor")
	}
	if len(resp.Transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(resp.Transactions))
	}
	if reader.params[0].AfterCreatedAt != nil {
		t.Fatalf("first page must not set a cursor")
	}
	if reader.params[0].Limit != defaultLimit+1 {
		t.Fatalf("expected limit %d, got %d", defaultLimit+1, reader.params[0].Limit)
	}
}

func TestListTransactionsHasMoreAndCursor(t *testing.T) {
	accountID := uuid.New()
	base := time.Now().UTC().Truncate(time.Second)
	rows := []models.Transaction{tx(base), tx(base.Add(-time.Second)), tx(base.Add(-2 * time.Second))}
	reader := &fakeReader{rows: rows}
	svc := NewTransactionService(reader)

	resp, err := svc.ListTransactions(context.Background(), accountID, accountID.String(), "", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasMore {
		t.Fatalf("expected hasMore=true")
	}
	if len(resp.Transactions) != 2 {
		t.Fatalf("expected page of 2, got %d", len(resp.Transactions))
	}
	if resp.NextCursor == nil {
		t.Fatalf("expected a nextCursor")
	}

	// The cursor must decode back to the last returned row.
	last := rows[1]
	gotTime, gotID, err := decodeCursor(*resp.NextCursor)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if !gotTime.Equal(last.CreatedAt) || gotID != last.TransactionID {
		t.Fatalf("cursor mismatch: got (%v,%v) want (%v,%v)", gotTime, gotID, last.CreatedAt, last.TransactionID)
	}
}

func TestListTransactionsSecondPage(t *testing.T) {
	accountID := uuid.New()
	base := time.Now().UTC().Truncate(time.Second)
	cursor := encodeCursor(base, uuid.New())
	reader := &fakeReader{}
	svc := NewTransactionService(reader)

	if _, err := svc.ListTransactions(context.Background(), accountID, accountID.String(), cursor, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader.params[0].AfterCreatedAt == nil || reader.params[0].AfterTransactionID == nil {
		t.Fatalf("expected cursor parameters to be set")
	}
}

func TestListTransactionsRejectsInvalidCursor(t *testing.T) {
	accountID := uuid.New()
	svc := NewTransactionService(&fakeReader{})
	_, err := svc.ListTransactions(context.Background(), accountID, accountID.String(), "not-base64!!", 5)
	if !errors.Is(err, apperrors.ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestListTransactionsDeniesOtherAccount(t *testing.T) {
	accountID := uuid.New()
	svc := NewTransactionService(&fakeReader{})
	_, err := svc.ListTransactions(context.Background(), accountID, uuid.NewString(), "", 5)
	if !errors.Is(err, apperrors.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestListTransactionsRejectsInjectionInCursorBeforeTouchingRepo(t *testing.T) {
	accountID := uuid.New()
	reader := &fakeReader{}
	svc := NewTransactionService(reader)

	_, err := svc.ListTransactions(context.Background(), accountID, accountID.String(), "'; DROP TABLE transactions; --", 5)
	if !errors.Is(err, apperrors.ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
	if len(reader.params) != 0 {
		t.Fatalf("an injection payload must never reach the repository, got %d queries", len(reader.params))
	}
}

func TestListTransactionsClampsLimit(t *testing.T) {
	accountID := uuid.New()
	reader := &fakeReader{}
	svc := NewTransactionService(reader)
	if _, err := svc.ListTransactions(context.Background(), accountID, accountID.String(), "", 5000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reader.params[0].Limit != maxLimit+1 {
		t.Fatalf("expected clamped limit %d, got %d", maxLimit+1, reader.params[0].Limit)
	}
}
