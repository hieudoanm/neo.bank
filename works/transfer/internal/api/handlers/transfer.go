package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	applierrors "github.com/hieudoanm/neo.bank/works/transfer/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/transfer/internal/models"
)

var idempotencyKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,255}$`)

// maxRequestBodyBytes bounds JSON request bodies; transfer payloads are a few
// hundred bytes, so 1 MiB is generous while preventing memory exhaustion.
const maxRequestBodyBytes = 1 << 20

func decodeJSON(c *gin.Context, v any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("unexpected trailing data")
		}
		return err
	}
	return nil
}

// TransferService initiates and reads transfers for an authenticated account.
type TransferService interface {
	InitiateTransfer(ctx context.Context, sourceAccountID string, req models.CreateTransferRequest, idempotencyKey string) (*models.TransferResponse, error)
	GetTransfer(ctx context.Context, transferID, callerAccountID string) (*models.TransferResponse, error)
}

type TransferHandler struct {
	service TransferService
}

func NewTransferHandler(service TransferService) *TransferHandler {
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
	if err := decodeJSON(c, &body); err != nil {
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
