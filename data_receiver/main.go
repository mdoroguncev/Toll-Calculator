package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mdoroguncev/toll-calculator/types"
)

func main() {
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
		log.Printf("error: %v", err)
		return
	}

	go dr.wsReceiverLoop(conn)
}

func (dr *DataReceiver) wsReceiverLoop(conn *websocket.Conn) {
	defer conn.Close()
	fmt.Println("New OBU connected!")
	for {
		var data types.OBUData
		if err := conn.ReadJSON(&data); err != nil {
			log.Println("read error: ", err)
			return
		}
		data.RequestID = rand.IntN(math.MaxInt)
		if err := dr.produceData(data); err != nil {
			fmt.Println("kafka produce error: ", err)
		}
	}
}
