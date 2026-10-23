package main

import (
	"math"

	"github.com/mdoroguncev/toll-calculator/types"
)

type CalculatorServicer interface {
	CalculateDistance(types.OBUData) (float64, error)
}

type Point struct {
	Lat  float64
	Long float64
}

type CalcService struct {
	prevPoint map[int]Point
}

func NewCalcService() CalculatorServicer {
	return &CalcService{
		prevPoint: make(map[int]Point),
	}
}

func (s *CalcService) CalculateDistance(data types.OBUData) (float64, error) {
	cur := Point{Lat: data.Lat, Long: data.Long}

	prev, ok := s.prevPoint[data.OBUID]

	s.prevPoint[data.OBUID] = cur

	if !ok {
		return 0.0, nil
	}

	distance := haversine(prev, cur)

	return distance, nil
}

func haversine(a, b Point) float64 {
	const earthRadiusKm = 6371.0

	// Переводим градусы в радианы
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLon := (b.Long - a.Long) * math.Pi / 180

	// Формула гаверсинуса
	sinLat := math.Sin(dLat / 2)
	sinLon := math.Sin(dLon / 2)

	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLon*sinLon

	// Угол между точками из центра Земли (в радианах)
	c := 2 * math.Asin(math.Sqrt(h))

	// Умножаем на радиус Земли — получаем расстояние в км
	return earthRadiusKm * c
}
