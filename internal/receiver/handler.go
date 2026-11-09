package receiver

import (
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mdoroguncev/toll-calculator/internal/types"
)

type Receiver struct {
	prod DataProducer
}

func NewDataReceiver(brokers, topic string) (*Receiver, error) {
	var (
		p   DataProducer
		err error
	)
	p, err = NewKafkaProducer(brokers, topic)
	if err != nil {
		return nil, err
	}

	p = NewLogMiddleware(p)
	return &Receiver{
		prod: p,
	}, nil
}

func (dr *Receiver) produceData(data types.OBUData) error {
	return dr.prod.ProduceData(data)
}

func (dr *Receiver) Close() {
	dr.prod.Close()
}

func (dr *Receiver) HandleWS(w http.ResponseWriter, r *http.Request) {
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

func (dr *Receiver) wsReceiverLoop(conn *websocket.Conn) {
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
