package aggregator

import (
	"log/slog"

	"github.com/mdoroguncev/toll-calculator/internal/types"
)

const basePrice = 3.15

type Service interface {
	AggregateDistance(types.Distance) error
	CalculateInvoice(int) (*types.Invoice, error)
}

type Storer interface {
	Insert(types.Distance) error
	Get(int) (float64, error)
}

type InvoiceAggregator struct {
	store Storer
}

func NewService(store Storer) Service {
	return &InvoiceAggregator{
		store: store,
	}
}

func (i *InvoiceAggregator) AggregateDistance(dist types.Distance) error {
	slog.Info("aggregating distance", "obuID", dist.OBUID, "distance", dist.Value, "unix", dist.Unix)
	return i.store.Insert(dist)
}

func (i *InvoiceAggregator) CalculateInvoice(obuID int) (*types.Invoice, error) {
	dist, err := i.store.Get(obuID)
	if err != nil {
		return nil, err
	}
	inv := &types.Invoice{
		OBUID:         obuID,
		TotalDistance: dist,
		TotalAmount:   basePrice * dist,
	}
	return inv, nil
}
