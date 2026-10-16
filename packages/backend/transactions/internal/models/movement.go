package models

import (
	"time"

	"github.com/google/uuid"
)

// TransferMovement links a transfer to the transaction row it produced for one
// movement (debit, credit, or compensation). It is the transaction service's
// own lookup so transfer lifecycle events can finalise the status of the rows
// they created, without a secondary index on transfer_id.
type TransferMovement struct {
	TransferID    uuid.UUID
	Movement      string
	TransactionID uuid.UUID
	AccountID     uuid.UUID
	CreatedAt     time.Time
}
