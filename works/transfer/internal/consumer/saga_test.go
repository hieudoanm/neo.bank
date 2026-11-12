package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/works/transfer/internal/models"
	"github.com/segmentio/kafka-go"
)

type transitionCall struct {
	from, to string
}

type fakeStore struct {
	transfer      *models.Transfer
	findErr       error
	transitionErr error
	transitions   []transitionCall
	apply         bool
}

func (f *fakeStore) WriteAudit(_ context.Context, _ models.AuditEvent) error { return nil }

func (f *fakeStore) FindByID(_ context.Context, _ uuid.UUID) (*models.Transfer, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.transfer, nil
}

func (f *fakeStore) Transition(_ context.Context, t *models.Transfer, from, to string, _ *string) (bool, error) {
	f.transitions = append(f.transitions, transitionCall{from: from, to: to})
	if f.transitionErr != nil {
		return false, f.transitionErr
	}
	if f.apply {
		t.Status = to
	}
	return f.apply, nil
}

type fakePublisher struct {
	commands   []models.AccountCommand
	events     []models.TransferEvent
	commandErr error
	eventErr   error
}

func (f *fakePublisher) PublishAccountCommand(_ context.Context, cmd models.AccountCommand) error {
	if f.commandErr != nil {
		return f.commandErr
	}
	f.commands = append(f.commands, cmd)
	return nil
}

func (f *fakePublisher) PublishTransferEvent(_ context.Context, ev models.TransferEvent) error {
	if f.eventErr != nil {
		return f.eventErr
	}
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

func newConsumerWithPub(store SagaStore, pub *fakePublisher) *SagaConsumer {
	return &SagaConsumer{
		repo:     store,
		producer: pub,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func accountMessage(t *testing.T, ev models.AccountEvent) kafka.Message {
	t.Helper()
	return message(t, ev)
}

func TestSagaMalformedPayloadDropped(t *testing.T) {
	c, pub := newConsumer(&fakeStore{})
	if err := c.handle(context.Background(), kafka.Message{Value: []byte("not json")}); err != nil {
		t.Fatalf("malformed payload must not error: %v", err)
	}
	if len(pub.commands)+len(pub.events) != 0 {
		t.Fatal("malformed payload must not produce output")
	}
}

func TestSagaUnknownEventTypeIgnored(t *testing.T) {
	c, pub := newConsumer(&fakeStore{apply: true})
	msg := message(t, models.TransferEvent{EventType: "something.else", EventVersion: models.EventVersionV1})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands)+len(pub.events) != 0 {
		t.Fatal("unknown event type must be ignored")
	}
}

func TestSagaInitiatedFindErrors(t *testing.T) {
	tr := sampleTransfer(models.StatusPending)
	t.Run("not found", func(t *testing.T) {
		c, pub := newConsumer(&fakeStore{transfer: nil})
		msg := message(t, models.TransferEvent{EventType: models.EventTypeTransferInitiated, EventVersion: models.EventVersionV1, TransferID: tr.ID.String()})
		if err := c.handle(context.Background(), msg); err != nil {
			t.Fatalf("handle: %v", err)
		}
		if len(pub.commands) != 0 {
			t.Fatal("missing transfer must not start the saga")
		}
	})
	t.Run("find error", func(t *testing.T) {
		c, _ := newConsumer(&fakeStore{findErr: errors.New("find")})
		msg := message(t, models.TransferEvent{EventType: models.EventTypeTransferInitiated, EventVersion: models.EventVersionV1, TransferID: tr.ID.String()})
		if err := c.handle(context.Background(), msg); err == nil {
			t.Fatal("expected find error to propagate")
		}
	})
	t.Run("transition error", func(t *testing.T) {
		c, _ := newConsumer(&fakeStore{transfer: tr, transitionErr: errors.New("transition")})
		msg := message(t, models.TransferEvent{EventType: models.EventTypeTransferInitiated, EventVersion: models.EventVersionV1, TransferID: tr.ID.String()})
		if err := c.handle(context.Background(), msg); err == nil {
			t.Fatal("expected transition error to propagate")
		}
	})
}

func TestSagaDebitedIgnoresNonProcessing(t *testing.T) {
	tr := sampleTransfer(models.StatusPending)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := accountMessage(t, models.AccountEvent{EventType: models.EventTypeAccountDebited, EventVersion: models.EventVersionV1, TransferID: tr.ID.String(), AccountID: tr.SourceAccountID.String()})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 0 {
		t.Fatal("credit must only be issued while processing")
	}
}

func TestSagaDebitedIgnoresForeignAccount(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := accountMessage(t, models.AccountEvent{EventType: models.EventTypeAccountDebited, EventVersion: models.EventVersionV1, TransferID: tr.ID.String(), AccountID: uuid.NewString()})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 0 {
		t.Fatal("a debit result for a different account must not trigger the credit command")
	}
}

func TestSagaDebitedUnmarshalError(t *testing.T) {
	c, _ := newConsumer(&fakeStore{})
	if err := c.handle(context.Background(), kafka.Message{Value: []byte("not json")}); err != nil {
		t.Fatalf("handle: %v", err)
	}
}

func TestSagaDebitFailedDefaultsReason(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := accountMessage(t, models.AccountEvent{EventType: models.EventTypeAccountDebitFailed, EventVersion: models.EventVersionV1, TransferID: tr.ID.String(), AccountID: tr.SourceAccountID.String()})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0].EventType != models.EventTypeTransferFailed {
		t.Fatalf("expected transfer_failed event")
	}
}

func TestSagaCreditedDuplicateDoesNotRepublish(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: false})
	msg := accountMessage(t, models.AccountEvent{EventType: models.EventTypeAccountCredited, EventVersion: models.EventVersionV1, TransferID: tr.ID.String(), AccountID: tr.DestAccountID.String()})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.events) != 0 {
		t.Fatal("a credit whose transition did not apply must not emit another event")
	}
}

func TestSagaCreditFailedIgnoresCompensationToSource(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := accountMessage(t, models.AccountEvent{EventType: models.EventTypeAccountCreditFailed, EventVersion: models.EventVersionV1, TransferID: tr.ID.String(), AccountID: tr.SourceAccountID.String()})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 0 || len(pub.events) != 0 {
		t.Fatal("a failed compensation credit must not trigger another compensation")
	}
}

func TestSagaCreditFailedIgnoresNonProcessing(t *testing.T) {
	tr := sampleTransfer(models.StatusCompleted)
	c, pub := newConsumer(&fakeStore{transfer: tr, apply: true})
	msg := accountMessage(t, models.AccountEvent{EventType: models.EventTypeAccountCreditFailed, EventVersion: models.EventVersionV1, TransferID: tr.ID.String(), AccountID: tr.DestAccountID.String()})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.commands) != 0 {
		t.Fatal("credit failure must not compensate a non-processing transfer")
	}
}

func TestSagaPublisherErrorsPropagate(t *testing.T) {
	tr := sampleTransfer(models.StatusProcessing)
	t.Run("initiated command error", func(t *testing.T) {
		c := newConsumerWithPub(&fakeStore{transfer: tr, apply: true}, &fakePublisher{commandErr: errors.New("pub")})
		msg := message(t, models.TransferEvent{EventType: models.EventTypeTransferInitiated, EventVersion: models.EventVersionV1, TransferID: tr.ID.String()})
		if err := c.handle(context.Background(), msg); err == nil {
			t.Fatal("expected publish error")
		}
	})
	t.Run("completed event error", func(t *testing.T) {
		c := newConsumerWithPub(&fakeStore{transfer: tr, apply: true}, &fakePublisher{eventErr: errors.New("pub")})
		msg := accountMessage(t, models.AccountEvent{EventType: models.EventTypeAccountCredited, EventVersion: models.EventVersionV1, TransferID: tr.ID.String(), AccountID: tr.DestAccountID.String()})
		if err := c.handle(context.Background(), msg); err == nil {
			t.Fatal("expected publish error")
		}
	})
}

func TestSagaLoadTransferRejectsInvalidID(t *testing.T) {
	c, _ := newConsumer(&fakeStore{})
	if err := c.handle(context.Background(), kafka.Message{Value: []byte(`{"eventType":"transfer_initiated","eventVersion":1,"transferId":"nope"}`)}); err == nil {
		t.Fatal("expected invalid transfer id to error")
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
