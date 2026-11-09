package aggregator

import (
	"context"
	"errors"
	"log/slog"
	"net"

	aggregatorv1 "github.com/mdoroguncev/toll-calculator/gen/aggregator/v1"
	"github.com/mdoroguncev/toll-calculator/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type grpcAggregatorServer struct {
	aggregatorv1.UnimplementedAggregatorServiceServer
	svc Service
}

func newAggregatorGRPCServer(svc Service) *grpcAggregatorServer {
	return &grpcAggregatorServer{
		svc: svc,
	}
}

func (s *grpcAggregatorServer) Aggregate(ctx context.Context, req *aggregatorv1.AggregateRequest) (*aggregatorv1.AggregateResponse, error) {
	distance := types.Distance{
		OBUID: int(req.GetObuId()),
		Value: req.GetValue(),
		Unix:  int(req.GetUnix()),
	}
	if err := s.svc.AggregateDistance(distance); err != nil {
		return nil, err
	}
	return &aggregatorv1.AggregateResponse{}, nil
}

func (s *grpcAggregatorServer) GetInvoice(ctx context.Context, req *aggregatorv1.GetInvoiceRequest) (*aggregatorv1.GetInvoiceResponse, error) {
	invoice, err := s.svc.CalculateInvoice(int(req.GetObuId()))
	if errors.Is(err, ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &aggregatorv1.GetInvoiceResponse{ObuId: req.GetObuId(), TotalDistance: invoice.TotalDistance, TotalAmount: invoice.TotalAmount}, nil
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
	aggregatorv1.RegisterAggregatorServiceServer(server, newAggregatorGRPCServer(svc))
	reflection.Register(server)
	return server, ln, nil
}
