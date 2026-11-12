package models

import (
	"time"

	"github.com/google/uuid"
)

// ProcessedCommand is the Account Service's own record of a saga command it has
// seen, used to make balance movements idempotent under at-least-once delivery.
type ProcessedCommand struct {
	IdempotencyKey  string
	TransferID      uuid.UUID
	AccountID       uuid.UUID
	EventType       string
	ResultEventType string
	BalanceCents    *int64
	FailureReason   *string
	CreatedAt       time.Time
	AppliedAt       *time.Time
}

// AuditEvent is a single row in the account_keyspace audit_events table.
type AuditEvent struct {
	Day           time.Time
	Timestamp     time.Time
	EventID       uuid.UUID
	Action        string
	ResourceID    uuid.UUID
	CorrelationID string
	Metadata      map[string]string
}
