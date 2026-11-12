package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/models"
	"github.com/segmentio/kafka-go"
)

// fetchBackoff pauses between fetch retries so a broker outage does not
// spin the consume loop.
const fetchBackoff = time.Second

// ProjectionStore is the persistence the read-model projection needs.
// *repositories.TransactionRepository satisfies it in production; tests provide a fake.
type ProjectionStore interface {
	FindMovement(ctx context.Context, transferID uuid.UUID, movement string) (*models.TransferMovement, error)
	UpsertMovement(ctx context.Context, m models.TransferMovement) error
	InsertTransaction(ctx context.Context, t models.Transaction) error
	UpdateStatus(ctx context.Context, accountID uuid.UUID, createdAt time.Time, transactionID uuid.UUID, status string) error
	ListMovements(ctx context.Context, transferID uuid.UUID) ([]models.TransferMovement, error)
	WriteAudit(ctx context.Context, ev models.AuditEvent) error
}

// ProjectionConsumer builds the append-only transaction read model from the
// transfer/account event stream. Account movement events create rows; transfer
// lifecycle events finalise their status. It is idempotent: a movement is keyed
// by (transfer_id, movement) so a redelivery never duplicates a row, and status
// updates are re-appliable.
type ProjectionConsumer struct {
	reader *kafka.Reader
	dlq    *kafka.Writer
	store  ProjectionStore
	logger *slog.Logger
}

func NewProjectionConsumer(brokers []string, topic, group, dlqTopic string, store ProjectionStore, logger *slog.Logger) *ProjectionConsumer {
	return &ProjectionConsumer{
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
		store:  store,
		logger: logger,
	}
}

func (c *ProjectionConsumer) Run(ctx context.Context) error {
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
			c.logger.Error("handle event", "err", err, "partition", msg.Partition, "offset", msg.Offset)
			continue // do not commit; retry
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit message", "err", err)
		}
	}
}

func (c *ProjectionConsumer) handle(ctx context.Context, msg kafka.Message) error {
	var env struct {
		EventType    string `json:"eventType"`
		EventVersion int    `json:"eventVersion"`
	}
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		return c.sendToDLQ(ctx, msg, "malformed payload")
	}
	if env.EventVersion != models.EventVersionV1 {
		return c.sendToDLQ(ctx, msg, fmt.Sprintf("unsupported event version %d", env.EventVersion))
	}

	switch env.EventType {
	case models.EventTypeAccountDebited, models.EventTypeAccountCredited:
		return c.projectMovement(ctx, msg)
	case models.EventTypeTransferCompleted:
		return c.finalize(ctx, msg, models.StatusCompleted, "transfer_completed")
	case models.EventTypeTransferFailed:
		return c.finalize(ctx, msg, models.StatusFailed, "transfer_failed")
	default:
		return nil // not a projection event (e.g. transfer_initiated, *_failed commands)
	}
}

func (c *ProjectionConsumer) projectMovement(ctx context.Context, msg kafka.Message) error {
	var ev models.AccountEvent
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		return err
	}
	transferID, err := uuid.Parse(ev.TransferID)
	if err != nil {
		return c.sendToDLQ(ctx, msg, "invalid transfer id")
	}
	accountID, err := uuid.Parse(ev.AccountID)
	if err != nil {
		return c.sendToDLQ(ctx, msg, "invalid account id")
	}

	movement := movementOf(ev)
	existing, err := c.store.FindMovement(ctx, transferID, movement)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil // already projected
	}

	txType := models.TypeDebit
	if ev.EventType == models.EventTypeAccountCredited {
		txType = models.TypeCredit
	}
	createdAt := nowUTC()
	if ts, err := time.Parse(time.RFC3339, ev.OccurredAt); err == nil {
		createdAt = ts.UTC()
	}
	transactionID := newUUID()

	t := models.Transaction{
		AccountID:             accountID,
		CreatedAt:             createdAt,
		TransactionID:         transactionID,
		TransferID:            &transferID,
		Type:                  txType,
		AmountCents:           ev.AmountCents,
		CounterpartyAccountID: parseUUIDOrNil(ev.CounterpartyAccountID),
		Status:                models.StatusCompleted,
	}
	if err := c.store.InsertTransaction(ctx, t); err != nil {
		return err
	}
	if err := c.store.UpsertMovement(ctx, models.TransferMovement{
		TransferID:    transferID,
		Movement:      movement,
		TransactionID: transactionID,
		AccountID:     accountID,
		CreatedAt:     createdAt,
	}); err != nil {
		return err
	}
	c.writeAudit(ctx, "transaction_projected", transactionID, ev.CorrelationID, map[string]string{
		"transferId": ev.TransferID,
		"type":       txType,
	})
	return nil
}

func (c *ProjectionConsumer) finalize(ctx context.Context, msg kafka.Message, status, action string) error {
	var ev models.TransferEvent
	if err := json.Unmarshal(msg.Value, &ev); err != nil {
		return err
	}
	transferID, err := uuid.Parse(ev.TransferID)
	if err != nil {
		return c.sendToDLQ(ctx, msg, "invalid transfer id")
	}
	movements, err := c.store.ListMovements(ctx, transferID)
	if err != nil {
		return err
	}
	for _, m := range movements {
		if err := c.store.UpdateStatus(ctx, m.AccountID, m.CreatedAt, m.TransactionID, status); err != nil {
			return err
		}
	}
	if len(movements) > 0 {
		c.writeAudit(ctx, "status_updated", transferID, ev.CorrelationID, map[string]string{
			"transferId": ev.TransferID,
			"status":     status,
		})
	}
	return nil
}

func (c *ProjectionConsumer) writeAudit(ctx context.Context, action string, resourceID uuid.UUID, correlationID string, metadata map[string]string) {
	at := nowUTC()
	if err := c.store.WriteAudit(ctx, models.AuditEvent{
		Day:           at,
		Timestamp:     at,
		EventID:       newUUID(),
		Action:        action,
		ResourceID:    resourceID,
		CorrelationID: correlationID,
		Metadata:      metadata,
	}); err != nil {
		c.logger.Error("write audit event", "err", err, "action", action)
	}
}

func (c *ProjectionConsumer) sendToDLQ(ctx context.Context, msg kafka.Message, reason string) error {
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

func (c *ProjectionConsumer) Close() error {
	_ = c.dlq.Close()
	return c.reader.Close()
}

// movementOf derives the movement identifier from the command idempotency key
// suffix ("<transferId>:<movement>"), falling back to the event type.
func movementOf(ev models.AccountEvent) string {
	if i := strings.LastIndex(ev.IdempotencyKey, ":"); i >= 0 && i+1 < len(ev.IdempotencyKey) {
		return ev.IdempotencyKey[i+1:]
	}
	if ev.EventType == models.EventTypeAccountDebited {
		return models.MovementDebit
	}
	return models.MovementCredit
}

func parseUUIDOrNil(s string) uuid.UUID {
	if id, err := uuid.Parse(s); err == nil {
		return id
	}
	return uuid.Nil
}

func nowUTC() time.Time { return time.Now().UTC() }

func newUUID() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}
