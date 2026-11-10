package client

import (
	"context"

	aggregatorv1 "github.com/mdoroguncev/toll-calculator/gen/aggregator/v1"
	"github.com/mdoroguncev/toll-calculator/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	client aggregatorv1.AggregatorServiceClient
	conn   *grpc.ClientConn
}

var _ Client = (*GRPCClient)(nil)

func NewGRPCClient(endpoint string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := aggregatorv1.NewAggregatorServiceClient(conn)
	return &GRPCClient{
		client: c,
		conn:   conn,
	}, nil
}

func (c *GRPCClient) Aggregate(ctx context.Context, dist types.Distance) error {
	req := &aggregatorv1.AggregateRequest{ObuId: int32(dist.OBUID), Value: dist.Value, Unix: int64(dist.Unix)}

	_, err := c.client.Aggregate(ctx, req)
	return err
}

func (c *GRPCClient) GetInvoice(ctx context.Context, id int) (*types.Invoice, error) {
	resp, err := c.client.GetInvoice(ctx, &aggregatorv1.GetInvoiceRequest{ObuId: int32(id)})
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	inv := &types.Invoice{
		OBUID:         int(resp.GetObuId()),
		TotalDistance: resp.GetTotalDistance(),
		TotalAmount:   resp.GetTotalAmount(),
	}

	return inv, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}
