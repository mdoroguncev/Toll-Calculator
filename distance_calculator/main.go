package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/mdoroguncev/toll-calculator/aggregator/client"
)

// type DistanceCalculator struct {
// 	consumer DataConsumer
// }

const (
	kafkaTopic         = "obuData"
	aggregatorEndpoint = "http://127.0.0.1:4000"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "distance_calculator")
	slog.SetDefault(logger)
	var (
		err error
		svc CalculatorServicer
	)
	svc = NewCalcService()
	svc = NewLogMiddleware(svc)

	htppClient := client.NewHTTPClient(aggregatorEndpoint)
	//grpcClient, err := client.NewGRPCClient(aggregatorEndpoint)
	// if err != nil {
	// log.Fatal(err)
	// }

	kafkaConsumer, err := NewKafkaConsumer(kafkaTopic, svc, htppClient)
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer.Start()
}
