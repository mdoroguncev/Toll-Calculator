package calculator

import (
	"log/slog"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/types"
)

type LogMiddleware struct {
	next Service
}

func NewLogMiddleware(next Service) Service {
	return &LogMiddleware{
		next: next,
	}
}

func (m *LogMiddleware) CalculateDistance(data types.OBUData) (dist float64, err error) {
	defer func(start time.Time) {
		slog.Info("calculate distance",
			"took", time.Since(start),
			"err", err,
			"dist", dist,
		)
	}(time.Now())

	dist, err = m.next.CalculateDistance(data)
	return
}
