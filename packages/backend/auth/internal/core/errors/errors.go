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
	ErrInvalidCredentials     = &AppError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials", HTTPStatus: 401}
	ErrAccountNotFound        = &AppError{Code: "ACCOUNT_NOT_FOUND", Message: "account not found", HTTPStatus: 404}
	ErrUnauthorized           = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized", HTTPStatus: 401}
	ErrEmailAlreadyRegistered = &AppError{Code: "EMAIL_ALREADY_REGISTERED", Message: "registration failed", HTTPStatus: 409}
	ErrValidationError        = &AppError{Code: "VALIDATION_ERROR", Message: "validation failed", HTTPStatus: 400}
	ErrTokenExpired           = &AppError{Code: "TOKEN_EXPIRED", Message: "access token expired or invalid", HTTPStatus: 401}
	ErrInvalidRefreshToken    = &AppError{Code: "INVALID_REFRESH_TOKEN", Message: "refresh token expired, revoked, or already used", HTTPStatus: 401}
	ErrInternal               = &AppError{Code: "INTERNAL_ERROR", Message: "an unexpected error occurred", HTTPStatus: 500}
	ErrServiceUnavailable     = &AppError{Code: "SERVICE_UNAVAILABLE", Message: "service temporarily unavailable", HTTPStatus: 503}
)

func ValidationError(fields map[string]string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: "validation failed", Details: fields, HTTPStatus: 400}
}
