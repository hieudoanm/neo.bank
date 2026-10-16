package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
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

func TestNewProducer(t *testing.T) {
	p := NewProducer([]string{"b"}, "t")
	if p == nil {
		t.Fatal("nil producer")
	}
}

func TestPublishTransferEvent(t *testing.T) {
	w := &fakeWriter{}
	p := &Producer{writer: w}
	ev := models.TransferEvent{SourceAccountID: "s", EventType: "transfer_initiated", CorrelationID: "c"}
	if err := p.PublishTransferEvent(context.Background(), ev); err != nil {
		t.Fatalf("PublishTransferEvent: %v", err)
	}
	if len(w.messages) != 1 || string(w.messages[0].Key) != "s" {
		t.Fatalf("unexpected message: %+v", w.messages)
	}
}

func TestPublishAccountCommand(t *testing.T) {
	w := &fakeWriter{}
	p := &Producer{writer: w}
	cmd := models.AccountCommand{AccountID: "a", EventType: "account.debit.requested", CorrelationID: "c"}
	if err := p.PublishAccountCommand(context.Background(), cmd); err != nil {
		t.Fatalf("PublishAccountCommand: %v", err)
	}
	if len(w.messages) != 1 {
		t.Fatalf("unexpected message count")
	}
}

func TestProducerErrorsAndClose(t *testing.T) {
	p := &Producer{writer: &fakeWriter{err: errors.New("x")}}
	if err := p.PublishTransferEvent(context.Background(), models.TransferEvent{}); err == nil {
		t.Fatal("expected error")
	}
	if err := p.PublishAccountCommand(context.Background(), models.AccountCommand{}); err == nil {
		t.Fatal("expected error")
	}
	w := &fakeWriter{}
	p = &Producer{writer: w}
	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !w.closed {
		t.Fatal("writer should be closed")
	}
}
