package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cnf/structhash"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	mgo "gopkg.in/mgo.v2"
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
	UID      string            `json:"device"`
	Identity map[string]string `json:"identity"`
}

type AuthQuery struct {
	Username string `query:"username"`
	Password string `query:"password"`
}

type AuthClaims struct {
	UID string `json:"uid"`

	jwt.StandardClaims
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	session, err := mgo.Dial("mongodb://mongo:27017")
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
			UID:      string(uid[:]),
			Identity: req.DeviceGrantData.Identity,
		}

		if err := db.C("devices").Insert(&d); err != nil {
			return err
		}

		now := jwt.TimeFunc()

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, AuthClaims{
			UID: string(uid[:]),

			StandardClaims: jwt.StandardClaims{
				NotBefore: now.Unix(),
				IssuedAt:  now.Unix(),
				Issuer:    "LetAuth",
				Subject:   "All",
			},
		})

		signature, err := token.SignedString(signKey)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, echo.Map{
			"token": signature,
		})
	})

	e.GET("/mqtt/auth", func(c echo.Context) error {
		q := AuthQuery{}

		if err := c.Bind(&q); err != nil {
			return err
		}

		if q.Username != "use-token-auth" {
			return errors.New("Invalid username")
		}

		fmt.Println(q.Password)

		token, err := jwt.ParseWithClaims(q.Password, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return verifyKey, nil
		})
		if err != nil {
			fmt.Println("Deu erro")
			return err
		}

		if claims, ok := token.Claims.(AuthClaims); ok && token.Valid {
			e.Logger.Info(claims)
			return nil
		}

		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))
}
