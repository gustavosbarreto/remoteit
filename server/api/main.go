package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
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
	Identity map[string]string `json:"identity"`
}

const (
	UserGrantType   = "user"
	DeviceGrantType = "device"
)

type Device struct {
	ID       bson.ObjectId     `json:"-" bson:"_id,omitempty"`
	UID      string            `json:"uid"`
	Identity map[string]string `json:"identity"`
	LastSeen time.Time         `json:"last_seen"`
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

	signBytes, err := ioutil.ReadFile("key.pem")
	if err != nil {
		panic(err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}

	verifyBytes, err := ioutil.ReadFile("key.pub")
	if err != nil {
		panic(err)
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
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
			UID:      hex.EncodeToString(uid[:]),
			Identity: req.DeviceGrantData.Identity,
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

	e.GET("/mqtt/auth", func(c echo.Context) error {
		q := AuthQuery{}

		if err := c.Bind(&q); err != nil {
			return err
		}

		ipaddr, err := net.LookupIP("server")
		if err != nil {
			return err
		}

		// Authorize connection from internal ssh-server client
		if q.IPAddr == ipaddr[0].String() {
			return nil
		}

		token, err := jwt.ParseWithClaims(q.Password, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return verifyKey, nil
		})
		if err != nil {
			return err
		}

		if claims, ok := token.Claims.(AuthClaims); ok && token.Valid {
			e.Logger.Info(claims)
			return nil
		}

		return nil
	})

	e.GET("/mqtt/acl", func(c echo.Context) error {
		q := ACLQuery{}

		if err := c.Bind(&q); err != nil {
			return err
		}

		return nil
	})

	e.POST("/mqtt/webhook", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		evt := WebHookEvent{}

		if err := c.Bind(&evt); err != nil {
			return err
		}

		switch evt.Action {
		case WebHookClientConnectedEventType:
		default:
			return nil
		}

		d := Device{}
		err = db.C("devices").Find(bson.M{"uid": evt.WebHookClientEvent.Username}).One(&d)
		if err != nil {
			return err
		}

		d.LastSeen = time.Now()

		_, err = db.C("devices").Upsert(bson.M{"uid": d.UID}, d)
		if err != nil {
			return err
		}

		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))
}
