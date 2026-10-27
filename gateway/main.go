package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/mdoroguncev/toll-calculator/aggregator/client"
	"github.com/sirupsen/logrus"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func main() {
	listenAddr := flag.String("listenAddr", ":6000", "the listen addres of the http server")
	aggregatorServiceAddr := flag.String("aggServiceAddr", "http://localhost:4000", "the listen addres of the aggregator service")
	flag.Parse()

	var (
		client     = client.NewHTTPClient(*aggregatorServiceAddr) //endpoint for the aggregator service
		invHandler = NewInvoiceHandler(client)
	)
	http.HandleFunc("/invoice", makeAPIFunc(invHandler.handleGetInvoice))
	logrus.Infof("gateway http is running on port %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))

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

	inv, err := h.client.GetInvoice(context.Background(), 1927769697)
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
		defer func(start time.Time) {
			logrus.WithFields(logrus.Fields{
				"took": time.Since(start),
				"uri":  r.RequestURI,
			}).Info("REQ: ")
		}(time.Now())
		if err := fn(w, r); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
}
