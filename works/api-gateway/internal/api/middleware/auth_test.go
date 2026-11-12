package middleware

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
	"github.com/hieudoanm/neo.bank/works/api-gateway/internal/core/auth"
)

func newTestKeyManager(t *testing.T) (*auth.KeyManager, *rsa.PrivateKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	km, err := auth.NewKeyManager(pubPEM, "test-key")
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	return km, privateKey
}

func signToken(t *testing.T, privateKey *rsa.PrivateKey, accountID, email string, expired bool) string {
	t.Helper()
	now := time.Now()
	claims := &testClaims{
		AccountID: accountID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			ID:        "jti-1",
		},
	}
	if expired {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Minute))
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

type testClaims struct {
	AccountID string `json:"accountId"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

func requireAuthHandler(c *gin.Context) {
	userID, _ := c.Get("userId")
	accountID, _ := c.Get("accountId")
	c.JSON(http.StatusOK, gin.H{"userId": userID, "accountId": accountID})
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	km, _ := newTestKeyManager(t)
	m := NewAuthMiddleware(km)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts/11111111-1111-1111-1111-111111111111/balance", nil)
	m.RequireAuth(c)
	requireAuthHandler(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "UNAUTHENTICATED") && !strings.Contains(w.Body.String(), "authentication") {
		t.Fatalf("expected auth error body, got %s", w.Body.String())
	}
	if c.IsAborted() == false {
		t.Fatal("expected request to be aborted")
	}
}

func TestRequireAuthAcceptsValidToken(t *testing.T) {
	km, privateKey := newTestKeyManager(t)
	m := NewAuthMiddleware(km)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts/11111111-1111-1111-1111-111111111111/balance", nil)
	c.Request.Header.Set("Authorization", "Bearer "+signToken(t, privateKey, "account-1", "user@example.com", false))
	m.RequireAuth(c)
	requireAuthHandler(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if userID, _ := c.Get("userId"); userID != "user-1" {
		t.Fatalf("expected userId user-1, got %v", userID)
	}
}

func TestRequireAuthRejectsExpiredToken(t *testing.T) {
	km, privateKey := newTestKeyManager(t)
	m := NewAuthMiddleware(km)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts/11111111-1111-1111-1111-111111111111/balance", nil)
	c.Request.Header.Set("Authorization", "Bearer "+signToken(t, privateKey, "account-1", "user@example.com", true))
	m.RequireAuth(c)
	requireAuthHandler(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", w.Code)
	}
}

func TestRequireAuthRejectsBadScheme(t *testing.T) {
	km, _ := newTestKeyManager(t)
	m := NewAuthMiddleware(km)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts/11111111-1111-1111-1111-111111111111/balance", nil)
	c.Request.Header.Set("Authorization", "Basic abc")
	m.RequireAuth(c)
	requireAuthHandler(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for bad auth scheme, got %d", w.Code)
	}
}
