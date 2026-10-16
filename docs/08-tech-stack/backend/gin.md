# NEO Bank — Gin Best Practices

_Backend · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md)._

## 1. Purpose

This document defines how Gin is used in NEO Bank: router setup, middleware, request binding, validation, error responses, and authentication. It is the **how**; the **why** (trade-offs vs Chi/Echo/Fiber/net/http) lives in [`§5.1`](../../08-tech-stack.md#51-go--gin).

Gin is the HTTP layer only. Business logic never lives in handlers; handlers bind input, call services, and write JSON responses.

## 2. Router Setup

```go
// internal/api/server.go
func NewRouter(cfg *config.Config, deps *dependencies) *gin.Engine {
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(structuredLogger(deps.logger))
    r.Use(requestIDMiddleware())
    r.Use(cors.New(cors.Config{
        AllowOrigins:     cfg.CORSOrigins,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Idempotency-Key", "X-Request-ID"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
    r.GET("/healthz", healthHandler(deps))
    r.GET("/readyz", readinessHandler(deps))

    auth := r.Group("/api/v1")
    {
        auth.POST("/register", deps.authHandler.Register)
        auth.POST("/login", deps.authHandler.Login)
        auth.POST("/refresh", deps.authHandler.Refresh)
        auth.POST("/logout", deps.authHandler.Logout)
    }

    api := r.Group("/api/v1")
    api.Use(deps.authMiddleware.RequireAuth)
    {
        api.GET("/accounts/:accountId/balance", deps.accountHandler.GetBalance)
        api.GET("/accounts/:accountId/transactions", deps.transactionHandler.ListTransactions)
        api.POST("/transfers", deps.transferHandler.CreateTransfer)
        api.GET("/transfers/:transferId", deps.transferHandler.GetTransfer)
    }
    return r
}
```

**Rules:**
- One `gin.Engine` per process; routes registered in `NewRouter`.
- Health endpoints are outside the auth group.
- Authenticated routes share `RequireAuth` middleware; no per-route auth duplication.
- Canonical paths: `GET /api/v1/accounts/{accountId}/balance`, `GET /api/v1/accounts/{accountId}/transactions`, `POST /api/v1/transfers`, `GET /api/v1/transfers/{transferId}`. No `/accounts/me` alias; the caller supplies the `accountId` from the login response / JWT claim, and the service verifies ownership (404 on mismatch).
- `gin.New()` (not `gin.Default()`) — we control logging and recovery explicitly.

## 3. Handler Pattern

Handlers are thin: bind input, delegate to services, write responses.

```go
func (h *TransferHandler) CreateTransfer(c *gin.Context) {
    requestID := getRequestID(c)
    var req CreateTransferRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", "Invalid request body", err.Error(), requestID))
        return
    }
    idempotencyKey := c.GetHeader("Idempotency-Key")
    ctx := c.Request.Context()
    transfer, err := h.transferService.InitiateTransfer(ctx, req, idempotencyKey)
    if err != nil {
        writeGinError(c, err, requestID)
        return
    }
    c.JSON(http.StatusAccepted, transfer)
}

func (h *TransferHandler) GetTransfer(c *gin.Context) {
    requestID := getRequestID(c)
    transferID := c.Param("transferId")
    ctx := c.Request.Context()
    transfer, err := h.transferService.GetTransfer(ctx, transferID)
    if err != nil {
        writeGinError(c, err, requestID)
        return
    }
    c.JSON(http.StatusOK, transfer)
}
```

**Rules:**
- Use `c.ShouldBindJSON` — handle validation explicitly.
- Path params via `c.Param`; query via `c.Query`; headers via `c.GetHeader`.
- Never call `c.Abort()` in handlers — return after writing response.
- Handlers know nothing about Cassandra or Kafka directly; services own that.

## 4. Request Binding & Validation

```go
type CreateTransferRequest struct {
    DestinationAccountID string `json:"destinationAccountId" binding:"required,uuid"`
    AmountCents          int64  `json:"amountCents" binding:"required,gt=0,lte=9999999999"`
}
```

**Rules:**
- Use Gin's `binding` tags for basic validation; complex rules in service layer.
- Idempotency key read from `Idempotency-Key` header, never from body.

## 5. Error Handling

```go
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

func writeGinError(c *gin.Context, err error, requestID string) {
    var appErr *core.AppError
    if errors.As(err, &appErr) {
        c.JSON(appErr.HTTPStatus(), ErrorResponse{
            Error: ErrorPayload{
                Code: appErr.Code, Message: appErr.Message,
                Details: appErr.Details, RequestID: requestID,
            },
        })
        return
    }
    core.Logger.ErrorContext(c.Request.Context(), "unhandled error", "err", err, "requestId", requestID)
    c.JSON(http.StatusInternalServerError, ErrorResponse{
        Error: ErrorPayload{
            Code: "INTERNAL_ERROR", Message: "An unexpected error occurred", RequestID: requestID,
        },
    })
}
```

**Rules:**
- All errors use the same nested envelope (`error.code`, `error.message`, `error.details`, `error.requestId`) — see [`error-handling.md`](./error-handling.md) and [`docs/06-api-specification.md`](../../06-api-specification.md#5-error-response-format).
- JSON body field names are camelCase (`destinationAccountId`, `amountCents`); structured log fields are camelCase too (`requestId`, `correlationId`, `latencyMs`).
- `404` for unauthorised access to another account's resource (never `403`).
- Never return stack traces, CQL, or credentials.

## 6. Authentication Middleware

```go
func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
    tokenStr := extractBearerToken(c.Request)
    if tokenStr == "" {
        c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHENTICATED", "Authentication required", "", getRequestID(c)))
        c.Abort()
        return
    }
    claims, err := m.jwtValidator.Validate(tokenStr)
    if err != nil {
        c.JSON(http.StatusUnauthorized, errorResponse("TOKEN_EXPIRED", "Access token expired or invalid", "", getRequestID(c)))
        c.Abort()
        return
    }
    c.Set("userId", claims.Subject)
    c.Set("accountId", claims.AccountID)
    c.Next()
}
```

**Rules:**
- `c.Set`/`c.Get` for per-request values (user_id, account_id).
- `c.Abort()` stops the chain after writing error.
- Account ownership verified for every account-scoped endpoint.

## 7. Idempotency Header

The `Idempotency-Key` header is **required** on `POST /api/v1/transfers` and optional on other mutating endpoints. It is read from the header (never the body) and passed to the service layer, which checks `idempotency_keys` in Cassandra. A duplicate key with an identical payload returns the cached response; the same key with a different payload returns `409 IDEMPOTENCY_KEY_EXISTS`.

## 8. Testing with Gin

```go
gin.SetMode(gin.TestMode)
w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)
c.Request, _ = http.NewRequest(method, path, strings.NewReader(body))
handler(c)
```

**Rules:**
- `gin.SetMode(gin.TestMode)` disables console output.
- `gin.CreateTestContext` + `httptest.NewRecorder` for handler tests.
- Table-driven tests for valid/invalid request bodies.
- Service layer mocked; handler tests verify binding and status codes only.

## 9. Code Style

- `*gin.Context` — never `http.ResponseWriter`/`*http.Request` in handlers.
- `c.ShouldBindJSON` for body binding; validate explicitly after.
- `c.Param` / `c.Query` / `c.GetHeader` — never parse `c.Request.URL` manually.
- `c.JSON` / `c.AbortWithStatusJSON` for all responses.
- No `c.AbortWithError` — use `writeGinError` for consistent envelopes.
- `c.Set` / `c.Get` for request-scoped values only.
- Grouped routes by auth requirement; no per-route auth duplication.

---
↑ Part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md).