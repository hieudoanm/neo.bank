package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/models"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/repositories"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// TransactionReader is the read model query the service needs.
// *repositories.TransactionRepository satisfies it in production.
type TransactionReader interface {
	ListByAccount(ctx context.Context, accountID uuid.UUID, params repositories.ListParams) ([]models.Transaction, error)
}

type TransactionService struct {
	repo TransactionReader
}

func NewTransactionService(repo TransactionReader) *TransactionService {
	return &TransactionService{repo: repo}
}

// ListTransactions returns a page of transaction history for an account. The
// caller may only read their own account's history; any other account is
// reported as not found to avoid disclosing its existence.
func (s *TransactionService) ListTransactions(ctx context.Context, accountID uuid.UUID, callerAccountID, cursor string, limit int) (*models.TransactionListResponse, error) {
	if callerAccountID != accountID.String() {
		return nil, errors.ErrAccountNotFound
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	params := repositories.ListParams{Limit: limit + 1}
	if cursor != "" {
		createdAt, transactionID, err := decodeCursor(cursor)
		if err != nil {
			return nil, err
		}
		params.AfterCreatedAt = &createdAt
		params.AfterTransactionID = &transactionID
	}

	rows, err := s.repo.ListByAccount(ctx, accountID, params)
	if err != nil {
		return nil, errors.ErrInternal
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	resp := &models.TransactionListResponse{
		Transactions: make([]models.TransactionResponse, 0, len(rows)),
		HasMore:      hasMore,
	}
	for i := range rows {
		resp.Transactions = append(resp.Transactions, models.NewTransactionResponse(&rows[i]))
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		next := encodeCursor(last.CreatedAt, last.TransactionID)
		resp.NextCursor = &next
	}
	return resp, nil
}
