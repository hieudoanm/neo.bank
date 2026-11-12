package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/core/auth"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/services"
)

func testRouter(t *testing.T) (*gin.Engine, *rsa.PrivateKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	km, err := auth.NewKeyManager(pubPEM, "test-key")
	if err != nil {
		t.Fatalf("key manager: %v", err)
	}
	svc := services.NewTransactionService(nil)
	router := NewRouter(handlers.NewTransactionHandler(svc), middleware.NewAuthMiddleware(km), nil)
	gin.SetMode(gin.TestMode)
	return router, priv
}

func signedToken(t *testing.T, priv *rsa.PrivateKey, accountID string) string {
	t.Helper()
	claims := auth.AccessClaims{
		AccountID: accountID,
		Email:     "customer@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(priv)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestListTransactionsRequiresAuth(t *testing.T) {
	router, _ := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+uuid.NewString()+"/transactions", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestListTransactionsRejectsOtherAccount(t *testing.T) {
	router, priv := testRouter(t)
	// The caller's token is for a different account than the path: the service
	// must not disclose the other account and returns 404 before touching storage.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+uuid.NewString()+"/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(t, priv, uuid.NewString()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestListTransactionsRejectsBadLimit(t *testing.T) {
	router, priv := testRouter(t)
	accountID := uuid.NewString()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+accountID+"/transactions?limit=0", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(t, priv, accountID))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHealthEndpoints(t *testing.T) {
	router, _ := testRouter(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, rec.Code)
		}
	}
}
