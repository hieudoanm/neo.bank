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
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/auth"
)

func newTestKeyManager(t *testing.T) (*auth.KeyManager, *rsa.PrivateKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	km, err := auth.NewKeyManager(pubPEM, "k")
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	return km, priv
}

func newTestEngine(t *testing.T) (*gin.Engine, *rsa.PrivateKey) {
	t.Helper()
	km, priv := newTestKeyManager(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewAuthMiddleware(km).RequireAuth)
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"accountId": c.GetString("accountId"), "email": c.GetString("email")})
	})
	return r, priv
}

func signedWith(t *testing.T, priv *rsa.PrivateKey, claims auth.AccessClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}
	return signed
}

func TestRequireAuthRejectsNonBearer(t *testing.T) {
	r, _ := newTestEngine(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic abc")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestRequireAuthRejectsInvalidToken(t *testing.T) {
	r, _ := newTestEngine(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestRequireAuthRejectsExpiredToken(t *testing.T) {
	r, priv := newTestEngine(t)
	claims := auth.AccessClaims{
		AccountID: uuid.NewString(),
		Email:     "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedWith(t, priv, claims))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestRequireAuthAllowsValidToken(t *testing.T) {
	r, priv := newTestEngine(t)
	accountID := uuid.NewString()
	claims := auth.AccessClaims{
		AccountID: accountID,
		Email:     "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedWith(t, priv, claims))
	req.Header.Set("X-Request-ID", "rid")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), accountID) {
		t.Fatalf("expected accountId in body: %s", w.Body.String())
	}
}
