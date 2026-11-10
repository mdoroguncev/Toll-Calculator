package client

import (
	"context"
	"errors"

	"github.com/mdoroguncev/toll-calculator/internal/types"
)

var ErrNotFound = errors.New("not found")

type Client interface {
	Aggregate(context.Context, types.Distance) error
	GetInvoice(context.Context, int) (*types.Invoice, error)
}
