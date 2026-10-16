package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/models"
	"github.com/segmentio/kafka-go"
)

// CommandStore is the persistence the account consumer needs.
// *repositories.AccountRepository satisfies it in production; tests provide a fake.
type CommandStore interface {
	ClaimCommand(ctx context.Context, key string, transferID, accountID uuid.UUID, eventType string, createdAt time.Time) (bool, error)
	FindCommandResult(ctx context.Context, key string) (*models.ProcessedCommand, error)
	Debit(ctx context.Context, accountID uuid.UUID, amountCents int64) (int64, error)
	Credit(ctx context.Context, accountID uuid.UUID, amountCents int64) (int64, error)
	RecordCommandResult(ctx context.Context, key, resultEventType string, balanceCents *int64, failureReason *string, appliedAt time.Time) error
	WriteAudit(ctx context.Context, ev models.AuditEvent) error
}

// EventPublisher publishes balance movement results.
type EventPublisher interface {
	PublishAccountEvent(ctx context.Context, ev models.AccountEvent) error
}

// AccountConsumer applies balance movements requested by the Transfer Service
// saga. Commands are delivered at-least-once, so each command is claimed in
// processed_commands and its result is re-emitted on redelivery.
type AccountConsumer struct {
	reader   *kafka.Reader
	dlq      *kafka.Writer
	repo     CommandStore
	producer EventPublisher
	logger   *slog.Logger
}

func NewAccountConsumer(brokers []string, topic, group, dlqTopic string, repo CommandStore, prod EventPublisher, logger *slog.Logger) *AccountConsumer {
	return &AccountConsumer{
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

func (c *AccountConsumer) Run(ctx context.Context) error {
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
			c.logger.Error("handle account command", "err", err, "partition", msg.Partition, "offset", msg.Offset)
			continue // do not commit; retry
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit message", "err", err)
		}
	}
}

func (c *AccountConsumer) handle(ctx context.Context, msg kafka.Message) error {
	var cmd models.AccountCommand
	if err := json.Unmarshal(msg.Value, &cmd); err != nil {
		c.logger.Error("unmarshal account command", "err", err)
		c.sendToDLQ(ctx, msg, "malformed payload")
		return nil
	}
	switch cmd.EventType {
	case models.EventTypeAccountDebitRequested, models.EventTypeAccountCreditRequested:
	default:
		return nil // not an account command
	}
	if cmd.EventVersion != models.EventVersionV1 {
		c.sendToDLQ(ctx, msg, fmt.Sprintf("unsupported event version %d", cmd.EventVersion))
		return nil
	}
	transferID, err := uuid.Parse(cmd.TransferID)
	if err != nil {
		c.sendToDLQ(ctx, msg, "invalid transfer id")
		return nil
	}
	accountID, err := uuid.Parse(cmd.AccountID)
	if err != nil {
		c.sendToDLQ(ctx, msg, "invalid account id")
		return nil
	}

	applied, err := c.repo.ClaimCommand(ctx, cmd.IdempotencyKey, transferID, accountID, cmd.EventType, time.Now().UTC())
	if err != nil {
		return err
	}
	if !applied {
		existing, err := c.repo.FindCommandResult(ctx, cmd.IdempotencyKey)
		if err != nil {
			return err
		}
		if existing != nil && existing.ResultEventType != "" {
			return c.emitResult(ctx, cmd, existing.ResultEventType, existing.BalanceCents, existing.FailureReason)
		}
	}
	return c.applyAndEmit(ctx, cmd, accountID)
}

func (c *AccountConsumer) applyAndEmit(ctx context.Context, cmd models.AccountCommand, accountID uuid.UUID) error {
	var (
		balanceCents *int64
		resultType   string
		failure      string
	)
	if cmd.EventType == models.EventTypeAccountDebitRequested {
		balance, err := c.repo.Debit(ctx, accountID, cmd.AmountCents)
		if err != nil {
			resultType, failure = models.EventTypeAccountDebitFailed, reasonOf(err)
		} else {
			resultType, balanceCents = models.EventTypeAccountDebited, &balance
		}
	} else {
		balance, err := c.repo.Credit(ctx, accountID, cmd.AmountCents)
		if err != nil {
			resultType, failure = models.EventTypeAccountCreditFailed, reasonOf(err)
		} else {
			resultType, balanceCents = models.EventTypeAccountCredited, &balance
		}
	}

	var failureReason *string
	if failure != "" {
		failureReason = &failure
	}
	appliedAt := time.Now().UTC()
	if err := c.repo.RecordCommandResult(ctx, cmd.IdempotencyKey, resultType, balanceCents, failureReason, appliedAt); err != nil {
		return err
	}
	c.writeAudit(ctx, cmd, resultType, accountID, appliedAt)
	return c.emitResult(ctx, cmd, resultType, balanceCents, failureReason)
}

func (c *AccountConsumer) emitResult(ctx context.Context, cmd models.AccountCommand, resultType string, balanceCents *int64, failureReason *string) error {
	ev := models.AccountEvent{
		EventID:               newUUID().String(),
		EventType:             resultType,
		EventVersion:          models.EventVersionV1,
		TransferID:            cmd.TransferID,
		IdempotencyKey:        cmd.IdempotencyKey,
		CorrelationID:         cmd.CorrelationID,
		AccountID:             cmd.AccountID,
		CounterpartyAccountID: cmd.CounterpartyAccountID,
		AmountCents:           cmd.AmountCents,
		OccurredAt:            time.Now().UTC().Format(time.RFC3339),
	}
	if balanceCents != nil {
		ev.BalanceCents = *balanceCents
	}
	if failureReason != nil {
		ev.FailureReason = *failureReason
	}
	return c.producer.PublishAccountEvent(ctx, ev)
}

func (c *AccountConsumer) writeAudit(ctx context.Context, cmd models.AccountCommand, resultType string, accountID uuid.UUID, at time.Time) {
	action := "debited"
	if resultType == models.EventTypeAccountCredited || resultType == models.EventTypeAccountCreditFailed {
		action = "credited"
	}
	ev := models.AuditEvent{
		Day:           at,
		Timestamp:     at,
		EventID:       newUUID(),
		Action:        action,
		ResourceID:    accountID,
		CorrelationID: cmd.CorrelationID,
		Metadata: map[string]string{
			"transferId": cmd.TransferID,
			"result":     resultType,
		},
	}
	if err := c.repo.WriteAudit(ctx, ev); err != nil {
		c.logger.Error("write audit event", "err", err, "transferId", cmd.TransferID)
	}
}

func (c *AccountConsumer) sendToDLQ(ctx context.Context, msg kafka.Message, reason string) {
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

func (c *AccountConsumer) Close() error {
	_ = c.dlq.Close()
	return c.reader.Close()
}

func reasonOf(err error) string {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return "internal error"
}

func newUUID() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}
