package errors

import "testing"

func TestAppErrorError(t *testing.T) {
	err := &AppError{Code: "CODE", Message: "boom"}
	if got := err.Error(); got != "CODE: boom" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestSentinelErrors(t *testing.T) {
	cases := []struct {
		err    *AppError
		code   string
		status int
	}{
		{ErrAccountNotFound, "ACCOUNT_NOT_FOUND", 404},
		{ErrUnauthorized, "UNAUTHORIZED", 401},
		{ErrTokenExpired, "TOKEN_EXPIRED", 401},
		{ErrInvalidCursor, "VALIDATION_ERROR", 400},
		{ErrInvalidLimit, "VALIDATION_ERROR", 400},
		{ErrInternal, "INTERNAL_ERROR", 500},
	}
	for _, c := range cases {
		if c.err.Code != c.code || c.err.HTTPStatus != c.status {
			t.Fatalf("unexpected sentinel: %+v", c.err)
		}
	}
}

func TestValidationError(t *testing.T) {
	fields := map[string]string{"limit": "out of range"}
	err := ValidationError(fields)
	if err.Code != "VALIDATION_ERROR" || err.HTTPStatus != 400 {
		t.Fatalf("unexpected validation error: %+v", err)
	}
	if err.Details["limit"] != "out of range" {
		t.Fatalf("unexpected details: %+v", err.Details)
	}
}
