package errors

import (
	"errors"
	"testing"
)

func TestAppErrorError(t *testing.T) {
	e := &AppError{Code: "SOME_CODE", Message: "some message", HTTPStatus: 400}
	if got := e.Error(); got != "SOME_CODE: some message" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestValidationError(t *testing.T) {
	fields := map[string]string{"email": "email is required"}
	e := ValidationError(fields)
	if e.Code != "VALIDATION_ERROR" || e.HTTPStatus != 400 {
		t.Fatalf("unexpected error: %+v", e)
	}
	if e.Details["email"] != "email is required" {
		t.Fatalf("details not set: %+v", e.Details)
	}
}

func TestPredefinedErrorsAs(t *testing.T) {
	var appErr *AppError
	if !errors.As(ErrInvalidCredentials, &appErr) {
		t.Fatal("ErrInvalidCredentials should be an *AppError")
	}
	if appErr.Code != "INVALID_CREDENTIALS" || appErr.HTTPStatus != 401 {
		t.Fatalf("unexpected ErrInvalidCredentials: %+v", appErr)
	}
}
