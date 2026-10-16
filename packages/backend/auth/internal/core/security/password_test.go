package security

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("Str0ng!Passw0rd")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == "Str0ng!Passw0rd" {
		t.Fatal("HashPassword returned plaintext password")
	}
}

func TestHashPasswordTooShort(t *testing.T) {
	_, err := HashPassword("short1A!")
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestVerifyPassword(t *testing.T) {
	hash, err := HashPassword("Str0ng!Passw0rd")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if !VerifyPassword("Str0ng!Passw0rd", hash) {
		t.Fatal("VerifyPassword returned false for correct password")
	}
	if VerifyPassword("wrong password", hash) {
		t.Fatal("VerifyPassword returned true for wrong password")
	}
	if VerifyPassword("Str0ng!Passw0rd", "not-a-hash") {
		t.Fatal("VerifyPassword returned true for malformed hash")
	}
}
