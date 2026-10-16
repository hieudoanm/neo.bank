package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/errors"
)

func TestFetchBalanceCrossAccountReturns404(t *testing.T) {
	svc := NewAccountService(nil)
	acc := uuid.New()
	_, err := svc.FetchBalance(context.Background(), acc, "different-account-id")
	if err != errors.ErrAccountNotFound {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}
