package main

import (
	"log"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/mdoroguncev/toll-calculator/types"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "data_receiver")
	slog.SetDefault(logger)
	recv, err := NewDataReceiver()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ws", recv.HandleWS)
	log.Fatal(http.ListenAndServe(":30000", nil))
}

type DataReceiver struct {
	prod DataProducer
}

func NewDataReceiver() (*DataReceiver, error) {
	var (
		p          DataProducer
		err        error
		kafkaTopic = "obuData"
	)
	p, err = NewKafkaProducer(kafkaTopic)
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
