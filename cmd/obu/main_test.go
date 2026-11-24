package main

import (
	"math"
	"testing"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/calculator"
	"github.com/mdoroguncev/toll-calculator/internal/types"
)

func TestDistanceAcrossTurn(t *testing.T) {
	v := vehicle{obuID: 1, offsetMeters: routeMeters - 20, speedKmh: 60, direction: 1}
	svc := calculator.NewService()
	for _, elapsed := range []time.Duration{0, 5 * time.Second, 5 * time.Second, 3 * time.Minute} {
		var meters float64
		for _, p := range v.advance(elapsed) {
			km, err := svc.CalculateDistance(types.OBUData{OBUID: v.obuID, Lat: p.Lat, Long: p.Long})
			if err != nil {
				t.Fatal(err)
			}
			meters += km * 1000
		}
		want := v.speedKmh / 3.6 * elapsed.Seconds()
		if math.Abs(meters-want) > 0.001 {
			t.Fatalf("after %v: got %.3f meters, want %.3f", elapsed, meters, want)
		}
	}
}
