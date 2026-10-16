package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/models"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
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

// PublishAccountEvent publishes the outcome of a balance movement. The message
// key is the affected account ID so movements for one account stay ordered.
func (p *Producer) PublishAccountEvent(ctx context.Context, ev models.AccountEvent) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal account event: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(ev.AccountID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(ev.CorrelationID)},
			{Key: "event_type", Value: []byte(ev.EventType)},
		},
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
