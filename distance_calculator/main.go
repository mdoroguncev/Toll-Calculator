package main

import (
	"log"

	"github.com/mdoroguncev/toll-calculator/aggregator/client"
)

// type DistanceCalculator struct {
// 	consumer DataConsumer
// }

const (
	kafkaTopic         = "obuData"
	aggregatorEndpoint = "http://127.0.0.1:3000"
)

func main() {
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
