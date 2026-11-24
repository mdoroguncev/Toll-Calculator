package main

import (
	"context"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mdoroguncev/toll-calculator/internal/calculator"
	"github.com/mdoroguncev/toll-calculator/internal/config"
	"github.com/mdoroguncev/toll-calculator/internal/types"
)

var route = [2]calculator.Point{
	{Lat: 55.7500, Long: 37.6100},
	{Lat: 55.7590, Long: 37.6100},
}

var routeMeters = calculator.DistanceKm(route[0], route[1]) * 1000

type vehicle struct {
	obuID        int
	offsetMeters float64
	speedKmh     float64
	direction    int
}

func (v *vehicle) position() calculator.Point {
	a, b := route[0], route[1]
	fraction := v.offsetMeters / routeMeters

	return calculator.Point{
		Lat:  a.Lat + (b.Lat-a.Lat)*fraction,
		Long: a.Long + (b.Long-a.Long)*fraction,
	}
}

func (v *vehicle) advance(elapsed time.Duration) []calculator.Point {
	var points []calculator.Point
	meters := v.speedKmh / 3.6 * elapsed.Seconds()
	v.offsetMeters += float64(v.direction) * meters

	for v.offsetMeters > routeMeters || v.offsetMeters < 0 {
		if v.offsetMeters > routeMeters {
			points = append(points, route[1])
			v.offsetMeters = 2*routeMeters - v.offsetMeters
			v.direction = -1
		} else {
			points = append(points, route[0])
			v.offsetMeters = -v.offsetMeters
			v.direction = 1
		}
	}
	return append(points, v.position())
}

var sendInterval = time.Second * 5

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "obu")
	slog.SetDefault(logger)

	wsEndpoint := config.Get("OBU_WS_URL", "ws://127.0.0.1:30000/ws")
	obuIDs := generateOBUIDs(20)
	vehicles := make([]vehicle, len(obuIDs))
	for i, id := range obuIDs {
		vehicles[i] = vehicle{
			obuID:        id,
			offsetMeters: float64(i) * routeMeters / float64(len(obuIDs)),
			speedKmh:     60,
			direction:    1,
		}
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(sendInterval)
	defer ticker.Stop()
	var elapsed time.Duration // Первый пакет содержит стартовые координаты.

	for {
		for i := range vehicles {
			v := &vehicles[i]
			for _, p := range v.advance(elapsed) {

				data := types.OBUData{
					OBUID: v.obuID,
					Lat:   p.Lat,
					Long:  p.Long,
				}
				if err := conn.WriteJSON(data); err != nil {
					log.Fatal(err)
				}

			}
		}
		select {
		case <-ctx.Done():
			if err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				slog.Warn("failed to send close message", "err", err)
			}
			conn.Close()
			slog.Info("obu stopped")
			return
		case <-ticker.C:
			elapsed = sendInterval
		}
	}
}

func generateOBUIDs(n int) []int {
	ids := make([]int, n)
	for i := 0; i < n; i++ {
		ids[i] = rand.Intn(999999)
	}
	return ids
}
