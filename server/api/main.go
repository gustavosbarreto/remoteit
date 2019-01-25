package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	mgo "gopkg.in/mgo.v2"
)

type AuthRequest struct {
	GrantType string `json:"grant_type" `
	UserGrantData
	AppGrantData
}

var merda string

type UserGrantData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AppGrantData struct {
	Identity string `json:"identity"`
}

const (
	UserGrantType = "user"
	AppGrantType  = "app"
)

func main() {
	r := gin.Default()

	session, err := mgo.Dial("mongodb://localhost:27017")
	if err != nil {
		panic(err)
	}

	r.Use(func(c *gin.Context) {
		s := session.Clone()

		defer s.Close()

		c.Set("db", s.DB("auth"))
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
		case AppGrantType:
			fmt.Println("app")
		default:
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		fmt.Println(req.UserGrantData)
		fmt.Println(req.AppGrantData)

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()
}
