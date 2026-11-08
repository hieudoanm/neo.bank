package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
)

const maxAmountCents = 9999999999

// TransferStore abstracts persistence over the transfer repository so the
// service can be unit tested with an in-memory fake.
type TransferStore interface {
	ReserveIdempotency(ctx context.Context, key string, transferID uuid.UUID, requestHash string) (bool, uuid.UUID, error)
	FindByID(ctx context.Context, transferID uuid.UUID) (*models.Transfer, error)
	Create(ctx context.Context, t *models.Transfer) error
	Delete(ctx context.Context, t *models.Transfer) error
}

// EventPublisher abstracts the Kafka producer so the service can be unit
// tested without a broker.
type EventPublisher interface {
	PublishTransferEvent(ctx context.Context, ev models.TransferEvent) error
}

type TransferService struct {
	repo     TransferStore
	producer EventPublisher
}

func NewTransferService(repo TransferStore, p EventPublisher) *TransferService {
	return &TransferService{repo: repo, producer: p}
}

// InitiateTransfer creates a pending transfer and emits transfer_initiated.
// It is idempotent: replaying the same Idempotency-Key with an identical
// payload returns the original transfer without re-executing, while reusing the
// key with a different payload is rejected with ErrIdempotencyKeyExists.
func (s *TransferService) InitiateTransfer(ctx context.Context, sourceAccountID string, req models.CreateTransferRequest, idempotencyKey string) (*models.TransferResponse, error) {
	if idempotencyKey == "" {
		return nil, errors.ValidationError(map[string]string{"Idempotency-Key": "required"})
	}
	if req.AmountCents <= 0 || req.AmountCents > maxAmountCents {
		return nil, errors.ErrInvalidAmount
	}
	dest, err := uuid.Parse(req.DestinationAccountId)
	if err != nil {
		return nil, errors.ErrInvalidDestination
	}
	src, err := uuid.Parse(sourceAccountID)
	if err != nil {
		return nil, errors.ErrAccountNotFound
	}
	if src == dest {
		return nil, errors.ErrSameAccount
	}

	requestHash := hashRequest(src.String(), dest.String(), req.AmountCents)
	transferID := newUUID()
	correlationID := newUUID().String()

	reserved, existingID, err := s.repo.ReserveIdempotency(ctx, idempotencyKey, transferID, requestHash)
	if err != nil {
		return nil, err
	}
	if !reserved {
		existing, err := s.repo.FindByID(ctx, existingID)
		if err != nil {
			return nil, err
		}
		return models.NewTransferResponse(existing), nil
	}

	now := time.Now().UTC()
	t := &models.Transfer{
		ID:              transferID,
		SourceAccountID: src,
		DestAccountID:   dest,
		AmountCents:     req.AmountCents,
		Status:          models.StatusPending,
		IdempotencyKey:  idempotencyKey,
		CorrelationID:   correlationID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, errors.ErrInternal
	}

	ev := models.TransferEvent{
		EventID:              newUUID().String(),
		EventType:            models.EventTypeTransferInitiated,
		EventVersion:         models.EventVersionV1,
		TransferID:           t.ID.String(),
		IdempotencyKey:       idempotencyKey,
		CorrelationID:        correlationID,
		SourceAccountID:      src.String(),
		DestinationAccountID: dest.String(),
		AmountCents:          t.AmountCents,
		OccurredAt:           now.Format(time.RFC3339),
	}
	if err := s.producer.PublishTransferEvent(ctx, ev); err != nil {
		if rollbackErr := s.repo.Delete(ctx, t); rollbackErr != nil {
			return nil, errors.ErrInternal
		}
		return nil, errors.ErrServiceUnavailable
	}

	return models.NewTransferResponse(t), nil
}

// GetTransfer returns a transfer only when the caller owns the source account.
// A transfer belonging to another account is reported as not found rather than
// disclosing its existence.
func (s *TransferService) GetTransfer(ctx context.Context, transferID, callerAccountID string) (*models.TransferResponse, error) {
	id, err := uuid.Parse(transferID)
	if err != nil {
		return nil, errors.ErrTransferNotFound
	}
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.SourceAccountID.String() != callerAccountID && t.DestAccountID.String() != callerAccountID {
		return nil, errors.ErrTransferNotFound
	}
	return models.NewTransferResponse(t), nil
}

func newUUID() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}

func hashRequest(source, dest string, amount int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d", source, dest, amount)))
	return hex.EncodeToString(sum[:])
}
