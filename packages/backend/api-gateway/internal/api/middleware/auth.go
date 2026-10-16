package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/core/errors"
)

type AuthMiddleware struct {
	keyManager *auth.KeyManager
}

func NewAuthMiddleware(keyManager *auth.KeyManager) *AuthMiddleware {
	return &AuthMiddleware{keyManager: keyManager}
}

func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
	tokenStr := extractBearerToken(c.Request)
	if tokenStr == "" {
		writeAuthError(c, errors.ErrInvalidCredentials.Code, "authentication required", http.StatusUnauthorized)
		c.Abort()
		return
	}
	claims, err := m.keyManager.Validate(tokenStr)
	if err != nil {
		code := errors.ErrInvalidCredentials.Code
		if strings.Contains(err.Error(), "expired") {
			code = errors.ErrTokenExpired.Code
		}
		writeAuthError(c, code, "access token expired or invalid", http.StatusUnauthorized)
		c.Abort()
		return
	}
	c.Set("userId", claims.Subject)
	c.Set("accountId", claims.AccountID)
	c.Set("email", claims.Email)
	c.Next()
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(prefix):])
}

func writeAuthError(c *gin.Context, code, message string, status int) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":      code,
			"message":   message,
			"requestId": getRequestID(c),
		},
	})
}

func getRequestID(c *gin.Context) string {
	if rid := c.GetHeader("X-Request-ID"); rid != "" {
		return rid
	}
	return ""
}
