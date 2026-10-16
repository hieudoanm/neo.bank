package main

import (
	"log"
	"net/http"

	"github.com/apache/cassandra-gocql-driver/v2"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/api"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/auth"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/core/config"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/repositories"
	"github.com/hieudoanm/neo.bank/packages/backend/auth/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	var keyManager *auth.KeyManager
	if len(cfg.JWTPrivateKeyPEM) > 0 && len(cfg.JWTPublicKeyPEM) > 0 {
		keyManager, err = auth.NewKeyManager(cfg.JWTPrivateKeyPEM, cfg.JWTPublicKeyPEM, cfg.JWTKeyID)
		if err != nil {
			log.Fatalf("jwt keys: %v", err)
		}
	} else {
		log.Println("WARNING: using ephemeral RSA key pair (set JWT_PRIVATE_KEY_PEM/JWT_PUBLIC_KEY_PEM for production)")
		keyManager, err = auth.GenerateDevKeyPair(cfg.JWTKeyID)
		if err != nil {
			log.Fatalf("generate dev keys: %v", err)
		}
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

	userRepo := repositories.NewUserRepository(session)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(session)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, keyManager)
	authHandler := handlers.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(keyManager)

	router := api.NewRouter(authHandler, authMiddleware)

	port := cfg.Port
	log.Printf("auth listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
