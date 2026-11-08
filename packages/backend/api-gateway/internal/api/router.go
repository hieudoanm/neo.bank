package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/api/middleware"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func NewRouter(authMiddleware *middleware.AuthMiddleware, authServiceURL, accountServiceURL, transferServiceURL, transactionServiceURL string, limits middleware.RateLimits, corsOrigins []string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestIDMiddleware())
	r.Use(corsMiddleware(corsOrigins))
	r.Use(loggerMiddleware())

	r.GET("/healthz", healthHandler)
	r.GET("/readyz", healthHandler)

	authProxy, err := newReverseProxy(authServiceURL)
	if err != nil {
		panic(err)
	}
	accountProxy, err := newReverseProxy(accountServiceURL)
	if err != nil {
		panic(err)
	}
	transferProxy, err := newReverseProxy(transferServiceURL)
	if err != nil {
		panic(err)
	}
	transactionProxy, err := newReverseProxy(transactionServiceURL)
	if err != nil {
		panic(err)
	}

	public := r.Group("/api/v1")
	if limits.AuthPerMinute > 0 {
		authLimiter := middleware.NewRateLimiter(limits.AuthPerMinute, time.Minute)
		public.Use(middleware.RateLimitMiddleware(authLimiter, middleware.ClientIPKey))
	}
	{
		public.POST("/register", gin.WrapH(authProxy))

		loginHandlers := make([]gin.HandlerFunc, 0, 2)
		if limits.LoginFailures > 0 {
			tracker := middleware.NewFailureTracker(limits.LoginFailures, limits.LoginWindow)
			loginHandlers = append(loginHandlers, middleware.LoginLockoutMiddleware(tracker))
		}
		loginHandlers = append(loginHandlers, gin.WrapH(authProxy))
		public.POST("/login", loginHandlers...)

		public.POST("/refresh", gin.WrapH(authProxy))
		public.POST("/logout", gin.WrapH(authProxy))
	}

	protected := r.Group("/api/v1")
	protected.Use(authMiddleware.RequireAuth)
	if limits.UserPerMinute > 0 {
		userLimiter := middleware.NewRateLimiter(limits.UserPerMinute, time.Minute)
		protected.Use(middleware.RateLimitMiddleware(userLimiter, middleware.AuthenticatedUserKey))
	}
	{
		protected.GET("/accounts/:accountId/balance", gin.WrapH(accountProxy))
		protected.GET("/accounts/:accountId/transactions", gin.WrapH(transactionProxy))
		protected.POST("/transfers", gin.WrapH(transferProxy))
		protected.GET("/transfers/:transferId", gin.WrapH(transferProxy))
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
		logger.Error("upstream request failed", "err", err, "path", r.URL.Path, "requestId", r.Header.Get("X-Request-ID"))
		body, _ := json.Marshal(gin.H{
			"error": gin.H{
				"code":      "SERVICE_UNAVAILABLE",
				"message":   "upstream service unavailable",
				"requestId": r.Header.Get("X-Request-ID"),
			},
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write(body)
	}
	return proxy, nil
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

func corsMiddleware(origins []string) gin.HandlerFunc {
	allowAll := false
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		if o == "*" {
			allowAll = true
			continue
		}
		allowed[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		switch {
		case allowAll:
			c.Header("Access-Control-Allow-Origin", "*")
		default:
			if _, ok := allowed[origin]; ok && origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
