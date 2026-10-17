package models

const (
	EventTypeAccountDebitRequested  = "account.debit.requested"
	EventTypeAccountCreditRequested = "account.credit.requested"

	EventTypeAccountDebited      = "account.debited"
	EventTypeAccountDebitFailed  = "account.debit.failed"
	EventTypeAccountCredited     = "account.credited"
	EventTypeAccountCreditFailed = "account.credit.failed"

	// EventTypeUserCreated is emitted by the Auth Service on registration and
	// consumed by the Account Service to provision the user's account.
	EventTypeUserCreated = "auth.user.created"

	EventVersionV1 = 1
)

// UserCreatedEvent announces a newly registered user. The Account Service
// consumes it to create the account row with the configured initial balance.
type UserCreatedEvent struct {
	EventID      string `json:"eventId"`
	EventType    string `json:"eventType"` // auth.user.created
	EventVersion int    `json:"eventVersion"`
	UserID       string `json:"userId"`
	AccountID    string `json:"accountId"`
	OccurredAt   string `json:"occurredAt"`
}

// AccountCommand is a saga command requesting a balance movement. It is produced
// by the Transfer Service and consumed by the Account Service.
type AccountCommand struct {
	EventID               string `json:"eventId"`
	EventType             string `json:"eventType"` // account.debit.requested | account.credit.requested
	EventVersion          int    `json:"eventVersion"`
	TransferID            string `json:"transferId"`
	IdempotencyKey        string `json:"idempotencyKey"`
	CorrelationID         string `json:"correlationId"`
	AccountID             string `json:"accountId"`
	CounterpartyAccountID string `json:"counterpartyAccountId"`
	AmountCents           int64  `json:"amountCents"`
	OccurredAt            string `json:"occurredAt"`
}

// AccountEvent is the result of applying an AccountCommand, produced by the
// Account Service and consumed by the Transfer Service saga.
type AccountEvent struct {
	EventID               string `json:"eventId"`
	EventType             string `json:"eventType"` // account.debited | account.debit.failed | account.credited | account.credit.failed
	EventVersion          int    `json:"eventVersion"`
	TransferID            string `json:"transferId"`
	IdempotencyKey        string `json:"idempotencyKey"`
	CorrelationID         string `json:"correlationId"`
	AccountID             string `json:"accountId"`
	CounterpartyAccountID string `json:"counterpartyAccountId"`
	AmountCents           int64  `json:"amountCents"`
	BalanceCents          int64  `json:"balanceCents,omitempty"`
	FailureReason         string `json:"failureReason,omitempty"`
	OccurredAt            string `json:"occurredAt"`
}
