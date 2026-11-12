package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hieudoanm/neo.bank/works/auth/internal/models"
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

// PublishUserEvent publishes an auth lifecycle event. The message key is the
// account ID so all events for one account stay ordered. Callers must treat a
// returned error as fatal for the current operation.
func (p *Producer) PublishUserEvent(ctx context.Context, ev models.UserCreatedEvent) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal user event: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(ev.AccountID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(ev.EventType)},
		},
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
