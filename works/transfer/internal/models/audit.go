package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditEvent is a single row in the transfer_keyspace audit_events table.
type AuditEvent struct {
	Day           time.Time
	Timestamp     time.Time
	EventID       uuid.UUID
	ActorUserID   *uuid.UUID
	Action        string
	ResourceType  string
	ResourceID    uuid.UUID
	CorrelationID string
	Metadata      map[string]string
}
