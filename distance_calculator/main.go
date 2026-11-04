package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/mdoroguncev/toll-calculator/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/internal/config"
)

// type DistanceCalculator struct {
// 	consumer DataConsumer
// }

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "distance_calculator")
	slog.SetDefault(logger)

	var (
		kafkaBrokers       = config.Get("KAFKA_BROKERS", "localhost:9092")
		kafkaTopic         = config.Get("KAFKA_TOPIC", "obuData")
		kafkaGroup         = config.Get("CALC_KAFKA_GROUP", "distance-calculator")
		aggregatorEndpoint = config.Get("AGG_HTTP_URL", "http://127.0.0.1:4000")
	)
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

	kafkaConsumer, err := NewKafkaConsumer(kafkaBrokers, kafkaTopic, kafkaGroup, svc, htppClient)
	if err != nil {
		log.Fatal(err)
	}
	kafkaConsumer.Start()
}
