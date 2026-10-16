package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	applierrors "github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/services"
)

type TransferHandler struct {
	service *services.TransferService
}

func NewTransferHandler(service *services.TransferService) *TransferHandler {
	return &TransferHandler{service: service}
}

func requestID(c *gin.Context) string { if r := c.GetHeader("X-Request-ID"); r != "" { return r }; return "" }

func writeError(c *gin.Context, err error, reqID string) {
	var appErr *applierrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message, "requestId": reqID}})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "an unexpected error occurred", "requestId": reqID}})
}

func (h *TransferHandler) CreateTransfer(c *gin.Context) {
	reqID := requestID(c)
	var body models.CreateTransferRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, applierrors.ValidationError(map[string]string{"body": "invalid body"}), reqID)
		return
	}
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}
	amount := body.AmountCents
	srcAcct, _ := c.Get("accountId")
	srcAcctStr, _ := srcAcct.(string)
	resp, err := h.service.CreateTransfer(c.Request.Context(), "", srcAcctStr, body.DestinationAccountId, amount, idempotencyKey)
	if err != nil {
		writeError(c, err, reqID)
		return
	}
	c.JSON(http.StatusAccepted, resp)
}
