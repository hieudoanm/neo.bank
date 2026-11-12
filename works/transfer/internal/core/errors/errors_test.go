package errors

import (
	"errors"
	"testing"
)

func TestAppErrorError(t *testing.T) {
	e := &AppError{Code: "TRANSFER_NOT_FOUND", Message: "transfer not found", HTTPStatus: 404}
	if got := e.Error(); got != "TRANSFER_NOT_FOUND: transfer not found" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestValidationError(t *testing.T) {
	e := ValidationError(map[string]string{"amountCents": "must be positive"})
	if e.Code != "VALIDATION_ERROR" || e.HTTPStatus != 400 || e.Details["amountCents"] != "must be positive" {
		t.Fatalf("unexpected error: %+v", e)
	}
}

func TestPredefinedErrorsAs(t *testing.T) {
	var appErr *AppError
	if !errors.As(ErrIdempotencyKeyExists, &appErr) {
		t.Fatal("ErrIdempotencyKeyExists should be an *AppError")
	}
	if appErr.HTTPStatus != 409 {
		t.Fatalf("unexpected status: %d", appErr.HTTPStatus)
	}
}
