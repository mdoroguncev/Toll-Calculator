package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mdoroguncev/toll-calculator/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/types"
)

type KafkaConsumer struct {
	consumer    *kafka.Consumer
	isRunning   bool
	calcService CalculatorServicer
	aggClient   client.Client
}

func NewKafkaConsumer(brokers, topic, group string, svc CalculatorServicer, aggClient client.Client) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          group,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		return nil, err
	}

	c.SubscribeTopics([]string{topic}, nil)
	// A signal handler or similar could be used to set this to false to break the loop.

	return &KafkaConsumer{
		consumer:    c,
		calcService: svc,
		aggClient:   aggClient,
	}, nil
}

func (c *KafkaConsumer) Close() {
	c.isRunning = false
}

func (c *KafkaConsumer) Start() {
	slog.Info("kafka transport started")
	c.isRunning = true
	c.readMessageLoop()
}

func (c *KafkaConsumer) readMessageLoop() {
	for c.isRunning {
		msg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			slog.Error("kafka consume error", "err", err)
			continue
		}
		var data types.OBUData
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.Error("JSON serialization error", "err", err)
			continue
		}
		distance, err := c.calcService.CalculateDistance(data)
		if err != nil {
			slog.Error("calculation error", "err", err)
			continue
		}
		req := &types.AggregateRequest{
			Value: distance,
			Unix:  int64(time.Now().UnixNano()),
			ObuID: int32(data.OBUID),
		}
		if err := c.aggClient.Aggregate(context.Background(), req); err != nil {
			slog.Error("aggregate error", "err", err)
			continue
		}
	}
}
