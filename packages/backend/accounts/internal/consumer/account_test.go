package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/models"
	"github.com/segmentio/kafka-go"
)

type fakeStore struct {
	claimed       bool
	claimErr      error
	existing      *models.ProcessedCommand
	findErr       error
	debitBalance  int64
	debitErr      error
	creditBalance int64
	creditErr     error
	recordErr     error

	debitCalls  int
	creditCalls int
	recorded    []string
	audits      int
}

func (f *fakeStore) ClaimCommand(_ context.Context, _ string, _, _ uuid.UUID, _ string, _ time.Time) (bool, error) {
	return f.claimed, f.claimErr
}

func (f *fakeStore) FindCommandResult(_ context.Context, _ string) (*models.ProcessedCommand, error) {
	return f.existing, f.findErr
}

func (f *fakeStore) Debit(_ context.Context, _ uuid.UUID, _ int64) (int64, error) {
	f.debitCalls++
	return f.debitBalance, f.debitErr
}

func (f *fakeStore) Credit(_ context.Context, _ uuid.UUID, _ int64) (int64, error) {
	f.creditCalls++
	return f.creditBalance, f.creditErr
}

func (f *fakeStore) RecordCommandResult(_ context.Context, _, resultEventType string, _ *int64, _ *string, _ time.Time) error {
	f.recorded = append(f.recorded, resultEventType)
	return f.recordErr
}

func (f *fakeStore) WriteAudit(_ context.Context, _ models.AuditEvent) error {
	f.audits++
	return nil
}

type fakePublisher struct {
	events []models.AccountEvent
	err    error
}

func (f *fakePublisher) PublishAccountEvent(_ context.Context, ev models.AccountEvent) error {
	if f.err != nil {
		return f.err
	}
	f.events = append(f.events, ev)
	return nil
}

func newConsumer(store CommandStore) (*AccountConsumer, *fakePublisher) {
	pub := &fakePublisher{}
	c := &AccountConsumer{
		repo:     store,
		producer: pub,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	return c, pub
}

func command(t *testing.T, eventType string) kafka.Message {
	t.Helper()
	data, err := json.Marshal(models.AccountCommand{
		EventType:      eventType,
		EventVersion:   models.EventVersionV1,
		TransferID:     uuid.NewString(),
		IdempotencyKey: uuid.NewString() + ":debit",
		CorrelationID:  uuid.NewString(),
		AccountID:      uuid.NewString(),
		AmountCents:    250,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return kafka.Message{Value: data}
}

func TestAccountDebitAppliesAndEmits(t *testing.T) {
	store := &fakeStore{claimed: true, debitBalance: 750}
	c, pub := newConsumer(store)
	if err := c.handle(context.Background(), command(t, models.EventTypeAccountDebitRequested)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.debitCalls != 1 {
		t.Fatalf("expected 1 debit, got %d", store.debitCalls)
	}
	if len(pub.events) != 1 || pub.events[0].EventType != models.EventTypeAccountDebited {
		t.Fatalf("expected account.debited event, got %+v", pub.events)
	}
	if pub.events[0].BalanceCents != 750 {
		t.Fatalf("expected balance 750, got %d", pub.events[0].BalanceCents)
	}
	if store.audits != 1 {
		t.Fatalf("expected 1 audit event, got %d", store.audits)
	}
}

func TestAccountDebitInsufficientFundsEmitsFailure(t *testing.T) {
	store := &fakeStore{claimed: true, debitErr: apperrors.ErrInsufficientFunds}
	c, pub := newConsumer(store)
	if err := c.handle(context.Background(), command(t, models.EventTypeAccountDebitRequested)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0].EventType != models.EventTypeAccountDebitFailed {
		t.Fatalf("expected account.debit.failed event")
	}
	if pub.events[0].FailureReason == "" {
		t.Fatalf("expected a failure reason")
	}
}

func TestAccountCreditAppliesAndEmits(t *testing.T) {
	store := &fakeStore{claimed: true, creditBalance: 1000}
	c, pub := newConsumer(store)
	if err := c.handle(context.Background(), command(t, models.EventTypeAccountCreditRequested)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.creditCalls != 1 {
		t.Fatalf("expected 1 credit, got %d", store.creditCalls)
	}
	if len(pub.events) != 1 || pub.events[0].EventType != models.EventTypeAccountCredited {
		t.Fatalf("expected account.credited event")
	}
}

func TestAccountDuplicateReemitsStoredResultWithoutMovingFunds(t *testing.T) {
	balance := int64(400)
	store := &fakeStore{
		claimed:  false,
		existing: &models.ProcessedCommand{ResultEventType: models.EventTypeAccountDebited, BalanceCents: &balance},
	}
	c, pub := newConsumer(store)
	if err := c.handle(context.Background(), command(t, models.EventTypeAccountDebitRequested)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.debitCalls != 0 {
		t.Fatalf("duplicate command must not move funds, got %d debit calls", store.debitCalls)
	}
	if len(pub.events) != 1 || pub.events[0].EventType != models.EventTypeAccountDebited {
		t.Fatalf("expected re-emitted account.debited event")
	}
}

func TestAccountIgnoresTransferLifecycleEvents(t *testing.T) {
	store := &fakeStore{claimed: true}
	c, pub := newConsumer(store)
	msg := kafka.Message{Value: []byte(`{"eventType":"transfer_initiated","eventVersion":1}`)}
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.debitCalls+store.creditCalls != 0 {
		t.Fatalf("transfer events must be ignored by the account consumer")
	}
	if len(pub.events) != 0 {
		t.Fatalf("no result event expected")
	}
}

func TestAccountMalformedPayloadIsDropped(t *testing.T) {
	store := &fakeStore{claimed: true}
	c, _ := newConsumer(store)
	msg := kafka.Message{Value: []byte("not json")}
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("malformed payload must not error so the partition advances: %v", err)
	}
	if store.debitCalls+store.creditCalls != 0 {
		t.Fatalf("malformed payload must not move funds")
	}
}

func TestAccountRejectsUnsupportedVersion(t *testing.T) {
	store := &fakeStore{claimed: true}
	c, pub := newConsumer(store)
	msg := kafka.Message{Value: []byte(`{"eventType":"account.debit.requested","eventVersion":99}`)}
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.debitCalls != 0 || len(pub.events) != 0 {
		t.Fatal("unsupported version must be dropped without moving funds")
	}
}

func TestAccountRejectsInvalidIDs(t *testing.T) {
	cases := map[string]string{
		"invalid transfer id": `{"eventType":"account.debit.requested","eventVersion":1,"transferId":"nope","accountId":"` + uuid.NewString() + `"}`,
		"invalid account id":  `{"eventType":"account.debit.requested","eventVersion":1,"transferId":"` + uuid.NewString() + `","accountId":"nope"}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			store := &fakeStore{claimed: true}
			c, pub := newConsumer(store)
			if err := c.handle(context.Background(), kafka.Message{Value: []byte(payload)}); err != nil {
				t.Fatalf("handle: %v", err)
			}
			if store.debitCalls+store.creditCalls != 0 || len(pub.events) != 0 {
				t.Fatal("invalid ids must not move funds or emit events")
			}
		})
	}
}

func TestAccountHandleErrors(t *testing.T) {
	t.Run("claim error", func(t *testing.T) {
		c, _ := newConsumer(&fakeStore{claimErr: errors.New("claim")})
		if err := c.handle(context.Background(), command(t, models.EventTypeAccountDebitRequested)); err == nil {
			t.Fatal("expected claim error to propagate")
		}
	})
	t.Run("duplicate lookup error", func(t *testing.T) {
		c, _ := newConsumer(&fakeStore{claimed: false, findErr: errors.New("find")})
		if err := c.handle(context.Background(), command(t, models.EventTypeAccountDebitRequested)); err == nil {
			t.Fatal("expected find error to propagate")
		}
	})
	t.Run("record error", func(t *testing.T) {
		c, _ := newConsumer(&fakeStore{claimed: true, recordErr: errors.New("record")})
		if err := c.handle(context.Background(), command(t, models.EventTypeAccountDebitRequested)); err == nil {
			t.Fatal("expected record error to propagate")
		}
	})
	t.Run("publish error", func(t *testing.T) {
		store := &fakeStore{claimed: true}
		c := &AccountConsumer{repo: store, producer: &fakePublisher{err: errors.New("publish")}, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
		if err := c.handle(context.Background(), command(t, models.EventTypeAccountCreditRequested)); err == nil {
			t.Fatal("expected publish error to propagate")
		}
	})
}

func TestAccountDuplicateWithoutStoredResultApplies(t *testing.T) {
	store := &fakeStore{claimed: false, debitBalance: 100}
	c, pub := newConsumer(store)
	if err := c.handle(context.Background(), command(t, models.EventTypeAccountDebitRequested)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.debitCalls != 1 {
		t.Fatalf("expected the movement to apply when no stored result exists, got %d", store.debitCalls)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected one emitted event")
	}
}

func TestReasonOf(t *testing.T) {
	if got := reasonOf(apperrors.ErrInsufficientFunds); got != apperrors.ErrInsufficientFunds.Message {
		t.Fatalf("reasonOf = %q", got)
	}
	if got := reasonOf(errors.New("boom")); got != "internal error" {
		t.Fatalf("reasonOf = %q", got)
	}
}
