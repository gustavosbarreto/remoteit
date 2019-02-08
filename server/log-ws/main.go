package main

import (
	"fmt"
	"log"
	"time"

	pubsub "github.com/alash3al/go-pubsub"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

var (
	upgrader = websocket.Upgrader{}
)

const pongWait = time.Minute

type Message struct {
	Device    string    `json:"device"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
}

var broker *pubsub.Broker

func createLogEntry(c echo.Context) error {
	msg := Message{}

	err := c.Bind(&msg)
	if err != nil {
		return err
	}

	if broker.Subscribers(msg.Device) > 0 {
		broker.Broadcast(msg, msg.Device)
	}

	return nil
}

func streamLog(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	defer ws.Close()

	s, err := broker.Attach()
	if err != nil {
		panic(err)
	}

	broker.Subscribe(s, c.Param("device"))

	for {
		msg := <-s.GetMessages()

		err := ws.WriteJSON(msg.GetPayload().(Message))
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	return nil
}

func main() {
	e := echo.New()

	broker = pubsub.NewBroker()

	e.POST("/log", createLogEntry)
	e.GET("/ws/:device", streamLog)

	log.Fatal(e.Start(":8080"))
}
