package main

import (
	"log"
	"net/http"

	"github.com/apache/cassandra-gocql-driver/v2"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/api"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/config"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/producer"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/repositories"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/services"
)

func main() {
	cfg := config.Load()

	keyManager, err := auth.NewKeyManager([]byte(cfg.JWTPrivateKeyPEM), []byte(cfg.JWTPublicKeyPEM), cfg.JWTKeyID)
	if err != nil {
		log.Fatalf("jwt keys: %v", err)
	}

	cluster := gocql.NewCluster(cfg.CassandraContactPoints...)
	cluster.Keyspace = cfg.CassandraKeyspace
	cluster.Timeout = cfg.CassandraTimeout
	cluster.NumConns = 4
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("connect cassandra: %v", err)
	}
	defer session.Close()

	userEvents := producer.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicTransfers)
	defer userEvents.Close()

	userRepo := repositories.NewUserRepository(session)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(session)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, keyManager, userEvents)
	authHandler := handlers.NewAuthHandler(authService)

	router := api.NewRouter(authHandler)

	port := cfg.Port
	log.Printf("auth listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
