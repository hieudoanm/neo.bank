package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	transferrors "github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
	"github.com/segmentio/kafka-go"
)

// SagaStore is the persistence the saga needs. *repositories.TransferRepository
// satisfies it in production; tests provide a fake.
type SagaStore interface {
	FindByID(ctx context.Context, transferID uuid.UUID) (*models.Transfer, error)
	Transition(ctx context.Context, t *models.Transfer, fromStatus, toStatus string, failureReason *string) (bool, error)
}

// EventPublisher publishes saga commands and transfer lifecycle events.
type EventPublisher interface {
	PublishAccountCommand(ctx context.Context, cmd models.AccountCommand) error
	PublishTransferEvent(ctx context.Context, ev models.TransferEvent) error
}

// SagaConsumer orchestrates the transfer saga. It reacts to transfer lifecycle
// events and account movement results on the transfers topic, advancing the
// transfer state machine and issuing the next command. Every step is guarded by
// a lightweight status transition and idempotent command keys so duplicate
// delivery cannot double-move funds.
type SagaConsumer struct {
	reader   *kafka.Reader
	dlq      *kafka.Writer
	repo     SagaStore
	producer EventPublisher
	logger   *slog.Logger
}

func NewSagaConsumer(brokers []string, topic, group, dlqTopic string, repo SagaStore, prod EventPublisher, logger *slog.Logger) *SagaConsumer {
	return &SagaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  group,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		dlq: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  dlqTopic,
			Balancer:               &kafka.LeastBytes{},
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: false,
		},
		repo:     repo,
		producer: prod,
		logger:   logger,
	}
}

func (c *SagaConsumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			c.logger.Error("fetch message", "err", err)
			continue
		}
		if err := c.handle(ctx, msg); err != nil {
			c.logger.Error("handle saga event", "err", err, "partition", msg.Partition, "offset", msg.Offset)
			continue // do not commit; retry
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit message", "err", err)
		}
	}
}

func (c *SagaConsumer) handle(ctx context.Context, msg kafka.Message) error {
	var env struct {
		EventType      string `json:"eventType"`
		EventVersion   int    `json:"eventVersion"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		c.logger.Error("unmarshal saga event", "err", err)
		c.sendToDLQ(ctx, msg, "malformed payload")
		return nil
	}
	if env.EventVersion != models.EventVersionV1 {
		c.sendToDLQ(ctx, msg, fmt.Sprintf("unsupported event version %d", env.EventVersion))
		return nil
	}

	switch env.EventType {
	case models.EventTypeTransferInitiated:
		return c.onTransferInitiated(ctx, msg)
	case models.EventTypeAccountDebited:
		return c.onAccountDebited(ctx, msg)
	case models.EventTypeAccountDebitFailed:
		return c.onAccountDebitFailed(ctx, msg)
	case models.EventTypeAccountCredited:
		return c.onAccountCredited(ctx, msg)
	case models.EventTypeAccountCreditFailed:
		return c.onAccountCreditFailed(ctx, msg)
	default:
		return nil
	}
}

func (c *SagaConsumer) onTransferInitiated(ctx context.Context, msg kafka.Message) error {
	var ev models.TransferEvent
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		return err
	}
	t, err := c.loadTransfer(ctx, ev.TransferID)
	if err != nil {
		return err
	}
	if t == nil {
		return nil
	}
	applied, err := c.repo.Transition(ctx, t, models.StatusPending, models.StatusProcessing, nil)
	if err != nil {
		return err
	}
	if !applied {
		return nil
	}
	return c.publishAccountCommand(ctx, models.EventTypeAccountDebitRequested, t, t.SourceAccountID, ":debit")
}

func (c *SagaConsumer) onAccountDebited(ctx context.Context, msg kafka.Message) error {
	ev, err := unmarshalAccountEvent(msg)
	if err != nil {
		return err
	}
	t, err := c.loadTransfer(ctx, ev.TransferID)
	if err != nil {
		return err
	}
	if t == nil || t.Status != models.StatusProcessing {
		return nil
	}
	return c.publishAccountCommand(ctx, models.EventTypeAccountCreditRequested, t, t.DestAccountID, ":credit")
}

func (c *SagaConsumer) onAccountDebitFailed(ctx context.Context, msg kafka.Message) error {
	ev, err := unmarshalAccountEvent(msg)
	if err != nil {
		return err
	}
	t, err := c.loadTransfer(ctx, ev.TransferID)
	if err != nil {
		return err
	}
	if t == nil {
		return nil
	}
	reason := failureReason(ev)
	applied, err := c.repo.Transition(ctx, t, models.StatusProcessing, models.StatusFailed, &reason)
	if err != nil {
		return err
	}
	if !applied {
		return nil
	}
	return c.publishTransferEvent(ctx, models.EventTypeTransferFailed, t)
}

func (c *SagaConsumer) onAccountCredited(ctx context.Context, msg kafka.Message) error {
	ev, err := unmarshalAccountEvent(msg)
	if err != nil {
		return err
	}
	t, err := c.loadTransfer(ctx, ev.TransferID)
	if err != nil {
		return err
	}
	if t == nil || ev.AccountID != t.DestAccountID.String() {
		return nil // ignore compensation credits to the source account
	}
	applied, err := c.repo.Transition(ctx, t, models.StatusProcessing, models.StatusCompleted, nil)
	if err != nil {
		return err
	}
	if !applied {
		return nil
	}
	return c.publishTransferEvent(ctx, models.EventTypeTransferCompleted, t)
}

func (c *SagaConsumer) onAccountCreditFailed(ctx context.Context, msg kafka.Message) error {
	ev, err := unmarshalAccountEvent(msg)
	if err != nil {
		return err
	}
	t, err := c.loadTransfer(ctx, ev.TransferID)
	if err != nil {
		return err
	}
	if t == nil {
		return nil
	}
	if ev.AccountID != t.DestAccountID.String() {
		c.logger.Error("compensation credit failed", "transferId", ev.TransferID, "reason", ev.FailureReason)
		return nil
	}
	if t.Status != models.StatusProcessing {
		return nil
	}
	// Compensate the successful debit by crediting the source account back.
	if err := c.publishAccountCommand(ctx, models.EventTypeAccountCreditRequested, t, t.SourceAccountID, ":compensation"); err != nil {
		return err
	}
	reason := failureReason(ev)
	applied, err := c.repo.Transition(ctx, t, models.StatusProcessing, models.StatusFailed, &reason)
	if err != nil {
		return err
	}
	if !applied {
		return nil
	}
	return c.publishTransferEvent(ctx, models.EventTypeTransferFailed, t)
}

func (c *SagaConsumer) loadTransfer(ctx context.Context, transferID string) (*models.Transfer, error) {
	id, err := uuid.Parse(transferID)
	if err != nil {
		return nil, fmt.Errorf("invalid transfer id: %w", err)
	}
	t, err := c.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferrors.ErrTransferNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

func (c *SagaConsumer) publishAccountCommand(ctx context.Context, eventType string, t *models.Transfer, accountID uuid.UUID, suffix string) error {
	counterparty := t.DestAccountID
	if accountID == t.DestAccountID {
		counterparty = t.SourceAccountID
	}
	cmd := models.AccountCommand{
		EventID:               newUUID().String(),
		EventType:             eventType,
		EventVersion:          models.EventVersionV1,
		TransferID:            t.ID.String(),
		IdempotencyKey:        t.ID.String() + suffix,
		CorrelationID:         t.CorrelationID,
		AccountID:             accountID.String(),
		CounterpartyAccountID: counterparty.String(),
		AmountCents:           t.AmountCents,
		OccurredAt:            time.Now().UTC().Format(time.RFC3339),
	}
	return c.producer.PublishAccountCommand(ctx, cmd)
}

func (c *SagaConsumer) publishTransferEvent(ctx context.Context, eventType string, t *models.Transfer) error {
	ev := models.TransferEvent{
		EventID:              newUUID().String(),
		EventType:            eventType,
		EventVersion:         models.EventVersionV1,
		TransferID:           t.ID.String(),
		IdempotencyKey:       t.IdempotencyKey,
		CorrelationID:        t.CorrelationID,
		SourceAccountID:      t.SourceAccountID.String(),
		DestinationAccountID: t.DestAccountID.String(),
		AmountCents:          t.AmountCents,
		OccurredAt:           time.Now().UTC().Format(time.RFC3339),
	}
	return c.producer.PublishTransferEvent(ctx, ev)
}

func (c *SagaConsumer) sendToDLQ(ctx context.Context, msg kafka.Message, reason string) {
	c.logger.Error("routing to dlq", "reason", reason, "partition", msg.Partition, "offset", msg.Offset)
	if c.dlq == nil {
		return
	}
	if err := c.dlq.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Headers: append(msg.Headers, kafka.Header{
			Key:   "failure_reason",
			Value: []byte(reason),
		}),
	}); err != nil {
		c.logger.Error("write dlq", "err", err)
	}
}

func (c *SagaConsumer) Close() error {
	_ = c.dlq.Close()
	return c.reader.Close()
}

func unmarshalAccountEvent(msg kafka.Message) (models.AccountEvent, error) {
	var ev models.AccountEvent
	err := json.Unmarshal(msg.Value, &ev)
	return ev, err
}

func failureReason(ev models.AccountEvent) string {
	if ev.FailureReason != "" {
		return ev.FailureReason
	}
	return "account movement failed"
}

func newUUID() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}
