package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	applierrors "github.com/hieudoanm/neo.bank/works/accounts/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/models"
)

type fakeBalanceService struct {
	resp *models.BalanceResponse
	err  error
}

func (f *fakeBalanceService) FetchBalance(_ context.Context, accountID uuid.UUID, _ string) (*models.BalanceResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.resp, nil
}

func newGin(svc BalanceService, callerAccountID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewAccountHandler(svc)
	r := gin.New()
	r.GET("/accounts/:accountId/balance", func(c *gin.Context) {
		c.Set("accountId", callerAccountID)
		h.GetBalance(c)
	})
	return r
}

func TestGetBalanceSuccess(t *testing.T) {
	acc := uuid.New()
	svc := &fakeBalanceService{resp: &models.BalanceResponse{AccountID: acc.String(), BalanceCents: 500, CurrencyCode: "USD", UpdatedAt: "2026-03-15T10:00:00Z"}}
	r := newGin(svc, acc.String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/accounts/"+acc.String()+"/balance", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp models.BalanceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.BalanceCents != 500 {
		t.Fatalf("unexpected body: %+v", resp)
	}
}

func TestGetBalanceInvalidUUIDReturns404(t *testing.T) {
	r := newGin(&fakeBalanceService{}, uuid.NewString())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/accounts/not-a-uuid/balance", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestGetBalanceMapsAppError(t *testing.T) {
	acc := uuid.New()
	r := newGin(&fakeBalanceService{err: applierrors.ErrAccountNotFound}, acc.String())
	req := httptest.NewRequest(http.MethodGet, "/accounts/"+acc.String()+"/balance", nil)
	req.Header.Set("X-Request-ID", "rid-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"requestId":"rid-1"`) {
		t.Fatalf("expected requestId in body: %s", w.Body.String())
	}
}

func TestGetBalanceMapsUnknownErrorTo500(t *testing.T) {
	acc := uuid.New()
	r := newGin(&fakeBalanceService{err: errors.New("boom")}, acc.String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/accounts/"+acc.String()+"/balance", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if !strings.Contains(w.Body.String(), "INTERNAL_ERROR") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
