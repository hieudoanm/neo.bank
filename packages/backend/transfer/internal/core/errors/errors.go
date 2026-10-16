package errors

import "fmt"

type AppError struct {
	Code       string
	Message    string
	Details    map[string]string
	HTTPStatus int
}
func (e *AppError) Error() string { return fmt.Sprintf("%s: %s",e.Code,e.Message) }

var (
	ErrTransferNotFound  = &AppError{Code:"TRANSFER_NOT_FOUND",Message:"transfer not found",HTTPStatus:404}
	ErrAccountNotFound   = &AppError{Code:"ACCOUNT_NOT_FOUND",Message:"account not found",HTTPStatus:404}
	ErrUnauthorized      = &AppError{Code:"UNAUTHORIZED",Message:"unauthorized",HTTPStatus:401}
	ErrInvalidAmount     = &AppError{Code:"VALIDATION_ERROR",Message:"amount must be positive",HTTPStatus:400}
	ErrSameAccount       = &AppError{Code:"VALIDATION_ERROR",Message:"source and destination must differ",HTTPStatus:400}
	ErrInvalidDestination= &AppError{Code:"VALIDATION_ERROR",Message:"invalid destination account id",HTTPStatus:400}
	ErrInsufficientFunds = &AppError{Code:"INVALID_AMOUNT",Message:"insufficient funds",HTTPStatus:422}
	ErrIdempotencyConflict=&AppError{Code:"IDEMPOTENCY_CONFLICT",Message:"idempotency key already used with a different body",HTTPStatus:409}
	ErrTokenExpired      = &AppError{Code:"TOKEN_EXPIRED",Message:"access token expired or invalid",HTTPStatus:401}
	ErrInternal          = &AppError{Code:"INTERNAL_ERROR",Message:"an unexpected error occurred",HTTPStatus:500}
)
func ValidationError(f map[string]string)*AppError { return &AppError{Code:"VALIDATION_ERROR",Message:"validation failed",Details:f,HTTPStatus:400} }
