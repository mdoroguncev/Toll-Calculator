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

type Store interface {
	Insert(types.Distance) error
	Get(int) (float64, error)
}

type service struct {
	store Store
}

func NewService(store Store) Service {
	return &service{
		store: store,
	}
}

func (s *service) AggregateDistance(dist types.Distance) error {
	slog.Info("aggregating distance", "obuID", dist.OBUID, "distance", dist.Value, "unix", dist.Unix)
	return s.store.Insert(dist)
}

func (s *service) CalculateInvoice(obuID int) (*types.Invoice, error) {
	dist, err := s.store.Get(obuID)
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
