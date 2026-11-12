package errors

import (
	"testing"
)

func TestAppErrorError(t *testing.T) {
	err := &AppError{Code: "CODE", Message: "something went wrong"}
	if got, want := err.Error(), "CODE: something went wrong"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError(map[string]string{"email": "required"})
	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("Code = %q, want VALIDATION_ERROR", err.Code)
	}
	if err.HTTPStatus != 400 {
		t.Errorf("HTTPStatus = %d, want 400", err.HTTPStatus)
	}
	if err.Details["email"] != "required" {
		t.Errorf("Details[email] = %q, want required", err.Details["email"])
	}
}
