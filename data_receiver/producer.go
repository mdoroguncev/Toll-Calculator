package main

import (
	"encoding/json"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mdoroguncev/toll-calculator/types"
)

type DataProducer interface {
	ProduceData(types.OBUData) error
	Close()
}

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaProducer(brokers, topic string) (DataProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return nil, err
	}
	// strat another goroutine to check if we have delivered the data.
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					// fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					// fmt.Printf("Delivred massage to %v\n", ev.TopicPartition)
				}
			}
		}
	}()
	return &KafkaProducer{
		producer: p,
		topic:    topic,
	}, nil
}

func (p *KafkaProducer) ProduceData(data types.OBUData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},

		Value: b,
	}, nil)
}

func (p *KafkaProducer) Close() {
	msgCount := p.producer.Flush(5000)
	if msgCount > 0 {
		slog.Warn("unsended messages count", "msgCount", msgCount)
	}

	p.producer.Close()
}
