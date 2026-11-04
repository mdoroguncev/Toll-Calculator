package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/mdoroguncev/toll-calculator/internal/config"
	"github.com/mdoroguncev/toll-calculator/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "aggregator")
	slog.SetDefault(logger)
	var (
		store          = makeStore(config.Get("AGG_STORE_TYPE", "memory"))
		svc            = NewInvoiceAggregator(store)
		GRPCListenAddr = config.Get("AGG_GRPC_ENDPOINT", ":3001")
		HTTPListenAddr = config.Get("AGG_HTTP_ENDPOINT", ":4000")
	)
	svc = NewMetricsMiddleware(svc)
	svc = NewLogMiddleware(svc)
	go func() {
		log.Fatal(makeGRPCTransport(GRPCListenAddr, svc))
	}()
	log.Fatal(makeHTTPTransport(HTTPListenAddr, svc))
}

func makeGRPCTransport(listenAddr string, svc Aggregator) error {
	slog.Info("grpc server started", "addr", listenAddr)
	// make a tcp listener
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	//make a new grpc native server with (options)
	server := grpc.NewServer([]grpc.ServerOption{}...)
	//register (our) grpc server implementation to the grpc package
	types.RegisterAggregatorServer(server, NewAggregatorGRPCServer(svc))
	return server.Serve(ln)
}

func makeHTTPTransport(listenAddr string, svc Aggregator) error {
	var (
		aggMetricHandler = newHTTPMetricHandler("aggregate")
		invMetricHandler = newHTTPMetricHandler("invoice")
		aggregateHandler = makeHTTPHandlerFunc(aggMetricHandler.instrument(handleAggregate(svc)))
		invoiceHandler   = makeHTTPHandlerFunc(invMetricHandler.instrument(handleGetInvoice(svc)))
	)

	http.HandleFunc("/invoice", invoiceHandler)
	http.HandleFunc("/aggregate", aggregateHandler)
	http.Handle("/metrics", promhttp.Handler())

	slog.Info("http server started", "addr", listenAddr)
	return http.ListenAndServe(listenAddr, nil)
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

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
