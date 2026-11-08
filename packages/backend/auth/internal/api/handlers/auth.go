package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	applierrors "github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/errors"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/models"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/services"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*services.RegisterResult, error)
	Login(ctx context.Context, email, password string) (*services.RegisterResult, error)
	Refresh(ctx context.Context, refreshToken string) (*services.RefreshResult, error)
	Logout(ctx context.Context, refreshToken string) error
}

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// maxRequestBodyBytes bounds JSON request bodies; auth payloads are a few
// hundred bytes, so 1 MiB is generous while preventing memory exhaustion.
const maxRequestBodyBytes = 1 << 20

func decodeJSON(c *gin.Context, v any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	// Reject trailing data: a second JSON value or garbage after the first
	// document means the body is not a single well-formed request.
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("unexpected trailing data")
		}
		return err
	}
	return nil
}

func writeJSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

func writeError(c *gin.Context, err error, requestID string) {
	var appErr *applierrors.AppError
	if errors.As(err, &appErr) {
		payload := gin.H{
			"code":      appErr.Code,
			"message":   appErr.Message,
			"requestId": requestID,
		}
		if len(appErr.Details) > 0 {
			payload["details"] = appErr.Details
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": payload})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":      "INTERNAL_ERROR",
			"message":   "an unexpected error occurred",
			"requestId": requestID,
		},
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	requestID := getRequestID(c)
	var req models.RegisterRequest
	if err := decodeJSON(c, &req); err != nil {
		writeError(c, applierrors.ValidationError(map[string]string{"body": "invalid request body"}), requestID)
		return
	}
	result, err := h.service.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeError(c, err, requestID)
		return
	}
	writeJSON(c, http.StatusCreated, models.TokenResponse{
		UserID:       result.UserID,
		AccountID:    result.AccountID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    services.AccessTokenExpiresIn(),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	requestID := getRequestID(c)
	var req models.LoginRequest
	if err := decodeJSON(c, &req); err != nil {
		writeError(c, applierrors.ValidationError(map[string]string{"body": "invalid request body"}), requestID)
		return
	}
	result, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeError(c, err, requestID)
		return
	}
	writeJSON(c, http.StatusOK, models.TokenResponse{
		UserID:       result.UserID,
		AccountID:    result.AccountID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    services.AccessTokenExpiresIn(),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	requestID := getRequestID(c)
	var req models.RefreshRequest
	if err := decodeJSON(c, &req); err != nil {
		writeError(c, applierrors.ValidationError(map[string]string{"body": "invalid request body"}), requestID)
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeError(c, err, requestID)
		return
	}
	writeJSON(c, http.StatusOK, models.TokenResponse{
		UserID:       result.UserID,
		AccountID:    result.AccountID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    services.AccessTokenExpiresIn(),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	requestID := getRequestID(c)
	var req models.LogoutRequest
	if err := decodeJSON(c, &req); err != nil {
		writeError(c, applierrors.ValidationError(map[string]string{"body": "invalid request body"}), requestID)
		return
	}
	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		writeError(c, err, requestID)
		return
	}
	c.Status(http.StatusNoContent)
}

func getRequestID(c *gin.Context) string {
	if rid := c.GetHeader("X-Request-ID"); rid != "" {
		return rid
	}
	return ""
}
