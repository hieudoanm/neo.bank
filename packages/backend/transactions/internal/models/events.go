package models

const (
	EventTypeTransferInitiated = "transfer_initiated"
	EventTypeTransferCompleted = "transfer_completed"
	EventTypeTransferFailed    = "transfer_failed"

	EventTypeAccountDebited      = "account.debited"
	EventTypeAccountDebitFailed  = "account.debit.failed"
	EventTypeAccountCredited     = "account.credited"
	EventTypeAccountCreditFailed = "account.credit.failed"

	EventVersionV1 = 1
)

// Movement identifiers, derived from the command idempotency key suffix. They
// distinguish the rows a single transfer produces in the read model.
const (
	MovementDebit        = "debit"
	MovementCredit       = "credit"
	MovementCompensation = "compensation"
)

// AccountEvent is a balance movement result produced by the Account Service. It
// is the only event that mutates the transaction read model (account movements
// carry the affected account, the counterparty, and the amount).
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

// TransferEvent is a transfer lifecycle event. The read model consumes
// transfer_completed/transfer_failed to finalise the status of the rows a
// transfer produced.
type TransferEvent struct {
	EventID              string `json:"eventId"`
	EventType            string `json:"eventType"` // transfer_initiated | transfer_completed | transfer_failed
	EventVersion         int    `json:"eventVersion"`
	TransferID           string `json:"transferId"`
	IdempotencyKey       string `json:"idempotencyKey"`
	CorrelationID        string `json:"correlationId"`
	SourceAccountID      string `json:"sourceAccountId"`
	DestinationAccountID string `json:"destinationAccountId"`
	AmountCents          int64  `json:"amountCents"`
	OccurredAt           string `json:"occurredAt"`
}
