package errors

import "fmt"

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrInvalidCredentials = &AppError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials"}
	ErrAccountNotFound    = &AppError{Code: "ACCOUNT_NOT_FOUND", Message: "account not found"}
	ErrUnauthorized       = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized"}
)
