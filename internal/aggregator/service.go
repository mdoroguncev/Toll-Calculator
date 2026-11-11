package aggregator

import (
	"context"
	"log/slog"

	"github.com/mdoroguncev/toll-calculator/internal/types"
)

const basePrice = 3.15

type Service interface {
	AggregateDistance(context.Context, types.Distance) error
	CalculateInvoice(context.Context, int) (*types.Invoice, error)
}

type Store interface {
	Insert(context.Context, types.Distance) error
	Get(context.Context, int) (float64, error)
	Close()
}

type service struct {
	store Store
}

func NewService(store Store) Service {
	return &service{
		store: store,
	}
}

func (s *service) AggregateDistance(ctx context.Context, dist types.Distance) error {
	slog.Info("aggregating distance", "obuID", dist.OBUID, "distance", dist.Value, "unix", dist.Unix)
	return s.store.Insert(ctx, dist)
}

func (s *service) CalculateInvoice(ctx context.Context, obuID int) (*types.Invoice, error) {
	dist, err := s.store.Get(ctx, obuID)
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
