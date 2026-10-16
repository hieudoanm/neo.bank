package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/models"
)

type fakeAccountStore struct {
	account *models.Account
	err     error
}

func (f *fakeAccountStore) FindByID(_ context.Context, _ uuid.UUID) (*models.Account, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.account, nil
}

func TestFetchBalanceCrossAccountReturns404(t *testing.T) {
	svc := NewAccountService(nil)
	acc := uuid.New()
	_, err := svc.FetchBalance(context.Background(), acc, "different-account-id")
	if err != apperrors.ErrAccountNotFound {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestFetchBalanceReturnsAccount(t *testing.T) {
	acc := uuid.New()
	updated := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	store := &fakeAccountStore{account: &models.Account{
		ID:           acc,
		BalanceCents: 12345,
		CurrencyCode: "USD",
		UpdatedAt:    updated,
	}}
	svc := NewAccountService(store)

	resp, err := svc.FetchBalance(context.Background(), acc, acc.String())
	if err != nil {
		t.Fatalf("FetchBalance: %v", err)
	}
	if resp.AccountID != acc.String() || resp.BalanceCents != 12345 || resp.CurrencyCode != "USD" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.UpdatedAt != "2026-03-15T10:00:00Z" {
		t.Fatalf("unexpected updatedAt: %s", resp.UpdatedAt)
	}
}

func TestFetchBalancePropagatesAppError(t *testing.T) {
	acc := uuid.New()
	svc := NewAccountService(&fakeAccountStore{err: apperrors.ErrAccountNotFound})
	if _, err := svc.FetchBalance(context.Background(), acc, acc.String()); err != apperrors.ErrAccountNotFound {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestFetchBalanceWrapsUnknownError(t *testing.T) {
	acc := uuid.New()
	svc := NewAccountService(&fakeAccountStore{err: errors.New("boom")})
	if _, err := svc.FetchBalance(context.Background(), acc, acc.String()); err != apperrors.ErrInternal {
		t.Fatalf("expected ErrInternal, got %v", err)
	}
}
