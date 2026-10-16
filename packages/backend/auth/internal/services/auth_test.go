package services

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	if err := validateEmail(""); err == nil {
		t.Fatal("expected error for empty email")
	}
	if err := validateEmail("user@example.com"); err != nil {
		t.Fatalf("expected no error for valid email, got %v", err)
	}
}

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		password string
		valid    bool
	}{
		{"", false},
		{"short1A!", false},
		{"nomatches", false},
		{"NoDigitsOrSymbols!", false},
		{"lowercaseandsymbols!", false},
		{"UPPERCASEANDSYMBOLS!", false},
		{"Str0ng!Passw0rd", true},
		{"Abcdefg12345!", true},
	}
	for _, c := range cases {
		err := validatePassword(c.password)
		if c.valid && err != nil {
			t.Errorf("expected %q to be valid, got %v", c.password, err)
		}
		if !c.valid && err == nil {
			t.Errorf("expected %q to be invalid", c.password)
		}
	}
}
