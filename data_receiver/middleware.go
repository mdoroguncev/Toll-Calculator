package main

import (
	"log/slog"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/types"
)

type LogMiddleware struct {
	next DataProducer
}

func NewLogMiddleware(next DataProducer) *LogMiddleware {
	return &LogMiddleware{
		next: next,
	}
}

func (l *LogMiddleware) ProduceData(data types.OBUData) error {
	defer func(start time.Time) {
		slog.Info("producing to kafka",
			"obuID", data.OBUID,
			"lat", data.Lat,
			"long", data.Long,
			"took", time.Since(start),
		)
	}(time.Now())
	return l.next.ProduceData(data)
}

func (l *LogMiddleware) Close() {
	l.next.Close()
}
