package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/bjorand/nombda/engine"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	log        = logrus.New()
	listenAddr string
	token      = os.Getenv("NOMBDA_TOKEN")
	configDir  = os.Getenv("CONFIG_DIR")
)

type tokenHeader struct {
	AuthToken string `header:"Auth-Token"`
}

func Base() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Do something
		c.Next()
	}
}

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "server listen address")
	flag.Parse()

	token = strings.TrimSpace(token)
	if token == "" {
		log.Fatal("Empty NOMBDA_TOKEN environment variable. Failing to start.")
	}

	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		log.Fatal("Empty CONFIG_DIR environment variable. Failing to start.")
	}

	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(Base())

	router.POST("/hooks/:id/:action", func(c *gin.Context) {
		t := tokenHeader{}
		if err := c.ShouldBindHeader(&t); err != nil {
			c.JSON(500, err)
			return
		}
		if t.AuthToken != token {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}
		hook, err := engine.ReadHook(configDir, c.Param("id"), c.Param("action"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		if err := hook.Run(); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	})

	router.Run(listenAddr)
}
