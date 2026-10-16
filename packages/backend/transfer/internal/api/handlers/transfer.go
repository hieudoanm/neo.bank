package handlers

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	applierrors "github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/models"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/services"
)

var idempotencyKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,255}$`)

type TransferHandler struct {
	service *services.TransferService
}

func NewTransferHandler(service *services.TransferService) *TransferHandler {
	return &TransferHandler{service: service}
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

func (h *TransferHandler) CreateTransfer(c *gin.Context) {
	reqID := requestID(c)
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if !idempotencyKeyPattern.MatchString(idempotencyKey) {
		writeError(c, applierrors.ValidationError(map[string]string{"Idempotency-Key": "required; must match ^[a-zA-Z0-9_-]{1,255}$"}), reqID)
		return
	}

	var body models.CreateTransferRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, applierrors.ValidationError(map[string]string{"body": "invalid body"}), reqID)
		return
	}

	callerAccountID, _ := c.Get("accountId")
	callerStr, _ := callerAccountID.(string)
	resp, err := h.service.InitiateTransfer(c.Request.Context(), callerStr, body, idempotencyKey)
	if err != nil {
		writeError(c, err, reqID)
		return
	}
	c.Header("Idempotency-Key", idempotencyKey)
	c.JSON(http.StatusAccepted, resp)
}

func (h *TransferHandler) GetTransfer(c *gin.Context) {
	reqID := requestID(c)
	callerAccountID, _ := c.Get("accountId")
	callerStr, _ := callerAccountID.(string)
	resp, err := h.service.GetTransfer(c.Request.Context(), c.Param("transferId"), callerStr)
	if err != nil {
		writeError(c, err, reqID)
		return
	}
	c.JSON(http.StatusOK, resp)
}
