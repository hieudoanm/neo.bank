package errors

import "fmt"

type AppError struct {
	Code       string
	Message    string
	Details    map[string]string
	HTTPStatus int
}

func (e *AppError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

var (
	ErrAccountNotFound = &AppError{Code: "ACCOUNT_NOT_FOUND", Message: "account not found", HTTPStatus: 404}
	ErrUnauthorized    = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized", HTTPStatus: 401}
	ErrTokenExpired    = &AppError{Code: "TOKEN_EXPIRED", Message: "access token expired or invalid", HTTPStatus: 401}
	ErrInvalidCursor   = &AppError{Code: "VALIDATION_ERROR", Message: "invalid pagination cursor", HTTPStatus: 400}
	ErrInvalidLimit    = &AppError{Code: "VALIDATION_ERROR", Message: "limit must be between 1 and 100", HTTPStatus: 400}
	ErrInternal        = &AppError{Code: "INTERNAL_ERROR", Message: "an unexpected error occurred", HTTPStatus: 500}
)

func ValidationError(f map[string]string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: "validation failed", Details: f, HTTPStatus: 400}
}
