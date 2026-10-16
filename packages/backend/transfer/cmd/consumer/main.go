package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/consumer"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/core/config"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/db"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/producer"
	"github.com/hieudoanm/neo.bank/packages/backend/transfer/internal/repositories"
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

	transferProducer := producer.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicTransfers)
	defer transferProducer.Close()

	repo := repositories.NewTransferRepository(session)
	c := consumer.NewSagaConsumer(cfg.KafkaBrokers, cfg.KafkaTopicTransfers, cfg.KafkaConsumerGroup, cfg.KafkaTopicTransfersDLQ, repo, transferProducer, logger)
	defer c.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("transfer saga consumer started", "topic", cfg.KafkaTopicTransfers, "group", cfg.KafkaConsumerGroup)
	if err := c.Run(ctx); err != nil {
		logger.Error("consumer stopped", "err", err)
		os.Exit(1)
	}
}
