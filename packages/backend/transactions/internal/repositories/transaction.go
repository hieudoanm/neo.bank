package repositories

import (
	"context"
	"fmt"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/models"
)

type TransactionRepository struct {
	session *gocql.Session
}

func NewTransactionRepository(session *gocql.Session) *TransactionRepository {
	return &TransactionRepository{session: session}
}

// ListParams bounds a page of transaction history. AfterCreatedAt/AfterTransactionID
// are the clustering key of the last row of the previous page (nil for the first page).
type ListParams struct {
	Limit              int
	AfterCreatedAt     *time.Time
	AfterTransactionID *uuid.UUID
}

const (
	selectTransactionsFirstPage = `SELECT account_id, created_at, transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status FROM transactions WHERE account_id = ? LIMIT ?`
	selectTransactionsAfter     = `SELECT account_id, created_at, transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status FROM transactions WHERE account_id = ? AND (created_at, transaction_id) > (?, ?) LIMIT ?`

	insertTransaction = `INSERT INTO transactions (account_id, created_at, transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	updateStatus      = `UPDATE transactions SET status = ? WHERE account_id = ? AND created_at = ? AND transaction_id = ?`

	selectMovement      = `SELECT transaction_id, account_id, created_at FROM transfer_movements WHERE transfer_id = ? AND movement = ?`
	selectMovementsByTx = `SELECT movement, transaction_id, account_id, created_at FROM transfer_movements WHERE transfer_id = ?`
	upsertMovement      = `INSERT INTO transfer_movements (transfer_id, movement, transaction_id, account_id, created_at) VALUES (?, ?, ?, ?, ?)`

	insertAudit = `INSERT INTO audit_events (service, day, timestamp, event_id, actor_user_id, action, resource_type, resource_id, correlation_id, metadata) VALUES ('transaction', ?, ?, ?, null, ?, 'transaction', ?, ?, ?)`
)

// ListByAccount returns up to params.Limit transactions for an account, newest
// first, ordered by (created_at, transaction_id) descending.
func (r *TransactionRepository) ListByAccount(ctx context.Context, accountID uuid.UUID, params ListParams) ([]models.Transaction, error) {
	var (
		iter *gocql.Iter
	)
	if params.AfterCreatedAt != nil && params.AfterTransactionID != nil {
		iter = r.session.Query(selectTransactionsAfter, accountID, *params.AfterCreatedAt, *params.AfterTransactionID, params.Limit).
			WithContext(ctx).Consistency(gocql.Quorum).Iter()
	} else {
		iter = r.session.Query(selectTransactionsFirstPage, accountID, params.Limit).
			WithContext(ctx).Consistency(gocql.Quorum).Iter()
	}

	var out []models.Transaction
	for {
		var t models.Transaction
		var transferID *uuid.UUID
		if !iter.Scan(&t.AccountID, &t.CreatedAt, &t.TransactionID, &transferID, &t.Type, &t.AmountCents, &t.CounterpartyAccountID, &t.Status) {
			break
		}
		t.TransferID = transferID
		out = append(out, t)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	return out, nil
}

func (r *TransactionRepository) InsertTransaction(ctx context.Context, t models.Transaction) error {
	var transferID *uuid.UUID
	if t.TransferID != nil {
		transferID = t.TransferID
	}
	if err := r.session.Query(insertTransaction, t.AccountID, t.CreatedAt, t.TransactionID, transferID, t.Type, t.AmountCents, t.CounterpartyAccountID, t.Status).
		WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, accountID uuid.UUID, createdAt time.Time, transactionID uuid.UUID, status string) error {
	if err := r.session.Query(updateStatus, status, accountID, createdAt, transactionID).
		WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}
	return nil
}

// FindMovement returns the recorded movement for a transfer, or nil when the
// movement has not been projected yet.
func (r *TransactionRepository) FindMovement(ctx context.Context, transferID uuid.UUID, movement string) (*models.TransferMovement, error) {
	var m models.TransferMovement
	err := r.session.Query(selectMovement, transferID, movement).
		WithContext(ctx).Consistency(gocql.Quorum).
		Scan(&m.TransactionID, &m.AccountID, &m.CreatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find movement: %w", err)
	}
	m.TransferID = transferID
	m.Movement = movement
	return &m, nil
}

func (r *TransactionRepository) ListMovements(ctx context.Context, transferID uuid.UUID) ([]models.TransferMovement, error) {
	iter := r.session.Query(selectMovementsByTx, transferID).
		WithContext(ctx).Consistency(gocql.Quorum).Iter()
	var out []models.TransferMovement
	for {
		var m models.TransferMovement
		if !iter.Scan(&m.Movement, &m.TransactionID, &m.AccountID, &m.CreatedAt) {
			break
		}
		m.TransferID = transferID
		out = append(out, m)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("list movements: %w", err)
	}
	return out, nil
}

func (r *TransactionRepository) UpsertMovement(ctx context.Context, m models.TransferMovement) error {
	if err := r.session.Query(upsertMovement, m.TransferID, m.Movement, m.TransactionID, m.AccountID, m.CreatedAt).
		WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("upsert movement: %w", err)
	}
	return nil
}

func (r *TransactionRepository) WriteAudit(ctx context.Context, ev models.AuditEvent) error {
	if err := r.session.Query(insertAudit, ev.Day, ev.Timestamp, ev.EventID, ev.Action, ev.ResourceID, ev.CorrelationID, ev.Metadata).
		WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}
