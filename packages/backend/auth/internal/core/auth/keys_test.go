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
)

func pemPrivatePKCS1(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

func pemPrivatePKCS8(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

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

func TestNewKeyManagerAcceptsPKCS1AndPKCS8(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	t.Run("pkcs1 with empty keyID generates one", func(t *testing.T) {
		m, err := NewKeyManager(pemPrivatePKCS1(t, key), pemPublicPKIX(t, &key.PublicKey), "")
		if err != nil {
			t.Fatalf("NewKeyManager: %v", err)
		}
		if m.PublicKey() == nil {
			t.Fatal("expected non-nil public key")
		}
	})

	t.Run("pkcs8 with keyID", func(t *testing.T) {
		m, err := NewKeyManager(pemPrivatePKCS8(t, key), pemPublicPKCS1(t, &key.PublicKey), "kid-1")
		if err != nil {
			t.Fatalf("NewKeyManager: %v", err)
		}
		token, err := m.SignAccessToken("u", "a", "e@example.com")
		if err != nil {
			t.Fatalf("SignAccessToken: %v", err)
		}
		if claims, err := m.Validate(token); err != nil || claims.Subject != "u" {
			t.Fatalf("Validate round-trip failed: claims=%v err=%v", claims, err)
		}
	})
}

func TestNewKeyManagerRejectsInvalidInput(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	validPriv := pemPrivatePKCS1(t, key)
	validPub := pemPublicPKIX(t, &key.PublicKey)

	if _, err := NewKeyManager([]byte("not pem"), validPub, "k"); err == nil {
		t.Fatal("expected error for invalid private PEM")
	}
	if _, err := NewKeyManager(validPriv, []byte("not pem"), "k"); err == nil {
		t.Fatal("expected error for invalid public PEM")
	}
}

func TestNewKeyManagerRejectsSmallKey(t *testing.T) {
	small, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if _, err := NewKeyManager(pemPrivatePKCS1(t, small), pemPublicPKIX(t, &small.PublicKey), "k"); err == nil {
		t.Fatal("expected error for RSA key smaller than 2048 bits")
	}
}

func TestNewKeyManagerRejectsNonRSAKeys(t *testing.T) {
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	ecDER, err := x509.MarshalPKCS8PrivateKey(ec)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	ecPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: ecDER})
	ecPubDER, err := x509.MarshalPKIXPublicKey(&ec.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	ecPubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: ecPubDER})

	if _, err := NewKeyManager(ecPrivPEM, ecPubPEM, "k"); err == nil {
		t.Fatal("expected error for non-RSA private key")
	}

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if _, err := NewKeyManager(pemPrivatePKCS1(t, rsaKey), ecPubPEM, "k"); err == nil {
		t.Fatal("expected error for non-RSA public key")
	}
}

func TestGenerateDevKeyPairEmptyKeyID(t *testing.T) {
	m, err := GenerateDevKeyPair("")
	if err != nil {
		t.Fatalf("GenerateDevKeyPair: %v", err)
	}
	token, err := m.SignAccessToken("user", "acct", "x@example.com")
	if err != nil {
		t.Fatalf("SignAccessToken: %v", err)
	}
	if _, err := m.Validate(token); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestSignAndValidateAccessToken(t *testing.T) {
	m, err := GenerateDevKeyPair("test-key-1")
	if err != nil {
		t.Fatalf("GenerateDevKeyPair: %v", err)
	}

	token, err := m.SignAccessToken("user-1", "account-1", "user@example.com")
	if err != nil {
		t.Fatalf("SignAccessToken: %v", err)
	}
	if token == "" {
		t.Fatal("SignAccessToken returned empty token")
	}

	claims, err := m.Validate(token)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
	if claims.AccountID != "account-1" {
		t.Fatalf("expected accountId account-1, got %s", claims.AccountID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", claims.Email)
	}
	if claims.ExpiresAt == nil || time.Now().After(claims.ExpiresAt.Time) {
		t.Fatal("token is expired or missing expiry")
	}
}

func TestValidateRejectsTamperedToken(t *testing.T) {
	m, err := GenerateDevKeyPair("test-key-2")
	if err != nil {
		t.Fatalf("GenerateDevKeyPair: %v", err)
	}
	token, err := m.SignAccessToken("user-1", "account-1", "user@example.com")
	if err != nil {
		t.Fatalf("SignAccessToken: %v", err)
	}
	tampered := token + "tampered"
	if _, err := m.Validate(tampered); err == nil {
		t.Fatal("expected Validate to reject tampered token")
	}
}

func TestValidateRejectsWrongKey(t *testing.T) {
	m1, err := GenerateDevKeyPair("key-1")
	if err != nil {
		t.Fatalf("GenerateDevKeyPair: %v", err)
	}
	m2, err := GenerateDevKeyPair("key-2")
	if err != nil {
		t.Fatalf("GenerateDevKeyPair: %v", err)
	}
	token, err := m1.SignAccessToken("user-1", "account-1", "user@example.com")
	if err != nil {
		t.Fatalf("SignAccessToken: %v", err)
	}
	if _, err := m2.Validate(token); err == nil {
		t.Fatal("expected Validate with wrong key to fail")
	}
}

func TestHashRefreshToken(t *testing.T) {
	h1 := HashRefreshToken("token-value")
	h2 := HashRefreshToken("token-value")
	h3 := HashRefreshToken("other-value")
	if h1 != h2 {
		t.Fatal("same token should hash identically")
	}
	if h1 == h3 {
		t.Fatal("different tokens should hash differently")
	}
	if strings.Contains(h1, "token-value") {
		t.Fatal("hash should not contain the raw token")
	}
}
