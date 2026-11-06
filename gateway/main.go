package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mdoroguncev/toll-calculator/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/internal/config"
)

type APIError struct {
	Code int
	Err  error
}

func (e APIError) Error() string {
	return e.Err.Error()
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

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
	handler := newRouter(aggClient)
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

func newRouter(c client.Client) http.Handler {
	invHandler := NewInvoiceHandler(c)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /invoice/{obuID}", makeAPIFunc(invHandler.handleGetInvoice))

	return mux
}

type InvoiceHandler struct {
	client client.Client
}

func NewInvoiceHandler(client client.Client) *InvoiceHandler {
	return &InvoiceHandler{
		client: client,
	}
}

func (h *InvoiceHandler) handleGetInvoice(w http.ResponseWriter, r *http.Request) error {
	intID, err := strconv.Atoi(r.PathValue("obuID"))
	if err != nil {
		return APIError{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid obu id %q", r.PathValue("obuID"))}
	}

	inv, err := h.client.GetInvoice(r.Context(), intID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, inv)
}

func writeJSON(w http.ResponseWriter, code int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func makeAPIFunc(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func(start time.Time) {
			slog.Info("REQ:", "took", time.Since(start), "uri", r.RequestURI, "err", err)
		}(time.Now())
		if err = fn(w, r); err != nil {
			if apiErr, ok := err.(APIError); ok {
				writeJSON(w, apiErr.Code, map[string]string{"error": apiErr.Error()})
			} else {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}
	}
}
