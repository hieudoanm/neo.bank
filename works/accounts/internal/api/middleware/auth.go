package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/core/auth"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/core/errors"
)

type AuthMiddleware struct {
	keyManager *auth.KeyManager
}

func NewAuthMiddleware(keyManager *auth.KeyManager) *AuthMiddleware {
	return &AuthMiddleware{keyManager: keyManager}
}

func (m *AuthMiddleware) RequireAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": errors.ErrUnauthorized.Code, "message": "unauthorized", "requestId": requestID(c)}})
		c.Abort()
		return
	}
	token := strings.TrimSpace(authHeader[len(prefix):])
	claims, err := m.keyManager.Validate(token)
	if err != nil {
		code := errors.ErrUnauthorized.Code
		if strings.Contains(err.Error(), "expired") {
			code = errors.ErrTokenExpired.Code
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": code, "message": "access token expired or invalid", "requestId": requestID(c)}})
		c.Abort()
		return
	}
	c.Set("userId", claims.Subject)
	c.Set("accountId", claims.AccountID)
	c.Set("email", claims.Email)
	c.Next()
}

func requestID(c *gin.Context) string {
	if rid := c.GetHeader("X-Request-ID"); rid != "" {
		return rid
	}
	return ""
}
