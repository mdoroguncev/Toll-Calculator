package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/config"
	"github.com/mdoroguncev/toll-calculator/internal/receiver"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "data_receiver")
	slog.SetDefault(logger)

	var (
		listenAddr   = config.Get("RECEIVER_ADDR", ":30000")
		kafkaBrokers = config.Get("KAFKA_BROKERS", "localhost:9092")
		kafkaTopic   = config.Get("KAFKA_TOPIC", "obuData")
	)
	recv, err := receiver.NewDataReceiver(kafkaBrokers, kafkaTopic)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", recv.HandleWS)
	slog.Info("receiver started", "addr", listenAddr, "kafka", kafkaBrokers, "topic", kafkaTopic)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: listenAddr, Handler: mux}
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down server")
	case err := <-errCh:
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown the server", "err", err)
	}

	recv.Close()
	slog.Info("receiver stopped")
}
