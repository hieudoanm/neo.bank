package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/models"
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

func TestPublishUserEvent(t *testing.T) {
	w := &fakeWriter{}
	p := &Producer{writer: w}
	ev := models.UserCreatedEvent{AccountID: "account-1", EventType: models.EventTypeUserCreated}
	if err := p.PublishUserEvent(context.Background(), ev); err != nil {
		t.Fatalf("PublishUserEvent: %v", err)
	}
	if len(w.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(w.messages))
	}
	msg := w.messages[0]
	if string(msg.Key) != "account-1" {
		t.Fatalf("unexpected key: %s", msg.Key)
	}
	if len(msg.Headers) != 1 || string(msg.Headers[0].Value) != models.EventTypeUserCreated {
		t.Fatalf("expected the user event header, got %+v", msg.Headers)
	}
}

func TestPublishUserEventPropagatesError(t *testing.T) {
	p := &Producer{writer: &fakeWriter{err: errors.New("write failed")}}
	if err := p.PublishUserEvent(context.Background(), models.UserCreatedEvent{}); err == nil {
		t.Fatal("expected write error")
	}
}
