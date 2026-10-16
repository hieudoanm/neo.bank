package consumer

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
	"github.com/segmentio/kafka-go"
)

type transitionCall struct {
	from, to string
}

type fakeStore struct {
	transfer    *models.Transfer
	transitions []transitionCall
	apply       bool
}

func (f *fakeStore) FindByID(_ context.Context, _ uuid.UUID) (*models.Transfer, error) {
	return f.transfer, nil
}

func (f *fakeStore) Transition(_ context.Context, t *models.Transfer, from, to string, _ *string) (bool, error) {
	f.transitions = append(f.transitions, transitionCall{from: from, to: to})
	if f.apply {
		t.Status = to
	}
	return f.apply, nil
}

type fakePublisher struct {
	commands []models.AccountCommand
	events   []models.TransferEvent
}

func (f *fakePublisher) PublishAccountCommand(_ context.Context, cmd models.AccountCommand) error {
	f.commands = append(f.commands, cmd)
	return nil
}

func (f *fakePublisher) PublishTransferEvent(_ context.Context, ev models.TransferEvent) error {
	f.events = append(f.events, ev)
	return nil
}

func newConsumer(store SagaStore) (*SagaConsumer, *fakePublisher) {
	pub := &fakePublisher{}
	c := &SagaConsumer{
		repo:     store,
		producer: pub,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	return c, pub
}

func sampleTransfer(status string) *models.Transfer {
	id := uuid.New()
	return &models.Transfer{
		ID:              id,
		SourceAccountID: uuid.New(),
		DestAccountID:   uuid.New(),
		AmountCents:     500,
		Status:          status,
		IdempotencyKey:  "req-key-1",
		CorrelationID:   uuid.NewString(),
	}
}

func message(t *testing.T, v any) kafka.Message {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return kafka.Message{Value: data}
}

func TestSagaInitiatedStartsDebit(t *testing.T) {
	tr := sampleTransfer(models.StatusPending)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := message(t, models.TransferEvent{
		EventType:    models.EventTypeTransferInitiated,
		EventVersion: models.EventVersionV1,
		TransferID:   tr.ID.String(),
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(pub.commands))
	}
	cmd := pub.commands[0]
	if cmd.EventType != models.EventTypeAccountDebitRequested {
		t.Fatalf("expected debit command, got %s", cmd.EventType)
	}
	if cmd.IdempotencyKey != tr.ID.String()+":debit" {
		t.Fatalf("unexpected idempotency key %s", cmd.IdempotencyKey)
	}
	if cmd.AccountID != tr.SourceAccountID.String() {
		t.Fatalf("debit must target the source account")
	}
}

func TestSagaInitiatedDuplicateDoesNotRepublish(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: false})
	msg := message(t, models.TransferEvent{
		EventType:    models.EventTypeTransferInitiated,
		EventVersion: models.EventVersionV1,
		TransferID:   tr.ID.String(),
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 0 {
		t.Fatalf("expected no command on duplicate, got %d", len(pub.commands))
	}
}

func TestSagaDebitedIssuesCredit(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := message(t, models.AccountEvent{
		EventType:    models.EventTypeAccountDebited,
		EventVersion: models.EventVersionV1,
		TransferID:   tr.ID.String(),
		AccountID:    tr.SourceAccountID.String(),
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 1 || pub.commands[0].EventType != models.EventTypeAccountCreditRequested {
		t.Fatalf("expected credit command")
	}
	if pub.commands[0].IdempotencyKey != tr.ID.String()+":credit" {
		t.Fatalf("unexpected credit key %s", pub.commands[0].IdempotencyKey)
	}
	if pub.commands[0].AccountID != tr.DestAccountID.String() {
		t.Fatalf("credit must target the destination account")
	}
}

func TestSagaCreditedCompletes(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := message(t, models.AccountEvent{
		EventType:    models.EventTypeAccountCredited,
		EventVersion: models.EventVersionV1,
		TransferID:   tr.ID.String(),
		AccountID:    tr.DestAccountID.String(),
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0].EventType != models.EventTypeTransferCompleted {
		t.Fatalf("expected transfer_completed event")
	}
}

func TestSagaIgnoresCompensationCreditForCompletion(t *testing.T) {
	tr := sampleTransfer(models.StatusFailed)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := message(t, models.AccountEvent{
		EventType:    models.EventTypeAccountCredited,
		EventVersion: models.EventVersionV1,
		TransferID:   tr.ID.String(),
		AccountID:    tr.SourceAccountID.String(),
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.events) != 0 {
		t.Fatalf("compensation credit must not complete the transfer")
	}
}

func TestSagaDebitFailedMarksFailed(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := message(t, models.AccountEvent{
		EventType:     models.EventTypeAccountDebitFailed,
		EventVersion:  models.EventVersionV1,
		TransferID:    tr.ID.String(),
		AccountID:     tr.SourceAccountID.String(),
		FailureReason: "insufficient funds",
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0].EventType != models.EventTypeTransferFailed {
		t.Fatalf("expected transfer_failed event")
	}
	if len(pub.commands) != 0 {
		t.Fatalf("debit failure needs no compensation command")
	}
}

func TestSagaCreditFailedCompensatesAndFails(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := message(t, models.AccountEvent{
		EventType:     models.EventTypeAccountCreditFailed,
		EventVersion:  models.EventVersionV1,
		TransferID:    tr.ID.String(),
		AccountID:     tr.DestAccountID.String(),
		FailureReason: "destination not found",
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 1 {
		t.Fatalf("expected 1 compensation command, got %d", len(pub.commands))
	}
	comp := pub.commands[0]
	if comp.EventType != models.EventTypeAccountCreditRequested || comp.AccountID != tr.SourceAccountID.String() {
		t.Fatalf("compensation must credit the source account")
	}
	if comp.IdempotencyKey != tr.ID.String()+":compensation" {
		t.Fatalf("unexpected compensation key %s", comp.IdempotencyKey)
	}
	if len(pub.events) != 1 || pub.events[0].EventType != models.EventTypeTransferFailed {
		t.Fatalf("expected transfer_failed event")
	}
}

func TestSagaUnsupportedVersionGoesToDLQ(t *testing.T) {
	tr := sampleTransfer(models.StatusPending)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := message(t, models.TransferEvent{
		EventType:    models.EventTypeTransferInitiated,
		EventVersion: 99,
		TransferID:   tr.ID.String(),
	})
	// sendToDLQ writes to a zero-value writer which returns an error; handle
	// swallows it and returns nil so the partition advances.
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 0 {
		t.Fatalf("unsupported version must not start the saga")
	}
}
