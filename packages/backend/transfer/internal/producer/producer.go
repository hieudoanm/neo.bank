package producer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type TransferEvent struct {
	TransferID           string `json:"transferId"`
	IdempotencyKey       string `json:"idempotencyKey"`
	SourceAccountID      string `json:"sourceAccountId"`
	DestinationAccountID string `json:"destinationAccountId"`
	AmountCents          int64  `json:"amountCents"`
	CurrencyCode         string `json:"currency"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		Async:        true,
		RequiredAcks: kafka.RequireOne,
	}
	return &Producer{writer: w}
}

func (p *Producer) PublishTransferInitiated(ctx context.Context, ev TransferEvent) {
	data, err := json.Marshal(ev)
	if err != nil {
		log.Printf("producer marshal: %v", err)
		return
	}
	err = p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(ev.TransferID), Value: data})
	if err != nil {
		log.Printf("producer publish: %v", err)
	}
}
