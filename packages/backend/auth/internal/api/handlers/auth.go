package handlers

import (
	"context"
	"encoding/json"
	"errors"
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

func decodeJSON(c *gin.Context, v any) error {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
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
	writeJSON(c, http.StatusOK, gin.H{
		"accessToken":  result.AccessToken,
		"refreshToken": result.RefreshToken,
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
