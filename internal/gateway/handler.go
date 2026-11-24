package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/internal/httpx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	invoiceRequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_invoice_requests_total",
	}, []string{"code"})
	invoiceRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gateway_invoice_request_duration_seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"code"})
)

type responseRecoder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecoder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecoder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		code := strconv.Itoa(rec.statusCode)

		invoiceRequestCounter.WithLabelValues(code).Inc()
		invoiceRequestDuration.WithLabelValues(code).Observe(duration)
	})
}

func NewRouter(c client.Client) http.Handler {
	invHandler := newInvoiceHandler(c)

	mux := http.NewServeMux()

	invoiceHandlerWithMetrics := metricsMiddleware(httpx.Handle(invHandler.handleGetInvoice))
	mux.Handle("GET /invoice/{obuID}", invoiceHandlerWithMetrics)
	mux.Handle("/metrics", promhttp.Handler())

	return httpx.Logging(mux)
}

type invoiceHandler struct {
	client client.Client
}

func newInvoiceHandler(client client.Client) *invoiceHandler {
	return &invoiceHandler{
		client: client,
	}
}

func (h *invoiceHandler) handleGetInvoice(w http.ResponseWriter, r *http.Request) error {
	intID, err := strconv.Atoi(r.PathValue("obuID"))
	if err != nil {
		return httpx.APIError{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid obu id %q", r.PathValue("obuID"))}
	}

	inv, err := h.client.GetInvoice(r.Context(), intID)
	if errors.Is(err, client.ErrNotFound) {
		return httpx.APIError{Code: http.StatusNotFound, Err: err}
	}
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, inv)
}
