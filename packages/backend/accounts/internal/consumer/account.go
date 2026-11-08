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

// fetchBackoff pauses between fetch retries so a broker outage does not
// spin the consume loop.
const fetchBackoff = time.Second

// CommandStore is the persistence the account consumer needs.
// *repositories.AccountRepository satisfies it in production; tests provide a fake.
type CommandStore interface {
	CreateAccount(ctx context.Context, accountID, userID uuid.UUID, balanceCents int64, currencyCode string, createdAt time.Time) (bool, error)
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

	initialBalanceCents int64
	currencyCode        string
}

func NewAccountConsumer(brokers []string, topic, group, dlqTopic string, initialBalanceCents int64, currencyCode string, repo CommandStore, prod EventPublisher, logger *slog.Logger) *AccountConsumer {
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
		repo:                repo,
		producer:            prod,
		logger:              logger,
		initialBalanceCents: initialBalanceCents,
		currencyCode:        currencyCode,
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
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(fetchBackoff):
			}
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
	var env struct {
		EventType    string `json:"eventType"`
		EventVersion int    `json:"eventVersion"`
	}
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		c.logger.Error("unmarshal event envelope", "err", err)
		return c.sendToDLQ(ctx, msg, "malformed payload")
	}
	if env.EventType == models.EventTypeUserCreated {
		if env.EventVersion != models.EventVersionV1 {
			return c.sendToDLQ(ctx, msg, fmt.Sprintf("unsupported event version %d", env.EventVersion))
		}
		return c.provisionAccount(ctx, msg)
	}

	var cmd models.AccountCommand
	if err := json.Unmarshal(msg.Value, &cmd); err != nil {
		c.logger.Error("unmarshal account command", "err", err)
		return c.sendToDLQ(ctx, msg, "malformed payload")
	}
	switch cmd.EventType {
	case models.EventTypeAccountDebitRequested, models.EventTypeAccountCreditRequested:
	default:
		return nil // not an account command
	}
	if cmd.EventVersion != models.EventVersionV1 {
		return c.sendToDLQ(ctx, msg, fmt.Sprintf("unsupported event version %d", cmd.EventVersion))
	}
	transferID, err := uuid.Parse(cmd.TransferID)
	if err != nil {
		return c.sendToDLQ(ctx, msg, "invalid transfer id")
	}
	accountID, err := uuid.Parse(cmd.AccountID)
	if err != nil {
		return c.sendToDLQ(ctx, msg, "invalid account id")
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
		// The command was claimed but no result was recorded, so the balance
		// movement outcome is unknown (e.g. a crash between applying it and
		// recording the result). Re-applying could move funds twice; route the
		// message to the DLQ for reconciliation instead.
		return c.sendToDLQ(ctx, msg, "command claimed but result missing")
	}
	return c.applyAndEmit(ctx, cmd, accountID)
}

// provisionAccount creates the account row for a newly registered user. The
// write is idempotent, so redelivery of the same event is a no-op once the
// account exists.
func (c *AccountConsumer) provisionAccount(ctx context.Context, msg kafka.Message) error {
	var ev models.UserCreatedEvent
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		c.logger.Error("unmarshal user created event", "err", err)
		return c.sendToDLQ(ctx, msg, "malformed user event")
	}
	userID, err := uuid.Parse(ev.UserID)
	if err != nil {
		return c.sendToDLQ(ctx, msg, "invalid user id")
	}
	accountID, err := uuid.Parse(ev.AccountID)
	if err != nil {
		return c.sendToDLQ(ctx, msg, "invalid account id")
	}

	now := time.Now().UTC()
	created, err := c.repo.CreateAccount(ctx, accountID, userID, c.initialBalanceCents, c.currencyCode, now)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	c.logger.Info("account provisioned", "accountId", ev.AccountID, "userId", ev.UserID)
	c.writeProvisionAudit(ctx, ev, accountID, now)
	return nil
}

func (c *AccountConsumer) writeProvisionAudit(ctx context.Context, ev models.UserCreatedEvent, accountID uuid.UUID, at time.Time) {
	ae := models.AuditEvent{
		Day:           at,
		Timestamp:     at,
		EventID:       newUUID(),
		Action:        "account_created",
		ResourceID:    accountID,
		CorrelationID: ev.EventID,
		Metadata: map[string]string{
			"userId": ev.UserID,
		},
	}
	if err := c.repo.WriteAudit(ctx, ae); err != nil {
		c.logger.Error("write provisioning audit", "err", err, "accountId", accountID)
	}
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

func (c *AccountConsumer) sendToDLQ(ctx context.Context, msg kafka.Message, reason string) error {
	c.logger.Error("routing to dlq", "reason", reason, "partition", msg.Partition, "offset", msg.Offset)
	if c.dlq == nil {
		return nil
	}
	// Copy headers before appending: append may write into the fetched
	// message's backing array.
	headers := make([]kafka.Header, 0, len(msg.Headers)+1)
	headers = append(headers, msg.Headers...)
	headers = append(headers, kafka.Header{
		Key:   "failure_reason",
		Value: []byte(reason),
	})
	if err := c.dlq.WriteMessages(ctx, kafka.Message{
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
	}); err != nil {
		c.logger.Error("write dlq", "err", err)
		return fmt.Errorf("write dlq: %w", err)
	}
	return nil
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
