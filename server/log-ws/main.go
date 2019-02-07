package main

import (
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo"
)

func createLogEntry(c echo.Context) error {
	var msg struct {
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		Level     string    `json:"level"`
	}

	err := c.Bind(&msg)
	if err != nil {
		return err
	}

	fmt.Println(msg)

	return nil
}

func main() {
	e := echo.New()

	e.POST("/log", createLogEntry)

	log.Fatal(e.Start(":8080"))
}
