package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdoroguncev/toll-calculator/internal/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/internal/calculator"
	"github.com/mdoroguncev/toll-calculator/internal/config"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "distance_calculator")
	slog.SetDefault(logger)

	var (
		kafkaBrokers = config.Get("KAFKA_BROKERS", "localhost:9092")
		kafkaTopic   = config.Get("KAFKA_TOPIC", "obuData")
		kafkaGroup   = config.Get("CALC_KAFKA_GROUP", "distance-calculator")
	)
	var (
		err error
		svc calculator.Service
	)
	svc = calculator.NewService()
	svc = calculator.NewLogMiddleware(svc)

	grpcClient, err := client.NewGRPCClient(config.Get("AGG_GRPC_URL", "127.0.0.1:3001"))
	if err != nil {
		log.Fatal(err)
	}

	kafkaConsumer, err := calculator.NewConsumer(kafkaBrokers, kafkaTopic, kafkaGroup, svc, grpcClient)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	kafkaConsumer.Start(ctx)
	if err := kafkaConsumer.Close(); err != nil {
		slog.Error("failed to close kafka consumer", "err", err)
	}
	if err = grpcClient.Close(); err != nil {
		slog.Error("failed to close grpc client", "err", err)
	}
	slog.Info("calculator stopped")
}
