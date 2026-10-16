package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api/middleware"
)

func NewRouter(transferHandler *handlers.TransferHandler, authMiddleware *middleware.AuthMiddleware) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", HealthzHandler)
	r.GET("/readyz", ReadyzHandler)

	v1 := r.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth)
	{
		v1.POST("/transfers", transferHandler.CreateTransfer)
		v1.GET("/transfers/:transferId", transferHandler.GetTransfer)
	}

	return r
}
