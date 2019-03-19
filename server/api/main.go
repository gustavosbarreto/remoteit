package main

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/cnf/structhash"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type AuthRequest struct {
	GrantType string `json:"grant_type" `
	UserGrantData
	DeviceGrantData
}

type UserGrantData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DeviceGrantData struct {
	Identity  map[string]string `json:"identity"`
	PublicKey string            `json:"public_key"`
}

const (
	UserGrantType   = "user"
	DeviceGrantType = "device"
)

type Device struct {
	ID        bson.ObjectId     `json:"-" bson:"_id,omitempty"`
	UID       string            `json:"uid"`
	Identity  map[string]string `json:"identity"`
	PublicKey string            `json:"public_key" bson:"public_key"`
	LastSeen  time.Time         `json:"last_seen"`
}

type AuthQuery struct {
	Username string `query:"username"`
	Password string `query:"password"`
	IPAddr   string `query:"ipaddr"`
}

type ACLQuery struct {
	Access   string `query:"access"`
	Username string `query:"username"`
	Topic    string `query:"topic"`
	IPAddr   string `query:"ipaddr"`
}

type AuthClaims struct {
	UID string `json:"uid"`

	jwt.StandardClaims
}

type WebHookEvent struct {
	Action string `json:"action"`

	WebHookClientEvent
}

type WebHookClientEvent struct {
	ClientID string `json:"client_id"`
	Username string `json:"username"`
}

const (
	WebHookClientConnectedEventType    = "client_connected"
	WebHookClientDisconnectedEventType = "client_disconnected"
)

var verifyKey *rsa.PublicKey

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	session, err := mgo.Dial("mongodb://mongo:27017")
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("devices").EnsureIndex(mgo.Index{
		Key:        []string{"uid"},
		Unique:     true,
		Name:       "uid",
		Background: false,
	})
	if err != nil {
		panic(err)
	}

	signBytes, err := ioutil.ReadFile(os.Getenv("API_PRIV_KEY_PATH"))
	if err != nil {
		panic(err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}

	verifyBytes, err := ioutil.ReadFile(os.Getenv("API_PUB_KEY_PATH"))
	if err != nil {
		panic(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			s := session.Clone()

			defer s.Close()

			c.Set("db", s.DB("main"))

			return next(c)
		}
	})

	e.POST("/auth", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		var req AuthRequest

		err := c.Bind(&req)
		if err != nil {
			return err
		}

		switch req.GrantType {
		case UserGrantType:
			// TODO: not implemented yet
			return errors.New("not implemented yet")
		case DeviceGrantType:
		default:
			return err
		}

		uid := sha256.Sum256(structhash.Dump(req.DeviceGrantData, 1))

		d := &Device{
			UID:       hex.EncodeToString(uid[:]),
			Identity:  req.DeviceGrantData.Identity,
			PublicKey: req.DeviceGrantData.PublicKey,
		}

		if err := db.C("devices").Insert(&d); err != nil && !mgo.IsDup(err) {
			return err
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, AuthClaims{
			UID: string(uid[:]),
		})

		signature, err := token.SignedString(signKey)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, echo.Map{
			"uid":   d.UID,
			"token": signature,
		})
	})

	e.GET("/devices", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		var devices []Device
		if err := db.C("devices").Find(bson.M{}).All(&devices); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, devices)
	})

	e.GET("/mqtt/auth", AuthenticateMqttClient)
	e.GET("/mqtt/acl", AuthorizeMqttClient)
	e.POST("/mqtt/webhook", ProcessMqttEvent)

	e.Logger.Fatal(e.Start(":8080"))
}
