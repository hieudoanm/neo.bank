package repositories

import (
	"context"
	"fmt"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
)

type TransferRepository struct {
	session *gocql.Session
}

func NewTransferRepository(session *gocql.Session) *TransferRepository {
	return &TransferRepository{session: session}
}

const (
	insertIdempotency = `INSERT INTO idempotency_keys (idempotency_key, transfer_id, request_hash, created_at) VALUES (?, ?, ?, ?) IF NOT EXISTS`
	getIdempotency    = `SELECT idempotency_key, transfer_id, request_hash, created_at FROM idempotency_keys WHERE idempotency_key = ?`
	insertTransfer    = `INSERT INTO transfers (transfer_id, idempotency_key, source_account_id, destination_account_id, amount_cents, currency_code, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?)`
	selectTransfer    = `SELECT transfer_id, idempotency_key, source_account_id, destination_account_id, amount_cents, currency_code, status, failure_reason, created_at, updated_at FROM transfers WHERE transfer_id = ?`
	insertByAccount   = `INSERT INTO transfers_by_account (account_id, transfer_id, status, created_at) VALUES (?, ?, 'pending', ?)`
)

func (r *TransferRepository) ReserveIdempotency(ctx context.Context, key, transferID uuid.UUID, requestHash string) (bool, uuid.UUID, error) {
	now := time.Now().UTC()
	applied, err := r.session.Query(insertIdempotency, key, transferID, requestHash, now).WithContext(ctx).Consistency(gocql.One).SerialConsistency(gocql.Serial).ScanCAS()
	if err != nil {
		return false, uuid.Nil, fmt.Errorf("reserve idempotency: %w", err)
	}
	if !applied {
		// Lookup existing
		var existingKey, existingTransfer uuid.UUID
		var existingHash string
		var createdAt time.Time
		err := r.session.Query(getIdempotency, key).WithContext(ctx).Consistency(gocql.One).Scan(&existingKey, &existingTransfer, &existingHash, &createdAt)
		if err != nil {
			return false, uuid.Nil, errors.ErrInternal
		}
		if existingHash != requestHash {
			return false, uuid.Nil, errors.ErrIdempotencyConflict
		}
		return false, existingTransfer, nil
	}
	return true, transferID, nil
}

func (r *TransferRepository) Create(ctx context.Context, t *models.Transfer) error {
	now := time.Now().UTC()
	err := r.session.Query(insertTransfer, t.ID, t.IdempotencyKey, t.SourceAccountID, t.DestAccountID, t.AmountCents, t.CurrencyCode, now, now).WithContext(ctx).Consistency(gocql.One).Exec()
	if err != nil {
		return fmt.Errorf("insert transfer: %w", err)
	}
	return r.session.Query(insertByAccount, t.SourceAccountID, t.ID, now).WithContext(ctx).Consistency(gocql.One).Exec()
}

func (r *TransferRepository) FindByID(ctx context.Context, transferID uuid.UUID) (*models.Transfer, error) {
	var t models.Transfer
	var createdAt, updatedAt time.Time
	var failure string
	err := r.session.Query(selectTransfer, transferID).WithContext(ctx).Consistency(gocql.One).Scan(
		&t.ID, &t.IdempotencyKey, &t.SourceAccountID, &t.DestAccountID, &t.AmountCents, &t.CurrencyCode, &t.Status, &failure, &createdAt, &updatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrTransferNotFound
		}
		return nil, fmt.Errorf("find transfer: %w", err)
	}
	t.FailureReason = failure
	t.CreatedAt = createdAt
	t.UpdatedAt = updatedAt
	return &t, nil
}
