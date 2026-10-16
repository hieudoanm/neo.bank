package main

import (
	"log"
	"net/http"

	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/api"
	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/api-gateway/internal/core/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	var keyManager *auth.KeyManager
	if len(cfg.JWTPublicKeyPEM) > 0 {
		keyManager, err = auth.NewKeyManager(cfg.JWTPublicKeyPEM, cfg.JWTKeyID)
		if err != nil {
			log.Fatalf("jwt public key: %v", err)
		}
	} else {
		log.Println("WARNING: JWT_PUBLIC_KEY_PEM not set; protected routes will reject all tokens")
		keyManager, err = auth.NewKeyManager([]byte{}, cfg.JWTKeyID)
		if err != nil {
			log.Fatalf("jwt public key: %v", err)
		}
	}

	authMiddleware := middleware.NewAuthMiddleware(keyManager)
	router := api.NewRouter(authMiddleware, cfg.AuthServiceURL, cfg.AccountServiceURL)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("api-gateway listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
