package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/hieudoanm/neo.bank/works/accounts/internal/models"
	"github.com/segmentio/kafka-go"
)

type fakeWriter struct {
	messages []kafka.Message
	err      error
	closed   bool
}

func (f *fakeWriter) WriteMessages(_ context.Context, msgs ...kafka.Message) error {
	if f.err != nil {
		return f.err
	}
	f.messages = append(f.messages, msgs...)
	return nil
}

func (f *fakeWriter) Close() error {
	f.closed = true
	return nil
}

func TestNewProducerConfiguresWriter(t *testing.T) {
	p := NewProducer([]string{"localhost:9092"}, "transfers")
	if p == nil || p.writer == nil {
		t.Fatal("expected a configured producer")
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestPublishAccountEvent(t *testing.T) {
	w := &fakeWriter{}
	p := &Producer{writer: w}
	ev := models.AccountEvent{AccountID: "account-1", EventType: "account.debited", CorrelationID: "corr-1"}
	if err := p.PublishAccountEvent(context.Background(), ev); err != nil {
		t.Fatalf("PublishAccountEvent: %v", err)
	}
	if len(w.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(w.messages))
	}
	msg := w.messages[0]
	if string(msg.Key) != "account-1" {
		t.Fatalf("unexpected key: %s", msg.Key)
	}
	if len(msg.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(msg.Headers))
	}
}

func TestPublishAccountEventPropagatesError(t *testing.T) {
	p := &Producer{writer: &fakeWriter{err: errors.New("write failed")}}
	if err := p.PublishAccountEvent(context.Background(), models.AccountEvent{}); err == nil {
		t.Fatal("expected write error")
	}
}

func TestPublishAccountEventMarshalError(t *testing.T) {
	p := &Producer{writer: &fakeWriter{}}
	if err := p.PublishAccountEvent(context.Background(), models.AccountEvent{}); err != nil {
		t.Fatalf("expected empty event to marshal, got %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
