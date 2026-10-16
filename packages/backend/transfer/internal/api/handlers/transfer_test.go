package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	applierrors "github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
)

type fakeService struct {
	initiateResp *models.TransferResponse
	initiateErr  error
	getResp      *models.TransferResponse
	getErr       error
}

func (f *fakeService) InitiateTransfer(_ context.Context, _ string, _ models.CreateTransferRequest, _ string) (*models.TransferResponse, error) {
	return f.initiateResp, f.initiateErr
}

func (f *fakeService) GetTransfer(_ context.Context, _, _ string) (*models.TransferResponse, error) {
	return f.getResp, f.getErr
}

func newRouter(svc TransferService, caller string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewTransferHandler(svc)
	r := gin.New()
	r.POST("/api/v1/transfers", func(c *gin.Context) {
		c.Set("accountId", caller)
		h.CreateTransfer(c)
	})
	r.GET("/api/v1/transfers/:transferId", func(c *gin.Context) {
		c.Set("accountId", caller)
		h.GetTransfer(c)
	})
	return r
}

func perform(r *gin.Engine, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreateTransferBadIdempotency(t *testing.T) {
	caller := uuid.NewString()
	dest := uuid.NewString()
	r := newRouter(&fakeService{}, caller)
	w := perform(r, http.MethodPost, "/api/v1/transfers", `{"destinationAccountId":"`+dest+`","amountCents":100}`, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestCreateTransferInvalidBody(t *testing.T) {
	caller := uuid.NewString()
	r := newRouter(&fakeService{}, caller)
	w := perform(r, http.MethodPost, "/api/v1/transfers", `{`, map[string]string{"Idempotency-Key": "abc123"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestCreateTransferSuccess(t *testing.T) {
	caller := uuid.NewString()
	resp := &models.TransferResponse{TransferId: uuid.NewString(), Status: models.StatusPending}
	r := newRouter(&fakeService{initiateResp: resp}, caller)
	w := perform(r, http.MethodPost, "/api/v1/transfers", `{"destinationAccountId":"`+uuid.NewString()+`","amountCents":100}`, map[string]string{"Idempotency-Key": "abc123"})
	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", w.Code)
	}
	if w.Header().Get("Idempotency-Key") != "abc123" {
		t.Fatalf("idempotency header not echoed")
	}
}

func TestCreateTransferMapsError(t *testing.T) {
	caller := uuid.NewString()
	r := newRouter(&fakeService{initiateErr: applierrors.ErrInsufficientFunds}, caller)
	w := perform(r, http.MethodPost, "/api/v1/transfers", `{"destinationAccountId":"`+uuid.NewString()+`","amountCents":100}`, map[string]string{"Idempotency-Key": "abc123"})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", w.Code)
	}
}

func TestCreateTransferInternalError(t *testing.T) {
	caller := uuid.NewString()
	r := newRouter(&fakeService{initiateErr: errors.New("boom")}, caller)
	w := perform(r, http.MethodPost, "/api/v1/transfers", `{"destinationAccountId":"`+uuid.NewString()+`","amountCents":100}`, map[string]string{"Idempotency-Key": "abc123"})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

func TestGetTransfer(t *testing.T) {
	caller := uuid.NewString()
	resp := &models.TransferResponse{TransferId: uuid.NewString()}
	r := newRouter(&fakeService{getResp: resp}, caller)
	w := perform(r, http.MethodGet, "/api/v1/transfers/"+uuid.NewString(), "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	r = newRouter(&fakeService{getErr: applierrors.ErrTransferNotFound}, caller)
	w = perform(r, http.MethodGet, "/api/v1/transfers/"+uuid.NewString(), "", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
