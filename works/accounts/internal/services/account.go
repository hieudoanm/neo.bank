package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/models"
)

// AccountStore is the persistence dependency for account records.
type AccountStore interface {
	FindByID(ctx context.Context, accountID uuid.UUID) (*models.Account, error)
}

type AccountService struct {
	repo AccountStore
}

func NewAccountService(repo AccountStore) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) FetchBalance(ctx context.Context, accountID uuid.UUID, callerAccountID string) (*models.BalanceResponse, error) {
	if callerAccountID != accountID.String() {
		return nil, errors.ErrAccountNotFound
	}
	account, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		if _, ok := err.(*errors.AppError); ok {
			return nil, err
		}
		return nil, errors.ErrInternal
	}
	return &models.BalanceResponse{
		AccountID:    account.ID.String(),
		BalanceCents: account.BalanceCents,
		CurrencyCode: account.CurrencyCode,
		UpdatedAt:    account.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}
