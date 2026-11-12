package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hieudoanm/neo.bank/works/accounts/internal/consumer"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/core/config"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/db"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/producer"
	"github.com/hieudoanm/neo.bank/works/accounts/internal/repositories"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.LoadConsumer()

	session, err := db.NewSession(cfg.CassandraContactPoints, cfg.CassandraKeyspace)
	if err != nil {
		logger.Error("connect cassandra", "err", err)
		os.Exit(1)
	}
	defer session.Close()

	accountProducer := producer.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicTransfers)
	defer accountProducer.Close()

	repo := repositories.NewAccountRepository(session)
	c := consumer.NewAccountConsumer(cfg.KafkaBrokers, cfg.KafkaTopicTransfers, cfg.KafkaConsumerGroup, cfg.KafkaTopicTransfersDLQ, cfg.InitialBalanceCents, cfg.CurrencyCode, repo, accountProducer, logger)
	defer c.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("account consumer started", "topic", cfg.KafkaTopicTransfers, "group", cfg.KafkaConsumerGroup)
	if err := c.Run(ctx); err != nil {
		logger.Error("consumer stopped", "err", err)
		os.Exit(1)
	}
}
