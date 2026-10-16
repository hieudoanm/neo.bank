package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api/middleware"
)

func NewRouter(accountHandler *handlers.AccountHandler, authMiddleware *middleware.AuthMiddleware) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", HealthzHandler)
	r.GET("/readyz", ReadyzHandler)

	v1 := r.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth)
	{
		v1.GET("/accounts/:accountId/balance", accountHandler.GetBalance)
	}

	return r
}
