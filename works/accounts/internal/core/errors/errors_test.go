package errors

import (
	"errors"
	"testing"
)

func TestAppErrorError(t *testing.T) {
	e := &AppError{Code: "ACCOUNT_NOT_FOUND", Message: "account not found", HTTPStatus: 404}
	if got := e.Error(); got != "ACCOUNT_NOT_FOUND: account not found" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestValidationError(t *testing.T) {
	e := ValidationError(map[string]string{"accountId": "invalid"})
	if e.Code != "VALIDATION_ERROR" || e.HTTPStatus != 400 || e.Details["accountId"] != "invalid" {
		t.Fatalf("unexpected error: %+v", e)
	}
}

func TestPredefinedErrorsAs(t *testing.T) {
	var appErr *AppError
	if !errors.As(ErrInsufficientFunds, &appErr) {
		t.Fatal("ErrInsufficientFunds should be an *AppError")
	}
	if appErr.HTTPStatus != 422 {
		t.Fatalf("unexpected status: %d", appErr.HTTPStatus)
	}
}
