package main

import (
	"log"
	"net/http"
	"time"

	"github.com/hieudoanm/neo.bank/works/api-gateway/internal/api"
	"github.com/hieudoanm/neo.bank/works/api-gateway/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/works/api-gateway/internal/core/auth"
	"github.com/hieudoanm/neo.bank/works/api-gateway/internal/core/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	keyManager, err := auth.NewKeyManager(cfg.JWTPublicKeyPEM, cfg.JWTKeyID)
	if err != nil {
		log.Fatalf("jwt public key: %v", err)
	}

	authMiddleware := middleware.NewAuthMiddleware(keyManager)
	router := api.NewRouter(authMiddleware, cfg.AuthServiceURL, cfg.AccountServiceURL, cfg.TransferServiceURL, cfg.TransactionServiceURL, middleware.RateLimits{
		AuthPerMinute: cfg.RateLimitAuthPerMin,
		UserPerMinute: cfg.RateLimitUserPerMin,
		LoginFailures: cfg.LoginLockoutFailures,
		LoginWindow:   cfg.LoginLockoutWindow,
	}, cfg.CORSOrigins)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("api-gateway listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
