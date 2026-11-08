package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/config"
	"github.com/mdoroguncev/toll-calculator/internal/httpx"
	"github.com/mdoroguncev/toll-calculator/internal/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
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
		store          = makeStore(config.Get("AGG_STORE_TYPE", "memory"))
		svc            = NewInvoiceAggregator(store)
		GRPCListenAddr = config.Get("AGG_GRPC_ENDPOINT", ":3001")
		HTTPListenAddr = config.Get("AGG_HTTP_ENDPOINT", ":4000")
	)
	svc = NewMetricsMiddleware(svc)
	svc = NewLogMiddleware(svc)

	grpcSrv, ln, err := makeGRPCServer(GRPCListenAddr, svc)
	if err != nil {
		slog.Error("grpc listen failed", "err", err)
		os.Exit(1)
	}
	httpSrv := makeHTTPServer(HTTPListenAddr, svc)

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

}

func makeGRPCServer(listenAddr string, svc Aggregator) (*grpc.Server, net.Listener, error) {
	slog.Info("grpc server created", "addr", listenAddr)
	// make a tcp listener
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, nil, err
	}
	//make a new grpc native server with (options)
	server := grpc.NewServer([]grpc.ServerOption{}...)
	//register (our) grpc server implementation to the grpc package
	types.RegisterAggregatorServer(server, NewAggregatorGRPCServer(svc))
	return server, ln, nil
}

func makeHTTPServer(listenAddr string, svc Aggregator) *http.Server {
	var (
		aggMetricHandler = newHTTPMetricHandler("aggregate")
		invMetricHandler = newHTTPMetricHandler("invoice")
		aggregateHandler = httpx.Handle(aggMetricHandler.instrument(handleAggregate(svc)))
		invoiceHandler   = httpx.Handle(invMetricHandler.instrument(handleGetInvoice(svc)))
		mux              = http.NewServeMux()
	)

	mux.HandleFunc("/invoice", invoiceHandler)
	mux.HandleFunc("/aggregate", aggregateHandler)
	mux.Handle("/metrics", promhttp.Handler())

	slog.Info("http server created", "addr", listenAddr)
	return &http.Server{Addr: listenAddr, Handler: mux}
}

func makeStore(storeType string) Storer {
	switch storeType {
	case "memory":
		return NewMemoryStore()
	default:
		log.Fatalf("invalid store type %s", storeType)
		return nil
	}

}
