package aggregator

import (
	"context"
	"log/slog"
	"net"

	"github.com/mdoroguncev/toll-calculator/internal/types"
	"google.golang.org/grpc"
)

type grpcAggregatorServer struct {
	types.UnimplementedAggregatorServer
	svc Service
}

func newAggregatorGRPCServer(svc Service) *grpcAggregatorServer {
	return &grpcAggregatorServer{
		svc: svc,
	}
}

func (s *grpcAggregatorServer) Aggregate(ctx context.Context, req *types.AggregateRequest) (*types.None, error) {
	distance := types.Distance{
		OBUID: int(req.ObuID),
		Value: req.Value,
		Unix:  int(req.Unix),
	}
	if err := s.svc.AggregateDistance(distance); err != nil {
		return nil, err
	}
	return &types.None{}, nil
}

func NewGRPCServer(listenAddr string, svc Service) (*grpc.Server, net.Listener, error) {
	slog.Info("grpc server created", "addr", listenAddr)
	// make a tcp listener
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, nil, err
	}
	//make a new grpc native server with (options)
	server := grpc.NewServer([]grpc.ServerOption{}...)
	//register (our) grpc server implementation to the grpc package
	types.RegisterAggregatorServer(server, newAggregatorGRPCServer(svc))
	return server, ln, nil
}
