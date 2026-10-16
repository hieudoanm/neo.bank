package main

import (
	"log"
	"net/http"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/core/config"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/repositories"
	"github.com/hieudoanm/neo.bank/packages/backend/accounts/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	km, err := auth.NewKeyManager(cfg.JWTPublicKeyPEM, cfg.JWTKeyID)
	if err != nil {
		log.Fatalf("jwt key: %v", err)
	}
	cluster := gocql.NewCluster(cfg.CassandraContactPoints...)
	cluster.Keyspace = cfg.CassandraKeyspace
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("cassandra: %v", err)
	}
	defer session.Close()

	repo := repositories.NewAccountRepository(session)
	svc := services.NewAccountService(repo)
	handler := handlers.NewAccountHandler(svc)
	mw := middleware.NewAuthMiddleware(km)
	router := api.NewRouter(handler, mw)

	log.Printf("accounts listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
