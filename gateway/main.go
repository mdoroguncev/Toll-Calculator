package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mdoroguncev/toll-calculator/aggregator/client"
	"github.com/sirupsen/logrus"
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
	listenAddr := flag.String("listenAddr", ":6000", "the listen addres of the http server")
	aggregatorServiceAddr := flag.String("aggServiceAddr", "http://localhost:4000", "the listen addres of the aggregator service")
	flag.Parse()

	var (
		aggClient = client.NewHTTPClient(*aggregatorServiceAddr) //endpoint for the aggregator service
	)
	logrus.Infof("gateway http is running on port %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, newRouter(aggClient)))

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
			logrus.WithFields(logrus.Fields{
				"took": time.Since(start),
				"uri":  r.RequestURI,
				"err":  err,
			}).Info("REQ: ")
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
