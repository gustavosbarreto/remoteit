package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/kelseyhightower/envconfig"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ConfigOptions struct {
	ServerAddress string `envconfig:"server_address"`
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

	pubKey, err := readPublicKey(opts.PrivateKey)
	if err != nil {
		logrus.Fatal(err)
	}

	var auth AuthResponse

	_, _, errs = gorequest.New().Post(endpoints.buildAPIUrl("/devices/auth")).Send(&AuthRequest{
		Identity: identity,
		PublicKey: string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pubKey),
		})),
	}).EndStruct(&auth)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{"errs": errs}).Panic("Failed authenticate device")
	}

	freePort, err := getFreePort()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "failed to get free port"))
	}

	server := NewSSHServer(freePort)
	client := NewSSHClient(opts.PrivateKey, endpoints.SSH, freePort)

	go func() {
		logrus.Fatal(server.ListenAndServe())
	}()

	b := NewBroker(endpoints.MQTT, auth.UID, auth.Token)

	b.Subscribe(fmt.Sprintf("connect/%s", auth.UID), client.connect)
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

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func readPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Failed to decode PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &key.PublicKey, nil
}
