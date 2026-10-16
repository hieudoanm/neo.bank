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
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/models"
	"github.com/segmentio/kafka-go"
)

type fakeStore struct {
	movements map[string]*models.TransferMovement
	inserted  []models.Transaction
	statuses  []string
	audits    int

	findErr   error
	insertErr error
	upsertErr error
	listErr   error
	updateErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{movements: map[string]*models.TransferMovement{}}
}

func key(transferID uuid.UUID, movement string) string {
	return transferID.String() + "/" + movement
}

func (f *fakeStore) FindMovement(_ context.Context, transferID uuid.UUID, movement string) (*models.TransferMovement, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.movements[key(transferID, movement)], nil
}

func (f *fakeStore) UpsertMovement(_ context.Context, m models.TransferMovement) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.movements[key(m.TransferID, m.Movement)] = &m
	return nil
}

func (f *fakeStore) InsertTransaction(_ context.Context, t models.Transaction) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.inserted = append(f.inserted, t)
	return nil
}

func (f *fakeStore) UpdateStatus(_ context.Context, _ uuid.UUID, _ time.Time, _ uuid.UUID, status string) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.statuses = append(f.statuses, status)
	return nil
}

func (f *fakeStore) ListMovements(_ context.Context, transferID uuid.UUID) ([]models.TransferMovement, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	var out []models.TransferMovement
	for _, m := range f.movements {
		if m.TransferID == transferID {
			out = append(out, *m)
		}
	}
	return out, nil
}

func (f *fakeStore) WriteAudit(_ context.Context, _ models.AuditEvent) error {
	f.audits++
	return nil
}

func newConsumer(store ProjectionStore) *ProjectionConsumer {
	return &ProjectionConsumer{
		store:  store,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func accountEventMsg(t *testing.T, ev any) kafka.Message {
	t.Helper()
	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return kafka.Message{Value: data}
}

func TestProjectDebitInsertsRow(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	transferID := uuid.New()
	accountID := uuid.New()
	counterparty := uuid.New()

	msg := accountEventMsg(t, models.AccountEvent{
		EventType:             models.EventTypeAccountDebited,
		EventVersion:          models.EventVersionV1,
		TransferID:            transferID.String(),
		IdempotencyKey:        transferID.String() + ":debit",
		AccountID:             accountID.String(),
		CounterpartyAccountID: counterparty.String(),
		AmountCents:           250,
		OccurredAt:            time.Now().UTC().Format(time.RFC3339),
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(store.inserted) != 1 {
		t.Fatalf("expected 1 inserted row, got %d", len(store.inserted))
	}
	row := store.inserted[0]
	if row.Type != models.TypeDebit || row.Status != models.StatusCompleted {
		t.Fatalf("unexpected row type/status: %s/%s", row.Type, row.Status)
	}
	if row.AccountID != accountID || row.CounterpartyAccountID != counterparty {
		t.Fatalf("row accounts mismatch")
	}
}

func TestProjectCreditUsesCreditType(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	transferID := uuid.New()
	msg := accountEventMsg(t, models.AccountEvent{
		EventType:      models.EventTypeAccountCredited,
		EventVersion:   models.EventVersionV1,
		TransferID:     transferID.String(),
		IdempotencyKey: transferID.String() + ":credit",
		AccountID:      uuid.NewString(),
		AmountCents:    100,
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(store.inserted) != 1 || store.inserted[0].Type != models.TypeCredit {
		t.Fatalf("expected a single credit row")
	}
}

func TestProjectIsIdempotent(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	transferID := uuid.New()
	msg := accountEventMsg(t, models.AccountEvent{
		EventType:      models.EventTypeAccountDebited,
		EventVersion:   models.EventVersionV1,
		TransferID:     transferID.String(),
		IdempotencyKey: transferID.String() + ":debit",
		AccountID:      uuid.NewString(),
		AmountCents:    100,
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle (redelivery): %v", err)
	}
	if len(store.inserted) != 1 {
		t.Fatalf("duplicate delivery must not insert twice, got %d", len(store.inserted))
	}
}

func TestCompensationCreditIsDistinctMovement(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	transferID := uuid.New()
	base := models.AccountEvent{
		EventVersion: models.EventVersionV1,
		TransferID:   transferID.String(),
		AccountID:    uuid.NewString(),
		AmountCents:  100,
	}
	credit := base
	credit.EventType = models.EventTypeAccountCredited
	credit.IdempotencyKey = transferID.String() + ":credit"
	comp := base
	comp.EventType = models.EventTypeAccountCredited
	comp.IdempotencyKey = transferID.String() + ":compensation"

	if err := c.handle(context.Background(), accountEventMsg(t, credit)); err != nil {
		t.Fatalf("handle credit: %v", err)
	}
	if err := c.handle(context.Background(), accountEventMsg(t, comp)); err != nil {
		t.Fatalf("handle compensation: %v", err)
	}
	if len(store.inserted) != 2 {
		t.Fatalf("expected 2 distinct rows, got %d", len(store.inserted))
	}
}

func TestFinalizeSetsStatusOnMovements(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	transferID := uuid.New()
	movementMsg := accountEventMsg(t, models.AccountEvent{
		EventType:      models.EventTypeAccountDebited,
		EventVersion:   models.EventVersionV1,
		TransferID:     transferID.String(),
		IdempotencyKey: transferID.String() + ":debit",
		AccountID:      uuid.NewString(),
		AmountCents:    100,
	})
	if err := c.handle(context.Background(), movementMsg); err != nil {
		t.Fatalf("handle movement: %v", err)
	}

	failMsg := accountEventMsg(t, models.TransferEvent{
		EventType:    models.EventTypeTransferFailed,
		EventVersion: models.EventVersionV1,
		TransferID:   transferID.String(),
	})
	if err := c.handle(context.Background(), failMsg); err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if len(store.statuses) != 1 || store.statuses[0] != models.StatusFailed {
		t.Fatalf("expected movement set to failed, got %v", store.statuses)
	}
}

func TestUnsupportedVersionRoutedToDLQ(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	msg := accountEventMsg(t, models.AccountEvent{
		EventType:    models.EventTypeAccountDebited,
		EventVersion: 99,
		TransferID:   uuid.NewString(),
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(store.inserted) != 0 {
		t.Fatalf("unsupported version must not project")
	}
}

func TestMalformedPayloadIgnored(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	if err := c.handle(context.Background(), kafka.Message{Value: []byte("not json")}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(store.inserted) != 0 {
		t.Fatalf("malformed payload must not project")
	}
}

func TestIgnoreNonProjectionEvents(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	msg := accountEventMsg(t, models.TransferEvent{
		EventType:    models.EventTypeTransferInitiated,
		EventVersion: models.EventVersionV1,
		TransferID:   uuid.NewString(),
	})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(store.inserted) != 0 || len(store.statuses) != 0 {
		t.Fatalf("transfer_initiated must not project")
	}
}

func TestProjectRejectsInvalidIDs(t *testing.T) {
	t.Run("invalid transfer id", func(t *testing.T) {
		store := newFakeStore()
		c := newConsumer(store)
		msg := accountEventMsg(t, models.AccountEvent{EventType: models.EventTypeAccountDebited, EventVersion: models.EventVersionV1, TransferID: "nope", AccountID: uuid.NewString(), AmountCents: 1})
		if err := c.handle(context.Background(), msg); err != nil {
			t.Fatalf("handle: %v", err)
		}
		if len(store.inserted) != 0 {
			t.Fatal("invalid transfer id must not project")
		}
	})
	t.Run("invalid account id", func(t *testing.T) {
		store := newFakeStore()
		c := newConsumer(store)
		msg := accountEventMsg(t, models.AccountEvent{EventType: models.EventTypeAccountDebited, EventVersion: models.EventVersionV1, TransferID: uuid.NewString(), AccountID: "nope", AmountCents: 1})
		if err := c.handle(context.Background(), msg); err != nil {
			t.Fatalf("handle: %v", err)
		}
		if len(store.inserted) != 0 {
			t.Fatal("invalid account id must not project")
		}
	})
}

func TestProjectStoreErrorsPropagate(t *testing.T) {
	base := models.AccountEvent{EventType: models.EventTypeAccountDebited, EventVersion: models.EventVersionV1, TransferID: uuid.NewString(), AccountID: uuid.NewString(), AmountCents: 1}
	t.Run("find error", func(t *testing.T) {
		c := newConsumer(&fakeStore{findErr: errors.New("find")})
		if err := c.handle(context.Background(), accountEventMsg(t, base)); err == nil {
			t.Fatal("expected find error")
		}
	})
	t.Run("insert error", func(t *testing.T) {
		c := newConsumer(&fakeStore{insertErr: errors.New("insert")})
		if err := c.handle(context.Background(), accountEventMsg(t, base)); err == nil {
			t.Fatal("expected insert error")
		}
	})
	t.Run("upsert error", func(t *testing.T) {
		c := newConsumer(&fakeStore{upsertErr: errors.New("upsert")})
		if err := c.handle(context.Background(), accountEventMsg(t, base)); err == nil {
			t.Fatal("expected upsert error")
		}
	})
}

func TestProjectFallsBackWhenOccurredAtInvalid(t *testing.T) {
	store := newFakeStore()
	c := newConsumer(store)
	msg := accountEventMsg(t, models.AccountEvent{EventType: models.EventTypeAccountDebited, EventVersion: models.EventVersionV1, TransferID: uuid.NewString(), AccountID: uuid.NewString(), AmountCents: 1, OccurredAt: "not-a-timestamp"})
	if err := c.handle(context.Background(), msg); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(store.inserted) != 1 {
		t.Fatal("invalid occurredAt must still project using now()")
	}
	if store.inserted[0].CounterpartyAccountID != uuid.Nil {
		t.Fatal("a missing counterparty must project as uuid.Nil")
	}
}

func TestFinalizeErrors(t *testing.T) {
	t.Run("invalid transfer id", func(t *testing.T) {
		c := newConsumer(newFakeStore())
		msg := accountEventMsg(t, models.TransferEvent{EventType: models.EventTypeTransferFailed, EventVersion: models.EventVersionV1, TransferID: "nope"})
		if err := c.handle(context.Background(), msg); err != nil {
			t.Fatalf("handle: %v", err)
		}
	})
	t.Run("list error", func(t *testing.T) {
		c := newConsumer(&fakeStore{listErr: errors.New("list")})
		msg := accountEventMsg(t, models.TransferEvent{EventType: models.EventTypeTransferFailed, EventVersion: models.EventVersionV1, TransferID: uuid.NewString()})
		if err := c.handle(context.Background(), msg); err == nil {
			t.Fatal("expected list error")
		}
	})
	t.Run("update error", func(t *testing.T) {
		store := &fakeStore{movements: map[string]*models.TransferMovement{}}
		transferID := uuid.New()
		store.movements[key(transferID, "debit")] = &models.TransferMovement{TransferID: transferID, Movement: "debit"}
		c := newConsumer(store)
		store.updateErr = errors.New("update")
		msg := accountEventMsg(t, models.TransferEvent{EventType: models.EventTypeTransferFailed, EventVersion: models.EventVersionV1, TransferID: transferID.String()})
		if err := c.handle(context.Background(), msg); err == nil {
			t.Fatal("expected update error")
		}
	})
}

func TestMovementOfFallsBackToEventType(t *testing.T) {
	debit := movementOf(models.AccountEvent{EventType: models.EventTypeAccountDebited, IdempotencyKey: "no-colon"})
	if debit != models.MovementDebit {
		t.Fatalf("movementOf = %q, want debit", debit)
	}
	credit := movementOf(models.AccountEvent{EventType: models.EventTypeAccountCredited, IdempotencyKey: "no-colon"})
	if credit != models.MovementCredit {
		t.Fatalf("movementOf = %q, want credit", credit)
	}
	comp := movementOf(models.AccountEvent{EventType: models.EventTypeAccountCredited, IdempotencyKey: "tid:compensation"})
	if comp != "compensation" {
		t.Fatalf("movementOf = %q, want compensation", comp)
	}
}
