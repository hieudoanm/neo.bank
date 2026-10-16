package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
	"github.com/segmentio/kafka-go"
)

// messageWriter is the subset of *kafka.Writer the Producer needs.
type messageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Producer struct {
	writer messageWriter
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: false,
		},
	}
}

// PublishTransferEvent publishes a transfer lifecycle event. The message key is
// the source account ID so all events for one account remain ordered. Callers
// must treat a returned error as fatal for the current operation.
func (p *Producer) PublishTransferEvent(ctx context.Context, ev models.TransferEvent) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal transfer event: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(ev.SourceAccountID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(ev.CorrelationID)},
			{Key: "event_type", Value: []byte(ev.EventType)},
		},
	})
}

// PublishAccountCommand publishes a saga command to the Account Service. The
// message key is the affected account ID so movements for one account stay
// ordered.
func (p *Producer) PublishAccountCommand(ctx context.Context, cmd models.AccountCommand) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("marshal account command: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(cmd.AccountID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(cmd.CorrelationID)},
			{Key: "event_type", Value: []byte(cmd.EventType)},
		},
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
