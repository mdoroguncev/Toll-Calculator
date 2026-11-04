package main

import (
	"log"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mdoroguncev/toll-calculator/internal/config"
	"github.com/mdoroguncev/toll-calculator/types"
)

var sendInterval = time.Second * 5

func genLatLong() (float64, float64) {
	return genCord(), genCord()
}

func genCord() float64 {
	n := float64(rand.Intn(180) - 90)
	f := rand.Float64()

	return n + f
}

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "obu")
	slog.SetDefault(logger)
	wsEndpoint := config.Get("OBU_WS_URL", "ws://127.0.0.1:30000/ws")
	obuIDs := generateOBUIDs(20)
	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}
	for {
		for i := 0; i < len(obuIDs); i++ {
			lat, long := genLatLong()
			data := types.OBUData{
				OBUID: obuIDs[i],
				Lat:   lat,
				Long:  long,
			}
			if err := conn.WriteJSON(data); err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(sendInterval)
	}
}

func generateOBUIDs(n int) []int {
	ids := make([]int, n)
	for i := 0; i < n; i++ {
		ids[i] = rand.Intn(999999)
	}
	return ids
}
