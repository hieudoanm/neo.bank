package services

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/producer"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/repositories"
)

type TransferService struct {
	repo     *repositories.TransferRepository
	producer *producer.Producer
}

func NewTransferService(repo *repositories.TransferRepository, p *producer.Producer) *TransferService {
	return &TransferService{repo: repo, producer: p}
}

func (s *TransferService) CreateTransfer(ctx context.Context, sourceUserID, sourceAccountID, destAccountIDStr string, amount int64, idempotencyKey string) (*models.CreateTransferResponse, error) {
	if amount <= 0 { return nil, errors.ErrInvalidAmount }
	dest, err := uuid.Parse(destAccountIDStr)
	if err != nil { return nil, errors.ErrInvalidDestination }
	src, err := uuid.Parse(sourceAccountID)
	if err != nil { return nil, errors.ErrAccountNotFound }
	if src == dest { return nil, errors.ErrSameAccount }

	key, err := uuid.Parse(idempotencyKey)
	if err != nil { key = uuid.New() }

	reqHash := hashBody(sourceAccountID, destAccountIDStr, amount)
	reserved, existingID, err := s.repo.ReserveIdempotency(ctx, key, uuid.New(), reqHash)
	if err != nil { return nil, err }
	if !reserved {
		// Existing transfer
		t, err := s.repo.FindByID(ctx, existingID)
		if err != nil { return nil, err }
		return &models.CreateTransferResponse{TransferId: t.ID.String(), Status: t.Status, SourceAccount: t.SourceAccountID.String(), DestAccount: t.DestAccountID.String(), AmountCents: t.AmountCents, Currency: t.CurrencyCode}, nil
	}

	t := &models.Transfer{ID: uuid.New(), IdempotencyKey: key, SourceAccountID: src, DestAccountID: dest, AmountCents: amount, CurrencyCode: "USD", Status: "pending"}
	if err := s.repo.Create(ctx, t); err != nil { return nil, errors.ErrInternal }

	s.producer.PublishTransferInitiated(ctx, producer.TransferEvent{
		TransferID: t.ID.String(), IdempotencyKey: key.String(), SourceAccountID: src.String(), DestinationAccountID: dest.String(), AmountCents: amount, CurrencyCode: "USD",
	})

	return &models.CreateTransferResponse{TransferId: t.ID.String(), Status: "pending", SourceAccount: src.String(), DestAccount: dest.String(), AmountCents: amount, Currency: "USD"}, nil
}

func hashBody(source, dest string, amount int64) string {
	h := fnv.New64a()
	h.Write([]byte(source + dest + fmt.Sprintf("%d", amount)))
	return fmt.Sprintf("%x", h.Sum64())
}
