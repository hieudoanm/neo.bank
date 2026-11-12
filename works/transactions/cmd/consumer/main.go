package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hieudoanm/neo.bank/works/transactions/internal/consumer"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/core/config"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/db"
	"github.com/hieudoanm/neo.bank/works/transactions/internal/repositories"
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

	store := repositories.NewTransactionRepository(session)
	c := consumer.NewProjectionConsumer(cfg.KafkaBrokers, cfg.KafkaTopicTransfers, cfg.KafkaConsumerGroup, cfg.KafkaTopicTransfersDLQ, store, logger)
	defer c.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("transaction projection consumer started", "topic", cfg.KafkaTopicTransfers, "group", cfg.KafkaConsumerGroup)
	if err := c.Run(ctx); err != nil {
		logger.Error("consumer stopped", "err", err)
		os.Exit(1)
	}
}
