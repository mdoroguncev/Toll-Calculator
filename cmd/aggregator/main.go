package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/aggregator"
	"github.com/mdoroguncev/toll-calculator/internal/config"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		slog.Error("failed to load .env file", "err", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "aggregator")
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		GRPCListenAddr = config.Get("AGG_GRPC_ENDPOINT", ":3001")
		HTTPListenAddr = config.Get("AGG_HTTP_ENDPOINT", ":4000")
	)
	storeType := config.Get("AGG_STORE_TYPE", "memory")
	dsn := config.Get("AGG_DB_DSN", "")
	connectCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	store, err := aggregator.NewStore(connectCtx, storeType, dsn)
	if err != nil {
		slog.Error("failed to make store", "err", err)
		os.Exit(1)
	}

	svc := aggregator.NewService(store)
	svc = aggregator.NewMetricsMiddleware(svc)
	svc = aggregator.NewLogMiddleware(svc)

	grpcSrv, ln, err := aggregator.NewGRPCServer(GRPCListenAddr, svc)
	if err != nil {
		slog.Error("grpc listen failed", "err", err)
		os.Exit(1)
	}
	httpSrv := aggregator.NewHTTPServer(HTTPListenAddr, svc)

	errCh := make(chan error, 2)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		if err := grpcSrv.Serve(ln); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down the server")
	case err = <-errCh:
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown the server", "err", err)
	}
	grpcSrv.GracefulStop()

	store.Close()
}
