package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mdoroguncev/toll-calculator/internal/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/internal/httpx"
)

func NewRouter(c client.Client) http.Handler {
	invHandler := newInvoiceHandler(c)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /invoice/{obuID}", httpx.Handle(invHandler.handleGetInvoice))

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
