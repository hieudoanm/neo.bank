package services

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/core/errors"
)

// encodeCursor packs the clustering key of the last row of a page
// (created_at, transaction_id) into an opaque, URL-safe token.
func encodeCursor(createdAt time.Time, transactionID uuid.UUID) string {
	raw := createdAt.UTC().Format(time.RFC3339Nano) + "|" + transactionID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// decodeCursor reverses encodeCursor. It returns an error wrapping
// errors.ErrInvalidCursor when the token is malformed.
func decodeCursor(token string) (time.Time, uuid.UUID, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, uuid.Nil, errors.ErrInvalidCursor
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, errors.ErrInvalidCursor
	}
	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, errors.ErrInvalidCursor
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("%w", errors.ErrInvalidCursor)
	}
	return createdAt, id, nil
}
