package main

import (
	"log"
	"net/http"

	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/api"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/core/config"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/db"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/repositories"
	"github.com/hieudoanm/neo.bank/packages/backend/transactions/internal/services"
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

	repo := repositories.NewTransactionRepository(session)
	transactionService := services.NewTransactionService(repo)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	authMiddleware := middleware.NewAuthMiddleware(keyManager)

	router := api.NewRouter(transactionHandler, authMiddleware, session)

	log.Printf("transactions listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
