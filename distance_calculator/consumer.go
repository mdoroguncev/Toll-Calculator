package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mdoroguncev/toll-calculator/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/internal/types"
)

type KafkaConsumer struct {
	consumer    *kafka.Consumer
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

	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer:    c,
		calcService: svc,
		aggClient:   aggClient,
	}, nil
}

func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	slog.Info("kafka transport started")
	c.readMessageLoop(ctx)
}

func (c *KafkaConsumer) readMessageLoop(ctx context.Context) {
	for ctx.Err() == nil {
		msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
		if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
			continue
		}
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
		if err := c.aggClient.Aggregate(ctx, req); err != nil {
			slog.Error("aggregate error", "err", err)
			continue
		}
	}
}
