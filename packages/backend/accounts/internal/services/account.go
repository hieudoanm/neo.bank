package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/models"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/repositories"
)

type AccountService struct {
	repo *repositories.AccountRepository
}

func NewAccountService(repo *repositories.AccountRepository) *AccountService {
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
	}, nil
}
