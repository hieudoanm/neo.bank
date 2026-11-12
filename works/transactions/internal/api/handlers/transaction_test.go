package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	applierrors "github.com/hieudoanm/neo.bank/works/transactions/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/models"
)

type fakeReader struct {
	resp      *models.TransactionListResponse
	err       error
	gotLimit  int
	gotCursor string
	gotCaller string
}

func (f *fakeReader) ListTransactions(_ context.Context, _ uuid.UUID, callerAccountID, cursor string, limit int) (*models.TransactionListResponse, error) {
	f.gotCaller = callerAccountID
	f.gotCursor = cursor
	f.gotLimit = limit
	return f.resp, f.err
}

func newRouter(svc TransactionReader, caller string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewTransactionHandler(svc)
	r := gin.New()
	r.GET("/api/v1/accounts/:accountId/transactions", func(c *gin.Context) {
		c.Set("accountId", caller)
		h.ListTransactions(c)
	})
	return r
}

func TestListTransactionsSuccess(t *testing.T) {
	accountID := uuid.NewString()
	svc := &fakeReader{resp: &models.TransactionListResponse{Transactions: []models.TransactionResponse{}}}
	r := newRouter(svc, accountID)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+accountID+"/transactions?limit=5&cursor=abc", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if svc.gotLimit != 5 || svc.gotCursor != "abc" || svc.gotCaller != accountID {
		t.Fatalf("unexpected args: %+v", svc)
	}
}

func TestListTransactionsRejectsInvalidAccount(t *testing.T) {
	r := newRouter(&fakeReader{}, uuid.NewString())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/not-a-uuid/transactions", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestListTransactionsRejectsInvalidLimit(t *testing.T) {
	accountID := uuid.NewString()
	for _, raw := range []string{"abc", "0", "101"} {
		r := newRouter(&fakeReader{}, accountID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+accountID+"/transactions?limit="+raw, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("limit=%s status = %d, want 400", raw, w.Code)
		}
	}
}

func TestListTransactionsMapsErrors(t *testing.T) {
	accountID := uuid.NewString()
	cases := []struct {
		err    error
		status int
	}{
		{applierrors.ErrInvalidCursor, http.StatusBadRequest},
		{applierrors.ErrAccountNotFound, http.StatusNotFound},
		{errors.New("boom"), http.StatusInternalServerError},
	}
	for _, c := range cases {
		r := newRouter(&fakeReader{err: c.err}, accountID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+accountID+"/transactions", nil)
		r.ServeHTTP(w, req)
		if w.Code != c.status {
			t.Fatalf("err=%v status = %d, want %d", c.err, w.Code, c.status)
		}
	}
}
