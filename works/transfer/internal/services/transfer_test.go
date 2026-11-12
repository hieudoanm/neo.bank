package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	applierrors "github.com/hieudoanm/neo.bank/works/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/transfer/internal/models"
)

func TestInitiateTransferRejectsMissingIdempotencyKey(t *testing.T) {
	svc := NewTransferService(nil, nil)
	_, err := svc.InitiateTransfer(context.Background(), uuid.NewString(), models.CreateTransferRequest{
		DestinationAccountId: uuid.NewString(),
		AmountCents:          100,
	}, "")
	assertAppErrorCode(t, err, "VALIDATION_ERROR")
}

func TestInitiateTransferRejectsNonPositiveAmount(t *testing.T) {
	svc := NewTransferService(nil, nil)
	for _, amount := range []int64{0, -1} {
		_, err := svc.InitiateTransfer(context.Background(), uuid.NewString(), models.CreateTransferRequest{
			DestinationAccountId: uuid.NewString(),
			AmountCents:          amount,
		}, "key-1")
		assertAppErrorCode(t, err, "VALIDATION_ERROR")
	}
}

func TestInitiateTransferRejectsAmountOverMaximum(t *testing.T) {
	svc := NewTransferService(nil, nil)
	_, err := svc.InitiateTransfer(context.Background(), uuid.NewString(), models.CreateTransferRequest{
		DestinationAccountId: uuid.NewString(),
		AmountCents:          maxAmountCents + 1,
	}, "key-1")
	assertAppErrorCode(t, err, "VALIDATION_ERROR")
}

func TestInitiateTransferRejectsInvalidDestination(t *testing.T) {
	svc := NewTransferService(nil, nil)
	_, err := svc.InitiateTransfer(context.Background(), uuid.NewString(), models.CreateTransferRequest{
		DestinationAccountId: "not-a-uuid",
		AmountCents:          100,
	}, "key-1")
	assertAppErrorCode(t, err, "VALIDATION_ERROR")
}

func TestInitiateTransferRejectsSameAccount(t *testing.T) {
	svc := NewTransferService(nil, nil)
	account := uuid.NewString()
	_, err := svc.InitiateTransfer(context.Background(), account, models.CreateTransferRequest{
		DestinationAccountId: account,
		AmountCents:          100,
	}, "key-1")
	assertAppErrorCode(t, err, "VALIDATION_ERROR")
}

func assertAppErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	var appErr *applierrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %v", err)
	}
	if appErr.Code != code {
		t.Fatalf("expected code %s, got %s", code, appErr.Code)
	}
}

type fakeIdempotency struct {
	transferID uuid.UUID
	hash       string
}

// fakeStore is an in-memory TransferStore. It reproduces the repository's
// ReserveIdempotency semantics: first caller reserves the key, a replay with
// the same payload returns the original ID, and a different payload conflicts.
type fakeStore struct {
	transfers map[uuid.UUID]*models.Transfer
	keys      map[string]fakeIdempotency
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		transfers: make(map[uuid.UUID]*models.Transfer),
		keys:      make(map[string]fakeIdempotency),
	}
}

func (f *fakeStore) ReserveIdempotency(_ context.Context, key string, transferID uuid.UUID, requestHash string) (bool, uuid.UUID, error) {
	if existing, ok := f.keys[key]; ok {
		if existing.hash != requestHash {
			return false, uuid.Nil, applierrors.ErrIdempotencyKeyExists
		}
		return false, existing.transferID, nil
	}
	f.keys[key] = fakeIdempotency{transferID: transferID, hash: requestHash}
	return true, transferID, nil
}

func (f *fakeStore) FindByID(_ context.Context, transferID uuid.UUID) (*models.Transfer, error) {
	stored, ok := f.transfers[transferID]
	if !ok {
		return nil, applierrors.ErrTransferNotFound
	}
	copied := *stored
	return &copied, nil
}

func (f *fakeStore) Create(_ context.Context, t *models.Transfer) error {
	copied := *t
	f.transfers[t.ID] = &copied
	return nil
}

func (f *fakeStore) WriteAudit(_ context.Context, _ models.AuditEvent) error { return nil }

func (f *fakeStore) Delete(_ context.Context, t *models.Transfer) error {
	delete(f.transfers, t.ID)
	return nil
}

type fakePublisher struct {
	events []models.TransferEvent
}

func (p *fakePublisher) PublishTransferEvent(_ context.Context, ev models.TransferEvent) error {
	p.events = append(p.events, ev)
	return nil
}

func initiateForTest(t *testing.T, svc *TransferService, source, dest string, amount int64, key string) *models.TransferResponse {
	t.Helper()
	resp, err := svc.InitiateTransfer(context.Background(), source, models.CreateTransferRequest{
		DestinationAccountId: dest,
		AmountCents:          amount,
	}, key)
	if err != nil {
		t.Fatalf("initiate transfer: %v", err)
	}
	return resp
}

func TestGetTransferCrossAccountReturnsNotFound(t *testing.T) {
	svc := NewTransferService(newFakeStore(), &fakePublisher{})
	source, dest := uuid.NewString(), uuid.NewString()
	created := initiateForTest(t, svc, source, dest, 100, "key-1")

	outsider := uuid.NewString()
	if outsider == source || outsider == dest {
		t.Fatal("caller must not own the transfer")
	}
	_, err := svc.GetTransfer(context.Background(), created.TransferId, outsider)
	assertAppErrorCode(t, err, "TRANSFER_NOT_FOUND")
}

func TestGetTransferAllowsSourceAndDestination(t *testing.T) {
	svc := NewTransferService(newFakeStore(), &fakePublisher{})
	source, dest := uuid.NewString(), uuid.NewString()
	created := initiateForTest(t, svc, source, dest, 100, "key-1")

	for _, caller := range []string{source, dest} {
		got, err := svc.GetTransfer(context.Background(), created.TransferId, caller)
		if err != nil {
			t.Fatalf("owner %s should be allowed: %v", caller, err)
		}
		if got.TransferId != created.TransferId {
			t.Fatalf("expected transfer %s, got %s", created.TransferId, got.TransferId)
		}
	}
}

func TestInitiateTransferReplaysSameIdempotencyKeyWithoutRepeating(t *testing.T) {
	store := newFakeStore()
	pub := &fakePublisher{}
	svc := NewTransferService(store, pub)
	source, dest := uuid.NewString(), uuid.NewString()

	first := initiateForTest(t, svc, source, dest, 100, "key-1")
	second := initiateForTest(t, svc, source, dest, 100, "key-1")

	if first.TransferId != second.TransferId {
		t.Fatalf("replay should return the original transfer: %s != %s", first.TransferId, second.TransferId)
	}
	if len(store.transfers) != 1 {
		t.Fatalf("expected exactly one stored transfer, got %d", len(store.transfers))
	}
	if len(pub.events) != 1 {
		t.Fatalf("replay must not publish a second transfer_initiated event, got %d", len(pub.events))
	}
}

func TestInitiateTransferRejectsSameKeyWithDifferentPayload(t *testing.T) {
	svc := NewTransferService(newFakeStore(), &fakePublisher{})
	source, dest := uuid.NewString(), uuid.NewString()

	initiateForTest(t, svc, source, dest, 100, "key-1")
	_, err := svc.InitiateTransfer(context.Background(), source, models.CreateTransferRequest{
		DestinationAccountId: dest,
		AmountCents:          200,
	}, "key-1")
	assertAppErrorCode(t, err, "IDEMPOTENCY_KEY_EXISTS")
}

const injectionPayload = "'; DROP TABLE transfers; --"

func TestInitiateTransferRejectsInjectionInDestination(t *testing.T) {
	svc := NewTransferService(newFakeStore(), &fakePublisher{})
	_, err := svc.InitiateTransfer(context.Background(), uuid.NewString(), models.CreateTransferRequest{
		DestinationAccountId: injectionPayload,
		AmountCents:          100,
	}, "key-1")
	assertAppErrorCode(t, err, "VALIDATION_ERROR")
}

func TestGetTransferRejectsInjectionInTransferID(t *testing.T) {
	svc := NewTransferService(newFakeStore(), &fakePublisher{})
	_, err := svc.GetTransfer(context.Background(), injectionPayload, uuid.NewString())
	assertAppErrorCode(t, err, "TRANSFER_NOT_FOUND")
}
