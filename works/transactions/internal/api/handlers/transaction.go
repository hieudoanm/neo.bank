package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	applierrors "github.com/hieudoanm/neo.bank/works/transactions/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/models"
)

// TransactionReader lists transaction history for an account.
type TransactionReader interface {
	ListTransactions(ctx context.Context, accountID uuid.UUID, callerAccountID, cursor string, limit int) (*models.TransactionListResponse, error)
}

type TransactionHandler struct {
	service TransactionReader
}

func NewTransactionHandler(service TransactionReader) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func requestID(c *gin.Context) string {
	return c.GetHeader("X-Request-ID")
}

func writeError(c *gin.Context, err error, reqID string) {
	var appErr *applierrors.AppError
	if errors.As(err, &appErr) {
		body := gin.H{"code": appErr.Code, "message": appErr.Message, "requestId": reqID}
		if len(appErr.Details) > 0 {
			body["details"] = appErr.Details
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": body})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "an unexpected error occurred", "requestId": reqID}})
}

func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	reqID := requestID(c)
	accountID, err := uuid.Parse(c.Param("accountId"))
	if err != nil {
		writeError(c, applierrors.ErrAccountNotFound, reqID)
		return
	}

	limit := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(c, applierrors.ErrInvalidLimit, reqID)
			return
		}
		limit = parsed
	}

	callerAccountID, _ := c.Get("accountId")
	callerStr, _ := callerAccountID.(string)
	resp, err := h.service.ListTransactions(c.Request.Context(), accountID, callerStr, c.Query("cursor"), limit)
	if err != nil {
		writeError(c, err, reqID)
		return
	}
	c.JSON(http.StatusOK, resp)
}
