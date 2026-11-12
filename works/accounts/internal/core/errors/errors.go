package errors

import "fmt"

type AppError struct {
	Code       string
	Message    string
	Details    map[string]string
	HTTPStatus int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrAccountNotFound    = &AppError{Code: "ACCOUNT_NOT_FOUND", Message: "account not found", HTTPStatus: 404}
	ErrInsufficientFunds  = &AppError{Code: "INSUFFICIENT_FUNDS", Message: "insufficient funds", HTTPStatus: 422}
	ErrUnauthorized       = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized", HTTPStatus: 401}
	ErrInternal           = &AppError{Code: "INTERNAL_ERROR", Message: "an unexpected error occurred", HTTPStatus: 500}
	ErrInvalidCredentials = &AppError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials", HTTPStatus: 401}
	ErrTokenExpired       = &AppError{Code: "TOKEN_EXPIRED", Message: "access token expired or invalid", HTTPStatus: 401}
)

func ValidationError(fields map[string]string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: "validation failed", Details: fields, HTTPStatus: 400}
}
