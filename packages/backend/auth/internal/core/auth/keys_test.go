package auth

import (
	"strings"
	"testing"
	"time"
)

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
