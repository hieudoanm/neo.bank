package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api/middleware"
)

func NewRouter(accountHandler *handlers.AccountHandler, authMiddleware *middleware.AuthMiddleware) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestIDMiddleware())
	r.Use(corsMiddleware())
	r.Use(loggerMiddleware())
	r.GET("/healthz", healthHandler)
	r.GET("/readyz", healthHandler)

	protected := r.Group("/api/v1")
	protected.Use(authMiddleware.RequireAuth)
	{
		protected.GET("/accounts/:accountId/balance", accountHandler.GetBalance)
	}
	return r
}

func healthHandler(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) }
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set("requestId", rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
func loggerMiddleware() gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }
