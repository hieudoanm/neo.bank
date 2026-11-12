package main

import (
	"log"
	"net/http"
	"time"

	"github.com/apache/cassandra-gocql-driver/v2"
	"github.com/hieudoanm/neo.bank/works/auth/internal/api"
	"github.com/hieudoanm/neo.bank/works/auth/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/works/auth/internal/core/auth"
	"github.com/hieudoanm/neo.bank/works/auth/internal/core/config"
	"github.com/hieudoanm/neo.bank/works/auth/internal/producer"
	"github.com/hieudoanm/neo.bank/works/auth/internal/repositories"
	"github.com/hieudoanm/neo.bank/works/auth/internal/services"
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

	router := api.NewRouter(authHandler, session)

	port := cfg.Port
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("auth listening on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
