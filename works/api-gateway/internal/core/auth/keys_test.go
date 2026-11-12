package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	return key
}

func pkixPublicKeyPEM(t *testing.T, pub *rsa.PublicKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal pkix public key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

func pkcs1PublicKeyPEM(t *testing.T, pub *rsa.PublicKey) []byte {
	t.Helper()
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(pub)})
}

func signToken(t *testing.T, key *rsa.PrivateKey, claims jwt.Claims) string {
	t.Helper()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func validClaims(now time.Time) *AccessClaims {
	return &AccessClaims{
		AccountID: "a1",
		Email:     "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "u1",
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
}

func TestNewKeyManagerAcceptsPKIXAndPKCS1(t *testing.T) {
	key := generateRSAKey(t)

	if _, err := NewKeyManager(pkixPublicKeyPEM(t, &key.PublicKey), "kid"); err != nil {
		t.Fatalf("NewKeyManager(pkix) error: %v", err)
	}
	if _, err := NewKeyManager(pkcs1PublicKeyPEM(t, &key.PublicKey), "kid"); err != nil {
		t.Fatalf("NewKeyManager(pkcs1) error: %v", err)
	}
}

func TestNewKeyManagerGeneratesKeyIDWhenEmpty(t *testing.T) {
	key := generateRSAKey(t)
	km, err := NewKeyManager(pkixPublicKeyPEM(t, &key.PublicKey), "")
	if err != nil {
		t.Fatalf("NewKeyManager error: %v", err)
	}
	if km.keyID == "" {
		t.Fatal("expected a generated keyID")
	}
}

func TestNewKeyManagerRejectsInvalidPEM(t *testing.T) {
	if _, err := NewKeyManager([]byte("not a pem block"), "kid"); err == nil {
		t.Fatal("expected error for missing PEM block")
	}
}

func TestNewKeyManagerRejectsNonRSAKey(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ecdsa key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&ecKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal pkix: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	if _, err := NewKeyManager(pemBytes, "kid"); err == nil {
		t.Fatal("expected error for non-RSA key")
	}
}

func TestNewKeyManagerRejectsGarbageBlock(t *testing.T) {
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("garbage")})
	if _, err := NewKeyManager(pemBytes, "kid"); err == nil {
		t.Fatal("expected error for unparseable key block")
	}
}

func TestValidateAcceptsValidToken(t *testing.T) {
	key := generateRSAKey(t)
	km, err := NewKeyManager(pkixPublicKeyPEM(t, &key.PublicKey), "kid")
	if err != nil {
		t.Fatalf("NewKeyManager error: %v", err)
	}

	claims, err := km.Validate(signToken(t, key, validClaims(time.Now())))
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	if claims.Subject != "u1" || claims.AccountID != "a1" || claims.Email != "user@example.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestValidateRejectsExpiredToken(t *testing.T) {
	key := generateRSAKey(t)
	km, _ := NewKeyManager(pkixPublicKeyPEM(t, &key.PublicKey), "kid")

	claims := validClaims(time.Now())
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))

	_, err := km.Validate(signToken(t, key, claims))
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestValidateRejectsWrongSigningMethod(t *testing.T) {
	key := generateRSAKey(t)
	km, _ := NewKeyManager(pkixPublicKeyPEM(t, &key.PublicKey), "kid")

	hsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims(time.Now()))
	signed, err := hsToken.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign hs token: %v", err)
	}

	if _, err := km.Validate(signed); err == nil {
		t.Fatal("expected error for unexpected signing method")
	}
}

func TestValidateRejectsMissingClaims(t *testing.T) {
	key := generateRSAKey(t)
	km, _ := NewKeyManager(pkixPublicKeyPEM(t, &key.PublicKey), "kid")

	claims := &AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "u1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	if _, err := km.Validate(signToken(t, key, claims)); err == nil {
		t.Fatal("expected error for missing required claims")
	}
}

func TestValidateRejectsMalformedToken(t *testing.T) {
	key := generateRSAKey(t)
	km, _ := NewKeyManager(pkixPublicKeyPEM(t, &key.PublicKey), "kid")

	if _, err := km.Validate("not.a.jwt"); err == nil {
		t.Fatal("expected error for malformed token")
	}
}
