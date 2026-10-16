package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/models"
)

type AccountRepository struct {
	session *gocql.Session
}

func NewAccountRepository(session *gocql.Session) *AccountRepository {
	return &AccountRepository{session: session}
}

const (
	selectAccountByID    = `SELECT account_id, user_id, balance_cents, currency_code, status, created_at, updated_at FROM accounts WHERE account_id = ?`
	selectAccountsByUser = `SELECT account_id FROM accounts_by_user WHERE user_id = ?`
)

func (r *AccountRepository) FindByID(ctx context.Context, accountID uuid.UUID) (*models.Account, error) {
	var a models.Account
	var createdAt, updatedAt time.Time
	var status string
	err := r.session.Query(selectAccountByID, accountID).
		WithContext(ctx).Consistency(gocql.One).
		Scan(&a.ID, &a.UserID, &a.BalanceCents, &a.CurrencyCode, &status, &createdAt, &updatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("find account by id: %w", err)
	}
	a.Status = status
	a.CreatedAt = createdAt
	a.UpdatedAt = updatedAt
	return &a, nil
}

func (r *AccountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	iter := r.session.Query(selectAccountsByUser, userID).WithContext(ctx).Consistency(gocql.One).Iter()
	var accountIDs []uuid.UUID
	for {
		var userID2, accountID uuid.UUID
		if !iter.Scan(&userID2, &accountID) {
			break
		}
		accountIDs = append(accountIDs, accountID)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("find accounts by user: %w", err)
	}
	return accountIDs, nil
}
