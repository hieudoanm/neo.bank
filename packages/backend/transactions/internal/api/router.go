package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/api/middleware"
)

func NewRouter(transactionHandler *handlers.TransactionHandler, authMiddleware *middleware.AuthMiddleware) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", HealthzHandler)
	r.GET("/readyz", ReadyzHandler)

	v1 := r.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth)
	{
		v1.GET("/accounts/:accountId/transactions", transactionHandler.ListTransactions)
	}

	return r
}
