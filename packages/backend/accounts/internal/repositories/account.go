package repositories

import (
	"context"
	"fmt"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
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
	insertAccount        = `INSERT INTO accounts (account_id, user_id, balance_cents, currency_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) IF NOT EXISTS`
	insertAccountByUser  = `INSERT INTO accounts_by_user (user_id, account_id, created_at) VALUES (?, ?, ?)`
	selectAccountByID    = `SELECT account_id, user_id, balance_cents, currency_code, created_at, updated_at FROM accounts WHERE account_id = ?`
	selectAccountsByUser = `SELECT account_id FROM accounts_by_user WHERE user_id = ?`
	debitAccount         = `UPDATE accounts SET balance_cents = balance_cents - ?, updated_at = ? WHERE account_id = ? IF balance_cents >= ?`
	creditAccount        = `UPDATE accounts SET balance_cents = balance_cents + ?, updated_at = ? WHERE account_id = ? IF EXISTS`

	claimCommand        = `INSERT INTO processed_commands (idempotency_key, transfer_id, account_id, event_type, created_at) VALUES (?, ?, ?, ?, ?) IF NOT EXISTS`
	selectCommandResult = `SELECT result_event_type, balance_cents, failure_reason, applied_at FROM processed_commands WHERE idempotency_key = ?`
	updateCommandResult = `UPDATE processed_commands SET result_event_type = ?, balance_cents = ?, failure_reason = ?, applied_at = ? WHERE idempotency_key = ?`

	insertAudit = `INSERT INTO audit_events (service, day, timestamp, event_id, actor_user_id, action, resource_type, resource_id, correlation_id, metadata) VALUES ('account', ?, ?, ?, null, ?, 'account', ?, ?, ?)`
)

func (r *AccountRepository) FindByID(ctx context.Context, accountID uuid.UUID) (*models.Account, error) {
	var a models.Account
	err := r.session.Query(selectAccountByID, accountID).
		WithContext(ctx).Consistency(gocql.One).
		Scan(&a.ID, &a.UserID, &a.BalanceCents, &a.CurrencyCode, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("find account by id: %w", err)
	}
	return &a, nil
}

func (r *AccountRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	iter := r.session.Query(selectAccountsByUser, userID).WithContext(ctx).Consistency(gocql.One).Iter()
	var accountIDs []uuid.UUID
	for {
		var accountID uuid.UUID
		if !iter.Scan(&accountID) {
			break
		}
		accountIDs = append(accountIDs, accountID)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("find accounts by user: %w", err)
	}
	return accountIDs, nil
}

// CreateAccount provisions an account owned by userID with the given initial
// balance and currency. The write is idempotent (INSERT ... IF NOT EXISTS) so
// at-least-once delivery of the provisioning event cannot create duplicate
// accounts or reset an existing balance. It returns true when this call created
// the account and false when the account already existed.
func (r *AccountRepository) CreateAccount(ctx context.Context, accountID, userID uuid.UUID, balanceCents int64, currencyCode string, now time.Time) (bool, error) {
	applied, err := r.session.Query(insertAccount, accountID, userID, balanceCents, currencyCode, now, now).
		WithContext(ctx).Consistency(gocql.Quorum).SerialConsistency(gocql.LocalSerial).ScanCAS()
	if err != nil {
		return false, fmt.Errorf("create account: %w", err)
	}
	// The user link is an idempotent upsert written even when the account row
	// already exists, so a retry after a partial failure cannot leave an
	// account that FindByUserID can never reach.
	if err := r.session.Query(insertAccountByUser, userID, accountID, now).
		WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return false, fmt.Errorf("link account to user: %w", err)
	}
	return applied, nil
}

// Debit atomically subtracts amount from the account balance using a
// lightweight transaction that refuses to go negative. It returns the resulting
// balance and errors.ErrInsufficientFunds when funds are insufficient or
// errors.ErrAccountNotFound when the account does not exist.
func (r *AccountRepository) Debit(ctx context.Context, accountID uuid.UUID, amountCents int64) (int64, error) {
	if _, err := r.FindByID(ctx, accountID); err != nil {
		return 0, err
	}
	var balance *int64
	var updatedAt *time.Time
	applied, err := r.session.Query(debitAccount, amountCents, time.Now().UTC(), accountID, amountCents).
		WithContext(ctx).Consistency(gocql.Quorum).SerialConsistency(gocql.LocalSerial).ScanCAS(&balance, &updatedAt)
	if err != nil {
		return 0, fmt.Errorf("debit account: %w", err)
	}
	if !applied {
		return 0, errors.ErrInsufficientFunds
	}
	return derefInt64(balance), nil
}

// Credit atomically adds amount to the account balance. It returns the resulting
// balance and errors.ErrAccountNotFound when the account does not exist.
func (r *AccountRepository) Credit(ctx context.Context, accountID uuid.UUID, amountCents int64) (int64, error) {
	var balance *int64
	var updatedAt *time.Time
	applied, err := r.session.Query(creditAccount, amountCents, time.Now().UTC(), accountID).
		WithContext(ctx).Consistency(gocql.Quorum).SerialConsistency(gocql.LocalSerial).ScanCAS(&balance, &updatedAt)
	if err != nil {
		return 0, fmt.Errorf("credit account: %w", err)
	}
	if !applied {
		return 0, errors.ErrAccountNotFound
	}
	return derefInt64(balance), nil
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// ClaimCommand records a pending marker for a saga command using IF NOT EXISTS.
// It returns true when this consumer owns the command and may apply it.
func (r *AccountRepository) ClaimCommand(ctx context.Context, key string, transferID, accountID uuid.UUID, eventType string, createdAt time.Time) (bool, error) {
	applied, err := r.session.Query(claimCommand, key, transferID, accountID, eventType, createdAt).
		WithContext(ctx).Consistency(gocql.Quorum).SerialConsistency(gocql.LocalSerial).ScanCAS()
	if err != nil {
		return false, fmt.Errorf("claim command: %w", err)
	}
	return applied, nil
}

// FindCommandResult returns the recorded result for a command, or nil when the
// command is still pending (no result recorded yet).
func (r *AccountRepository) FindCommandResult(ctx context.Context, key string) (*models.ProcessedCommand, error) {
	var pc models.ProcessedCommand
	var resultEventType *string
	var balanceCents *int64
	var failureReason *string
	var appliedAt *time.Time
	err := r.session.Query(selectCommandResult, key).
		WithContext(ctx).Consistency(gocql.Quorum).
		Scan(&resultEventType, &balanceCents, &failureReason, &appliedAt)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find command result: %w", err)
	}
	if resultEventType != nil {
		pc.ResultEventType = *resultEventType
	}
	pc.BalanceCents = balanceCents
	pc.FailureReason = failureReason
	pc.AppliedAt = appliedAt
	return &pc, nil
}

func (r *AccountRepository) RecordCommandResult(ctx context.Context, key, resultEventType string, balanceCents *int64, failureReason *string, appliedAt time.Time) error {
	if err := r.session.Query(updateCommandResult, resultEventType, balanceCents, failureReason, appliedAt, key).
		WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("record command result: %w", err)
	}
	return nil
}

func (r *AccountRepository) WriteAudit(ctx context.Context, ev models.AuditEvent) error {
	if err := r.session.Query(insertAudit, ev.Day, ev.Timestamp, ev.EventID, ev.Action, ev.ResourceID, ev.CorrelationID, ev.Metadata).
		WithContext(ctx).Consistency(gocql.Quorum).Exec(); err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}
