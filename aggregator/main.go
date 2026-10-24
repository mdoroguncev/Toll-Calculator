package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mdoroguncev/toll-calculator/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	var (
		store          = makeStore()
		svc            = NewInvoiceAggregator(store)
		GRPCListenAddr = os.Getenv("AGG_GRPC_ENDPOINT")
		HTTPListenAddr = os.Getenv("AGG_HTTP_ENDPOINT")
	)
	svc = NewMetricsMiddleware(svc)
	svc = NewLogMiddleware(svc)
	go func() {
		log.Fatal(makeGRPCTransport(GRPCListenAddr, svc))
	}()
	log.Fatal(makeHTTPTransport(HTTPListenAddr, svc))
}

func makeGRPCTransport(listenAddr string, svc Aggregator) error {
	fmt.Println("GRPC running on port ", listenAddr)
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

	fmt.Println("HTTP running on port ", listenAddr)
	return http.ListenAndServe(listenAddr, nil)
}

func makeStore() Storer {
	storeType := os.Getenv("AGG_STORE_TYPE")
	switch storeType {
	case "memory":
		return NewMemoryStore()
	default:
		log.Fatalf("invalid store type %s", storeType)
		return nil
	}

}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
