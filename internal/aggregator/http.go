package aggregator

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/httpx"
	"github.com/mdoroguncev/toll-calculator/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type httpMetricHandler struct {
	reqCounter prometheus.Counter
	errCounter prometheus.Counter
	reqLatency prometheus.Histogram
}

func newHTTPMetricHandler(reqName string) *httpMetricHandler {
	reqCounter := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: fmt.Sprintf("http_%s_%s", reqName, "request_counter"),
		Name:      "aggregator",
	})
	errCounter := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: fmt.Sprintf("http_%s_%s", reqName, "err_counter"),
		Name:      "aggregator",
	})

	reqLatency := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: fmt.Sprintf("http_%s_%s", reqName, "request_latency"),
		Name:      "aggregator",
		Buckets:   []float64{0.1, 0.5, 1},
	})
	return &httpMetricHandler{
		reqCounter: reqCounter,
		reqLatency: reqLatency,
		errCounter: errCounter,
	}
}

func (h *httpMetricHandler) instrument(next httpx.HandlerFunc) httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var err error
		defer func(start time.Time) {
			latency := time.Since(start).Seconds()
			slog.Info("http request", "latency", latency, "request", r.RequestURI, "err", err)
			h.reqLatency.Observe(latency)
			h.reqCounter.Inc()
			if err != nil {
				h.errCounter.Inc()
			}
		}(time.Now())
		err = next(w, r)
		return err
	}
}

func handleGetInvoice(svc Service) httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != "GET" {
			return httpx.APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("invalid http method %s", r.Method),
			}
		}

		values, ok := r.URL.Query()["obu"]
		if !ok {
			return httpx.APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("missing obu id"),
			}
		}

		obuID, err := strconv.Atoi(values[0])
		if err != nil {
			return httpx.APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("invalid obu id %s", values[0]),
			}
		}

		invoice, err := svc.CalculateInvoice(r.Context(), obuID)
		if errors.Is(err, ErrNotFound) {
			return httpx.APIError{
				Code: http.StatusNotFound,
				Err:  err,
			}
		} else if err != nil {
			return httpx.APIError{
				Code: http.StatusInternalServerError,
				Err:  err,
			}
		}
		return httpx.WriteJSON(w, http.StatusOK, invoice)
	}
}

func handleAggregate(svc Service) httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			return httpx.APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("method not supported %s", r.Method),
			}
		}

		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			return httpx.APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to decode the response body: %s", err),
			}
		}

		if err := svc.AggregateDistance(r.Context(), distance); err != nil {
			return httpx.APIError{
				Code: http.StatusInternalServerError,
				Err:  err,
			}
		}
		return httpx.WriteJSON(w, http.StatusOK, map[string]string{"msg": "ok"})
	}
}

func NewHTTPServer(listenAddr string, svc Service) *http.Server {
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
