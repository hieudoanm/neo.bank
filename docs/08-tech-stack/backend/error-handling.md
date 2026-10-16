# NEO Bank — Go Best Practices — Error Handling

_Part of [go.md](./go.md)._

## Purpose

Defines error types, the error response envelope, and error handling patterns used consistently across all NEO Bank backend services.

## Error Types

```go
// internal/core/errors/errors.go
package errors

import (
    "errors"
    "fmt"
    "net/http"
)

// AppError is a structured application error with a stable code.
type AppError struct {
    Code    string            `json:"-"`
    Message string            `json:"-"`
    Details map[string]string `json:"-"`
    Err     error             `json:"-"`
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// HTTPStatus maps error codes to HTTP status codes (docs/06-api-specification.md §5).
func (e *AppError) HTTPStatus() int {
    switch e.Code {
    case "VALIDATION_ERROR", "INVALID_AMOUNT", "INSUFFICIENT_FUNDS", "SAME_ACCOUNT":
        return http.StatusBadRequest
    case "UNAUTHENTICATED", "INVALID_CREDENTIALS", "INVALID_REFRESH_TOKEN", "TOKEN_EXPIRED":
        return http.StatusUnauthorized
    case "ACCOUNT_NOT_FOUND", "TRANSFER_NOT_FOUND":
        return http.StatusNotFound
    case "EMAIL_ALREADY_REGISTERED", "IDEMPOTENCY_KEY_EXISTS":
        return http.StatusConflict
    case "RATE_LIMITED":
        return http.StatusTooManyRequests
    case "INTERNAL_ERROR":
        return http.StatusInternalServerError
    case "SERVICE_UNAVAILABLE":
        return http.StatusServiceUnavailable
    default:
        return http.StatusInternalServerError
    }
}
```

### Predefined sentinel errors

```go
var (
    ErrUnauthorized       = &AppError{Code: "UNAUTHENTICATED", Message: "Authentication required"}
    ErrAccountNotFound    = &AppError{Code: "ACCOUNT_NOT_FOUND", Message: "Account not found"}
    ErrTransferNotFound   = &AppError{Code: "TRANSFER_NOT_FOUND", Message: "Transfer not found"}
    ErrValidation         = &AppError{Code: "VALIDATION_ERROR", Message: "Invalid input"}
    ErrInsufficientFunds  = &AppError{Code: "INSUFFICIENT_FUNDS", Message: "Insufficient funds"}
    ErrIdempotency        = &AppError{Code: "IDEMPOTENCY_KEY_EXISTS", Message: "Idempotency key already used"}
    ErrInternal           = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
    ErrSameAccount        = &AppError{Code: "SAME_ACCOUNT", Message: "Source and destination are the same account"}
    ErrInvalidAmount      = &AppError{Code: "INVALID_AMOUNT", Message: "Amount must be a positive integer"}
    ErrServiceUnavailable = &AppError{Code: "SERVICE_UNAVAILABLE", Message: "Service temporarily unavailable"}
)
```

## Error Response Envelope

```go
// internal/api/response.go
package api

// ErrorPayload is the canonical error object, nested under "error".
type ErrorPayload struct {
    Code      string            `json:"code"`
    Message   string            `json:"message"`
    Details   map[string]string `json:"details,omitempty"`
    RequestID string            `json:"requestId"`
}

// ErrorResponse is the single error envelope: {"error": {...}}.
type ErrorResponse struct {
    Error ErrorPayload `json:"error"`
}

func respondError(w http.ResponseWriter, err error, requestID string) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        writeJSON(w, appErr.HTTPStatus(), ErrorResponse{
            Error: ErrorPayload{
                Code:      appErr.Code,
                Message:   appErr.Message,
                Details:   appErr.Details,
                RequestID: requestID,
            },
        })
        return
    }
    // Unknown error - log and return generic
    logger.Error("unhandled error", "err", err, "requestId", requestID)
    writeJSON(w, http.StatusInternalServerError, ErrorResponse{
        Error: ErrorPayload{
            Code:      "INTERNAL_ERROR",
            Message:   "An unexpected error occurred",
            RequestID: requestID,
        },
    })
}
```

### Example error responses

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Password must be at least 12 characters",
    "details": {
      "password": "min length 12, requires uppercase, lowercase, number, symbol"
    },
    "requestId": "req_550e8400-e29b-41d4-a716-446655440000"
  }
}
```

```json
{
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Insufficient funds for this transfer",
    "details": {
      "availableCents": "1000",
      "requestedCents": "5000"
    },
    "requestId": "req_550e8400-e29b-41d4-a716-446655440001"
  }
}
```

## Rules

- **All errors use the same nested envelope** (`error.code`, `error.message`, `error.details`, `error.requestId`) exactly as defined in [`docs/06-api-specification.md`](../../06-api-specification.md#5-error-response-format).
- **Every error `code` is canonical and maps to one HTTP status** (for example `INSUFFICIENT_FUNDS` → `400`, `EMAIL_ALREADY_REGISTERED` → `409`, `SERVICE_UNAVAILABLE` → `503`); never invent codes outside the documented set.
- **`404` is returned for unauthorised access** to another account's resource (never `403`, to avoid information leakage).
- **Never return stack traces, CQL, or credentials** in error responses.
- **Errors are wrapped with `%w`** for unwrapping with `errors.As` and `errors.Is`.
- **Log errors server-side** with correlation ID; return generic messages to clients.
- **Every error code is stable and documented** — clients can rely on them.
- **Error messages are actionable** and written in plain language.

## Error Handling Patterns

### Handler pattern

```go
func (h *TransferHandler) CreateTransfer(c *gin.Context) {
    var req CreateTransferRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        writeGinError(c, errors.ErrValidation, getRequestID(c), map[string]string{"body": err.Error()})
        return
    }

    ctx := c.Request.Context()
    transfer, err := h.transferService.InitiateTransfer(ctx, req, c.GetHeader("Idempotency-Key"))
    if err != nil {
        writeGinError(c, err, getRequestID(c))
        return
    }

    c.JSON(http.StatusAccepted, transfer)
}
```

### Service pattern

```go
func (s *TransferService) InitiateTransfer(ctx context.Context, req CreateTransferRequest, idempotencyKey string) (*TransferResponse, error) {
    if req.Amount <= 0 {
        return nil, fmt.Errorf("validate amount: %w", errors.ErrInvalidAmount)
    }

    account, err := s.accountRepo.GetByUserID(ctx, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("get account: %w", err)
    }

    if account.BalanceCents < req.Amount {
        return nil, fmt.Errorf("check funds: %w", errors.ErrInsufficientFunds)
    }

    // ... proceed with transfer
}
```

## Cross-references

- Go best practices index: [`go.md`](./go.md)
- Project structure: [`structure.md`](./structure.md)
- API specification: [`docs/06-api-specification.md`](../../06-api-specification.md#5-error-response-format)
- Security: [`docs/04-architecture.md`](../../04-architecture.md#9-security-boundaries)
- Failure recovery: [`docs/04-architecture.md`](../../04-architecture.md#8-failure-recovery)
- Planned test: `test_error_handling` (every error code has a corresponding test)