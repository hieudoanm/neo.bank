package api

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/api/middleware"
)

func NewRouter(authMiddleware *middleware.AuthMiddleware, authServiceURL string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestIDMiddleware())
	r.Use(corsMiddleware())
	r.Use(loggerMiddleware())

	r.GET("/healthz", healthHandler)
	r.GET("/readyz", healthHandler)

	authProxy, err := newReverseProxy(authServiceURL)
	if err != nil {
		panic(err)
	}

	public := r.Group("/api/v1")
	{
		public.POST("/register", gin.WrapH(authProxy))
		public.POST("/login", gin.WrapH(authProxy))
		public.POST("/refresh", gin.WrapH(authProxy))
		public.POST("/logout", gin.WrapH(authProxy))
	}

	protected := r.Group("/api/v1")
	protected.Use(authMiddleware.RequireAuth)
	{
		protected.GET("/accounts/me", meHandler)
	}

	return r
}

func newReverseProxy(target string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, `{"error":{"code":"SERVICE_UNAVAILABLE","message":"upstream service unavailable","requestId":""}}`)
	}
	return proxy, nil
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func meHandler(c *gin.Context) {
	userID, _ := c.Get("userId")
	accountID, _ := c.Get("accountId")
	email, _ := c.Get("email")
	c.JSON(http.StatusOK, gin.H{
		"userId":    userID,
		"accountId": accountID,
		"email":     email,
	})
}

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

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
