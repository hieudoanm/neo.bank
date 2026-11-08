package api

import (
	"log/slog"
	"os"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api/middleware"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func NewRouter(transferHandler *handlers.TransferHandler, authMiddleware *middleware.AuthMiddleware, session *gocql.Session) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestIDMiddleware())
	r.Use(loggerMiddleware())

	r.GET("/healthz", HealthzHandler)
	r.GET("/readyz", ReadyzHandler(session))

	v1 := r.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth)
	{
		v1.POST("/transfers", transferHandler.CreateTransfer)
		v1.GET("/transfers/:transferId", transferHandler.GetTransfer)
	}

	return r
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set("requestId", rid)
		c.Request.Header.Set("X-Request-ID", rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"durationMs", time.Since(start).Milliseconds(),
			"requestId", c.GetString("requestId"),
		)
	}
}
