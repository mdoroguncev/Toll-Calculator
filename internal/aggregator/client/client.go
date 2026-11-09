package client

import (
	"context"

	"github.com/mdoroguncev/toll-calculator/internal/types"
)

type Client interface {
	Aggregate(context.Context, types.Distance) error
	GetInvoice(context.Context, int) (*types.Invoice, error)
}
