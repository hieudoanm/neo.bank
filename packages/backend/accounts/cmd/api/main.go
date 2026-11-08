package main

import (
	"log"
	"net/http"

	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/config"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/db"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/repositories"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/services"
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

	session, err := db.NewSession(cfg.CassandraContactPoints, cfg.CassandraKeyspace)
	if err != nil {
		log.Fatalf("connect cassandra: %v", err)
	}
	defer session.Close()

	repo := repositories.NewAccountRepository(session)
	accountService := services.NewAccountService(repo)
	accountHandler := handlers.NewAccountHandler(accountService)
	authMiddleware := middleware.NewAuthMiddleware(keyManager)

	router := api.NewRouter(accountHandler, authMiddleware, session)

	log.Printf("accounts listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
