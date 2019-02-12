package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

type ConfigOptions struct {
	DeviceID      string `envconfig:"device_id"`
	ServerAddress string `envconfig:"server_address"`
	AuthToken     string `envconfig:"auth_token"`
	PrivateKey    string `envconfig:"private_key"`
}

type Endpoints struct {
	API  string `json:"api"`
	SSH  string `json:"ssh"`
	MQTT string `json:"mqtt"`
}

func main() {
	opts := ConfigOptions{}

	err := envconfig.Process("", &opts)
	if err != nil {
		logrus.Panic(err)
	}

	endpoints := Endpoints{}

	_, _, errs := gorequest.New().Get(fmt.Sprintf("%s/endpoints", opts.ServerAddress)).EndStruct(&endpoints)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{"err": errs[0]}).Panic("Failed to get endpoints")
	}

	fmt.Println(endpoints)

	b := NewBroker(endpoints.MQTT, opts.DeviceID, opts.AuthToken)

	b.Connect()

	l, err := NewLogWatcher()
	if err != nil {
		panic(err)
	}

	go l.Watch()

	for {
		_ = <-l.Channel()
		//fmt.Println(e)
	}
}
