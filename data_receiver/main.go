package main

import (
	"log"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/mdoroguncev/toll-calculator/internal/config"
	"github.com/mdoroguncev/toll-calculator/types"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "data_receiver")
	slog.SetDefault(logger)

	var (
		listenAddr   = config.Get("RECEIVER_ADDR", ":30000")
		kafkaBrokers = config.Get("KAFKA_BROKERS", "localhost:9092")
		kafkaTopic   = config.Get("KAFKA_TOPIC", "obuData")
	)
	recv, err := NewDataReceiver(kafkaBrokers, kafkaTopic)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ws", recv.HandleWS)
	slog.Info("receiver started", "addr", listenAddr, "kafka", kafkaBrokers, "topic", kafkaTopic)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

type DataReceiver struct {
	prod DataProducer
}

func NewDataReceiver(brokers, topic string) (*DataReceiver, error) {
	var (
		p   DataProducer
		err error
	)
	p, err = NewKafkaProducer(brokers, topic)
	if err != nil {
		return nil, err
	}

	p = NewLogMiddleware(p)
	return &DataReceiver{
		prod: p,
	}, nil
}

func (dr *DataReceiver) produceData(data types.OBUData) error {
	return dr.prod.ProduceData(data)
}

func (dr *DataReceiver) HandleWS(w http.ResponseWriter, r *http.Request) {
	u := websocket.Upgrader{
		ReadBufferSize:  1028,
		WriteBufferSize: 1028,
	}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "err", err)
		return
	}

	go dr.wsReceiverLoop(conn)
}

func (dr *DataReceiver) wsReceiverLoop(conn *websocket.Conn) {
	defer conn.Close()
	slog.Info("New OBU connected", "remote", conn.RemoteAddr())
	for {
		var data types.OBUData
		if err := conn.ReadJSON(&data); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Error("unexpected close error", "err", err)
			} else {
				slog.Info("obu disconnected")
			}
			return
		}
		data.RequestID = rand.IntN(math.MaxInt)
		if err := dr.produceData(data); err != nil {
			slog.Error("kafka produce error", "err", err)
		}
	}
}
