package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func relayHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer conn.Close()

	for {
		var msg struct {
			Message   string    `json:"message"`
			Timestamp time.Time `json:"timestamp"`
			Level     string    `json:"level"`
		}

		err = conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		fmt.Println(msg)
	}
}

func main() {
	http.HandleFunc("/ws", relayHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
