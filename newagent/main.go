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

func (e *Endpoints) buildAPIUrl(uri string) string {
	return fmt.Sprintf("http://%s/api/%s", e.API, uri)
}

type AuthRequest struct {
	Identity  *DeviceIdentity `json:"identity"`
	PublicKey string          `json:"public_key"`
}

type AuthResponse struct {
	UID   string `json:"uid"`
	Token string `json:"token"`
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

	identity, err := NewDeviceIdentity()
	if err != nil {
		logrus.Fatal(err)
	}

	var auth AuthResponse
	_, _, errs = gorequest.New().Post(endpoints.buildAPIUrl("/devices/auth")).Send(&AuthRequest{
		Identity:  identity,
		PublicKey: "testing",
	}).EndStruct(&auth)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{"errs": errs}).Panic("Failed authenticate device")
	}

	fmt.Println(auth)

	b := NewBroker(endpoints.MQTT, opts.DeviceID, opts.AuthToken)

	b.Connect()

	l, err := NewLogWatcher()
	if err != nil {
		panic(err)
	}

	logWatcher := l.Watch()

	for {
		<-logWatcher
		//		e := <-logWatcher
		//		fmt.Println(e)
	}
}
