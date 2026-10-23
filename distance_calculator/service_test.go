package main

import (
	"math"
	"testing"

	"github.com/mdoroguncev/toll-calculator/types"
)

func TestHaversine(t *testing.T) {
	tests := []struct {
		name string
		a, b Point
		want float64
	}{
		{name: "one degree of longitude on equator", a: Point{0, 0}, b: Point{0, 1}, want: 111.19},
		{name: "Moscow to SPB", a: Point{Lat: 55.7558, Long: 37.6173}, b: Point{Lat: 59.9343, Long: 30.3351}, want: 633.0},
		{name: "The same point", a: Point{Lat: 55.7558, Long: 37.6173}, b: Point{Lat: 55.7558, Long: 37.6173}, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := haversine(tt.a, tt.b)
			if math.Abs(got-tt.want) > 0.5 {
				t.Errorf("haversine(%v, %v) = %.2f km, want %.2f km", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCalculateDistance_MultipleOBUs(t *testing.T) {
	svc := NewCalcService()

	got, err := svc.CalculateDistance(types.OBUData{OBUID: 1, Lat: 0, Long: 0})
	if err != nil {
		t.Fatalf("OBU 1, step 1: %v", err)
	}
	if got != 0 {
		t.Fatalf("OBU 1, step 1: want 0, got %v", got)
	}

	got, err = svc.CalculateDistance(types.OBUData{OBUID: 2, Lat: 50, Long: 50})
	if err != nil {
		t.Fatalf("OBU 2, step 1: %v", err)
	}
	if got != 0 {
		t.Fatalf("OBU 2, step 1: want 0, got %v", got)
	}

	got, err = svc.CalculateDistance(types.OBUData{OBUID: 1, Lat: 0, Long: 1})
	if err != nil {
		t.Fatalf("OBU 1, step 2: %v", err)
	}
	if math.Abs(got-111.19) > 0.1 {
		t.Errorf("OBU 1, step 2: want ~111.19 km, got %v km", got)
	}
}
