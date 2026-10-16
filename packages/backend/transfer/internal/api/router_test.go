package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/services"
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
	svc := services.NewTransferService(nil, nil)
	router := NewRouter(handlers.NewTransferHandler(svc), middleware.NewAuthMiddleware(km))
	gin.SetMode(gin.TestMode)
	return router, priv
}

func signedToken(t *testing.T, priv *rsa.PrivateKey) string {
	t.Helper()
	claims := auth.AccessClaims{
		AccountID: uuid.NewString(),
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

func TestCreateTransferRequiresAuth(t *testing.T) {
	router, _ := testRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transfers", strings.NewReader(`{"destinationAccountId":"`+uuid.NewString()+`","amountCents":100}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateTransferRequiresIdempotencyKey(t *testing.T) {
	router, priv := testRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transfers", strings.NewReader(`{"destinationAccountId":"`+uuid.NewString()+`","amountCents":100}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signedToken(t, priv))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing Idempotency-Key, got %d", rec.Code)
	}
}

func TestCreateTransferRejectsMalformedIdempotencyKey(t *testing.T) {
	router, priv := testRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transfers", strings.NewReader(`{"destinationAccountId":"`+uuid.NewString()+`","amountCents":100}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signedToken(t, priv))
	req.Header.Set("Idempotency-Key", "bad key with spaces")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed Idempotency-Key, got %d", rec.Code)
	}
}

func TestGetTransferRequiresAuth(t *testing.T) {
	router, _ := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transfers/"+uuid.NewString(), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetTransferUnknownReturnsNotFound(t *testing.T) {
	router, priv := testRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transfers/not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(t, priv))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an unknown transfer, got %d", rec.Code)
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
