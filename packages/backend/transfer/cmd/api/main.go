package main

import (
	"log"
	"net/http"
	"time"

	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/config"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/db"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/producer"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/repositories"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/services"
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

	transferProducer := producer.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicTransfers)
	defer transferProducer.Close()

	repo := repositories.NewTransferRepository(session)
	transferService := services.NewTransferService(repo, transferProducer)
	transferHandler := handlers.NewTransferHandler(transferService)
	authMiddleware := middleware.NewAuthMiddleware(keyManager)

	router := api.NewRouter(transferHandler, authMiddleware, session)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("transfer listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
