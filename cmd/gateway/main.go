package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/internal/config"
	"github.com/mdoroguncev/toll-calculator/internal/gateway"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "gateway")
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Flags override environment variables, environment variables override the defaults.
	aggregatorServiceAddr := flag.String("aggServiceAddr", config.Get("AGG_HTTP_URL", "http://localhost:4000"), "the listen addres of the aggregator service")
	listenAddr := flag.String("listenAddr", config.Get("GATEWAY_ADDR", ":6000"), "the listen addres of the http server")

	flag.Parse()

	aggClient := client.NewHTTPClient(*aggregatorServiceAddr) //endpoint for the aggregator service
	handler := gateway.NewRouter(aggClient)
	srv := &http.Server{Addr: *listenAddr, Handler: handler}
	slog.Info("gateway http is running on port", "addr", *listenAddr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown the server", "err", err)
	}
}
