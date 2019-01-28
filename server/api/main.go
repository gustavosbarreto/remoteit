package main

import (
	"crypto/sha256"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/cnf/structhash"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

func main() {
	r := gin.Default()

	session, err := mgo.Dial("mongodb://mongo:27017")
	if err != nil {
		panic(err)
	}

	r.Use(func(c *gin.Context) {
		s := session.Clone()

		defer s.Close()

		c.Set("db", s.DB("main"))
		c.Next()
	})

	r.POST("/auth", func(c *gin.Context) {
		db := c.MustGet("db").(*mgo.Database)

		var req AuthRequest

		err := c.MustBindWith(&req, binding.JSON)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		switch req.GrantType {
		case UserGrantType:
			// TODO: not implemented yet
			c.AbortWithError(http.StatusInternalServerError, errors.New("not implemented yet"))
			return
		case DeviceGrantType:
		default:
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		uid := sha256.Sum256(structhash.Dump(req.DeviceGrantData, 1))

		d := &Device{
			UID:      string(uid[:]),
			Identity: req.DeviceGrantData.Identity,
		}

		if err := db.C("devices").Insert(&d); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"uid": uid,
		})

		secretKey, err := ioutil.ReadFile("private.key")
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		signature, err := token.SignedString(secretKey)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": signature,
		})
	})

	r.Run()
}
