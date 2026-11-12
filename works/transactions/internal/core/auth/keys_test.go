package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func pemPublicPKIX(t *testing.T, key *rsa.PublicKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

func pemPublicPKCS1(t *testing.T, key *rsa.PublicKey) []byte {
	t.Helper()
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(key)})
}

func signToken(t *testing.T, priv *rsa.PrivateKey, claims AccessClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}
	return signed
}

func validClaims(expiry time.Time) AccessClaims {
	return AccessClaims{
		AccountID: "account-1",
		Email:     "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}
}

func TestNewKeyManagerAcceptsPKIXAndPKCS1(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	t.Run("pkix with keyID", func(t *testing.T) {
		if _, err := NewKeyManager(pemPublicPKIX(t, &priv.PublicKey), "kid"); err != nil {
			t.Fatalf("NewKeyManager: %v", err)
		}
	})
	t.Run("pkcs1 with empty keyID generates one", func(t *testing.T) {
		km, err := NewKeyManager(pemPublicPKCS1(t, &priv.PublicKey), "")
		if err != nil {
			t.Fatalf("NewKeyManager: %v", err)
		}
		if km.keyID == "" {
			t.Fatal("expected generated keyID")
		}
	})
	t.Run("empty PEM leaves public key unset", func(t *testing.T) {
		km, err := NewKeyManager(nil, "kid")
		if err != nil {
			t.Fatalf("NewKeyManager: %v", err)
		}
		if km.publicKey != nil {
			t.Fatal("expected nil public key")
		}
	})
}

func TestNewKeyManagerRejectsBadKeys(t *testing.T) {
	if _, err := NewKeyManager([]byte("not pem"), "k"); err == nil {
		t.Fatal("expected error for invalid PEM")
	}

	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&ec.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	ecPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	if _, err := NewKeyManager(ecPEM, "k"); err == nil {
		t.Fatal("expected error for non-RSA public key")
	}
}

func TestValidateWithoutKeyFails(t *testing.T) {
	km, err := NewKeyManager(nil, "k")
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	if _, err := km.Validate("anything"); err == nil {
		t.Fatal("expected error when no public key is configured")
	}
}

func TestValidateAcceptsSignedToken(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	km, err := NewKeyManager(pemPublicPKIX(t, &priv.PublicKey), "k")
	if err != nil {
		t.Fatalf("NewKeyManager: %v", err)
	}
	token := signToken(t, priv, validClaims(time.Now().Add(time.Hour)))
	claims, err := km.Validate(token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.AccountID != "account-1" || claims.Email != "user@example.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestValidateRejectsExpired(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	km, _ := NewKeyManager(pemPublicPKIX(t, &priv.PublicKey), "k")
	token := signToken(t, priv, validClaims(time.Now().Add(-time.Hour)))
	_, err = km.Validate(token)
	if err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestValidateRejectsWrongKeyAndMissingClaims(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	other, _ := rsa.GenerateKey(rand.Reader, 2048)

	km, _ := NewKeyManager(pemPublicPKIX(t, &other.PublicKey), "k")
	token := signToken(t, priv, validClaims(time.Now().Add(time.Hour)))
	if _, err := km.Validate(token); err == nil {
		t.Fatal("expected token signed by a different key to be rejected")
	}

	km, _ = NewKeyManager(pemPublicPKIX(t, &priv.PublicKey), "k")
	missing := validClaims(time.Now().Add(time.Hour))
	missing.AccountID = ""
	if _, err := km.Validate(signToken(t, priv, missing)); err == nil {
		t.Fatal("expected token missing accountId to be rejected")
	}
}

func TestValidateRejectsWrongSigningMethod(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	km, _ := NewKeyManager(pemPublicPKIX(t, &priv.PublicKey), "k")
	claims := validClaims(time.Now().Add(time.Hour))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}
	if _, err := km.Validate(signed); err == nil {
		t.Fatal("expected HMAC token to be rejected")
	}
}
