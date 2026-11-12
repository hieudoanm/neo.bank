package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	applierrors "github.com/hieudoanm/neo.bank/works/accounts/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/models"
)

// BalanceService fetches an account balance for its owner.
type BalanceService interface {
	FetchBalance(ctx context.Context, accountID uuid.UUID, callerAccountID string) (*models.BalanceResponse, error)
}

type AccountHandler struct {
	service BalanceService
}

func NewAccountHandler(service BalanceService) *AccountHandler {
	return &AccountHandler{service: service}
}

func writeError(c *gin.Context, err error, requestID string) {
	var appErr *applierrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, gin.H{
			"error": gin.H{"code": appErr.Code, "message": appErr.Message, "requestId": requestID},
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{"code": "INTERNAL_ERROR", "message": "an unexpected error occurred", "requestId": requestID},
	})
}

func (h *AccountHandler) GetBalance(c *gin.Context) {
	requestID := requestID(c)
	accountIDStr := c.Param("accountId")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		writeError(c, applierrors.ErrAccountNotFound, requestID)
		return
	}
	caller, _ := c.Get("accountId")
	callerStr, _ := caller.(string)
	resp, err := h.service.FetchBalance(c.Request.Context(), accountID, callerStr)
	if err != nil {
		writeError(c, err, requestID)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func requestID(c *gin.Context) string {
	if rid := c.GetHeader("X-Request-ID"); rid != "" {
		return rid
	}
	return ""
}
