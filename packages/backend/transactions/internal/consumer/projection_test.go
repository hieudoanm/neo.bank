package consumer

import (
	"context"
	"encoding/json"
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
}

func newFakeStore() *fakeStore {
	return &fakeStore{movements: map[string]*models.TransferMovement{}}
}

func key(transferID uuid.UUID, movement string) string {
	return transferID.String() + "/" + movement
}

func (f *fakeStore) FindMovement(_ context.Context, transferID uuid.UUID, movement string) (*models.TransferMovement, error) {
	return f.movements[key(transferID, movement)], nil
}

func (f *fakeStore) UpsertMovement(_ context.Context, m models.TransferMovement) error {
	f.movements[key(m.TransferID, m.Movement)] = &m
	return nil
}

func (f *fakeStore) InsertTransaction(_ context.Context, t models.Transaction) error {
	f.inserted = append(f.inserted, t)
	return nil
}

func (f *fakeStore) UpdateStatus(_ context.Context, _ uuid.UUID, _ time.Time, _ uuid.UUID, status string) error {
	f.statuses = append(f.statuses, status)
	return nil
}

func (f *fakeStore) ListMovements(_ context.Context, transferID uuid.UUID) ([]models.TransferMovement, error) {
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
