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

const (
	idempotencyTTL = 24 * time.Hour

	insertIdempotencyPending = `INSERT INTO idempotency_keys (idempotency_key, operation_type, resource_id, request_hash, response_status, response_body, expires_at, created_at) VALUES (?, 'transfer', ?, ?, null, null, ?, ?) IF NOT EXISTS USING TTL ?`
	selectIdempotency        = `SELECT resource_id, request_hash FROM idempotency_keys WHERE idempotency_key = ?`

	insertTransfer   = `INSERT INTO transfers (transfer_id, source_account_id, destination_account_id, amount_cents, status, idempotency_key, correlation_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	selectTransfer   = `SELECT transfer_id, source_account_id, destination_account_id, amount_cents, status, idempotency_key, correlation_id, created_at, updated_at, failure_reason FROM transfers WHERE transfer_id = ?`
	deleteTransfer   = `DELETE FROM transfers WHERE transfer_id = ?`
	transitionStatus = `UPDATE transfers SET status = ?, failure_reason = ?, updated_at = ? WHERE transfer_id = ? IF status = ?`
	updateByAccount  = `UPDATE transfers_by_account SET status = ?, failure_reason = ?, updated_at = ? WHERE account_id = ? AND transfer_id = ?`
	insertByAccount  = `INSERT INTO transfers_by_account (account_id, transfer_id, direction, counterparty_account_id, amount_cents, status, correlation_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	deleteByAccount  = `DELETE FROM transfers_by_account WHERE account_id = ? AND transfer_id = ?`
	deleteIdem       = `DELETE FROM idempotency_keys WHERE idempotency_key = ?`
)

type TransferRepository struct {
	session *gocql.Session
}

func NewTransferRepository(session *gocql.Session) *TransferRepository {
	return &TransferRepository{session: session}
}

// ReserveIdempotency atomically records a pending marker for key. It returns
// reserved=true when this caller owns the key and may proceed. When the key
// already exists it returns reserved=false together with the existing transfer
// ID, or ErrIdempotencyKeyExists when the same key was used with a different
// request payload.
func (r *TransferRepository) ReserveIdempotency(ctx context.Context, key string, transferID uuid.UUID, requestHash string) (bool, uuid.UUID, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(idempotencyTTL)
	applied, err := r.session.Query(insertIdempotencyPending, key, transferID, requestHash, expiresAt, now, int(idempotencyTTL.Seconds())).
		WithContext(ctx).Consistency(gocql.One).SerialConsistency(gocql.LocalSerial).ScanCAS()
	if err != nil {
		return false, uuid.Nil, fmt.Errorf("reserve idempotency: %w", err)
	}
	if applied {
		return true, transferID, nil
	}

	var existingResourceID uuid.UUID
	var existingHash string
	if err := r.session.Query(selectIdempotency, key).WithContext(ctx).Consistency(gocql.One).Scan(&existingResourceID, &existingHash); err != nil {
		if err == gocql.ErrNotFound {
			return false, uuid.Nil, errors.ErrInternal
		}
		return false, uuid.Nil, fmt.Errorf("read idempotency: %w", err)
	}
	if existingHash != requestHash {
		return false, uuid.Nil, errors.ErrIdempotencyKeyExists
	}
	return false, existingResourceID, nil
}

func (r *TransferRepository) Create(ctx context.Context, t *models.Transfer) error {
	if err := r.session.Query(insertTransfer, t.ID, t.SourceAccountID, t.DestAccountID, t.AmountCents, t.Status, t.IdempotencyKey, t.CorrelationID, t.CreatedAt, t.UpdatedAt).
		WithContext(ctx).Consistency(gocql.One).Exec(); err != nil {
		return fmt.Errorf("insert transfer: %w", err)
	}
	if err := r.insertLookup(ctx, t.SourceAccountID, t, "source", t.DestAccountID); err != nil {
		return err
	}
	return r.insertLookup(ctx, t.DestAccountID, t, "destination", t.SourceAccountID)
}

func (r *TransferRepository) insertLookup(ctx context.Context, accountID uuid.UUID, t *models.Transfer, direction string, counterparty uuid.UUID) error {
	if err := r.session.Query(insertByAccount, accountID, t.ID, direction, counterparty, t.AmountCents, t.Status, t.CorrelationID, t.CreatedAt, t.UpdatedAt).
		WithContext(ctx).Consistency(gocql.One).Exec(); err != nil {
		return fmt.Errorf("insert transfers_by_account: %w", err)
	}
	return nil
}

// Delete removes a pending transfer and its lookup rows, and releases the
// idempotency marker. Used to roll back when publishing to Kafka fails.
func (r *TransferRepository) Delete(ctx context.Context, t *models.Transfer) error {
	if err := r.session.Query(deleteTransfer, t.ID).WithContext(ctx).Consistency(gocql.One).Exec(); err != nil {
		return fmt.Errorf("delete transfer: %w", err)
	}
	for _, accountID := range []uuid.UUID{t.SourceAccountID, t.DestAccountID} {
		if err := r.session.Query(deleteByAccount, accountID, t.ID).WithContext(ctx).Consistency(gocql.One).Exec(); err != nil {
			return fmt.Errorf("delete transfers_by_account: %w", err)
		}
	}
	if err := r.session.Query(deleteIdem, t.IdempotencyKey).WithContext(ctx).Consistency(gocql.One).Exec(); err != nil {
		return fmt.Errorf("delete idempotency: %w", err)
	}
	return nil
}

func (r *TransferRepository) FindByID(ctx context.Context, transferID uuid.UUID) (*models.Transfer, error) {
	var t models.Transfer
	var failureReason *string
	err := r.session.Query(selectTransfer, transferID).WithContext(ctx).Consistency(gocql.One).Scan(
		&t.ID, &t.SourceAccountID, &t.DestAccountID, &t.AmountCents, &t.Status, &t.IdempotencyKey, &t.CorrelationID, &t.CreatedAt, &t.UpdatedAt, &failureReason)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrTransferNotFound
		}
		return nil, fmt.Errorf("find transfer: %w", err)
	}
	if failureReason != nil {
		t.FailureReason = *failureReason
	}
	return &t, nil
}

// Transition atomically moves a transfer from fromStatus to toStatus using a
// lightweight transaction, then keeps the per-account lookup rows in sync. It
// returns false when the transfer is not currently in fromStatus, which makes
// saga steps idempotent under at-least-once delivery.
func (r *TransferRepository) Transition(ctx context.Context, t *models.Transfer, fromStatus, toStatus string, failureReason *string) (bool, error) {
	now := time.Now().UTC()
	applied, err := r.session.Query(transitionStatus, toStatus, failureReason, now, t.ID, fromStatus).
		WithContext(ctx).Consistency(gocql.Quorum).SerialConsistency(gocql.LocalSerial).ScanCAS()
	if err != nil {
		return false, fmt.Errorf("transition transfer: %w", err)
	}
	if !applied {
		return false, nil
	}
	for _, accountID := range []uuid.UUID{t.SourceAccountID, t.DestAccountID} {
		if err := r.session.Query(updateByAccount, toStatus, failureReason, now, accountID, t.ID).
			WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
			return false, fmt.Errorf("update transfers_by_account status: %w", err)
		}
	}
	return true, nil
}
