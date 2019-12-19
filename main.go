package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bjorand/nombda/engine"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	log         = logrus.New()
	listenAddr  string
	token       = os.Getenv("NOMBDA_TOKEN")
	configDir   = os.Getenv("CONFIG_DIR")
	version     string
	showVersion bool
	hookEngine  *engine.HookEngine
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

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isTokenAuthenticated(c) {
			c.Next()
			return
		}
		// c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		// c.

	}
}

func isTokenAuthenticated(c *gin.Context) bool {
	t := tokenHeader{}
	if err := c.ShouldBindHeader(&t); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return false
	}
	if t.AuthToken != token {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	}
	return false
}

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "server listen address")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()

	if showVersion {
		fmt.Printf("nopm-sh version %s", version)
		os.Exit(0)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		log.Fatal("Empty NOMBDA_TOKEN environment variable. Failing to start.")
	}

	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		log.Fatal("Empty CONFIG_DIR environment variable. Failing to start.")
	}

	hookEngine := engine.NewHookEngine(configDir)
	hookEngine.Secrets = engine.ReadSecretFromEnv()

	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(Base())

	authorized := router.Group("/")

	authorized.Use(AuthRequired())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": version})
	})

	authorized.GET("/hooks", func(c *gin.Context) {
		hooks, err := hookEngine.Hooks()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"hooks": hooks})
	})

	authorized.POST("/hooks/:id/:action", func(c *gin.Context) {
		hook, err := hookEngine.ReadHook(configDir, c.Param("id"), c.Param("action"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		run, err := hook.Run()
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": run.ID})
	})

	authorized.GET("/hooks/:id/:action/:run_id", func(c *gin.Context) {
		hook, err := hookEngine.ReadHook(configDir, c.Param("id"), c.Param("action"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		run, err := hook.GetRun(c.Param("run_id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"run": gin.H{
			"completed": run.Completed,
			"id":        run.ID,
			"exit_code": run.ExitCode,
		}})
	})

	authorized.GET("/hooks/:id/:action/:run_id/log", func(c *gin.Context) {
		hook, err := hookEngine.ReadHook(configDir, c.Param("id"), c.Param("action"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		run, err := hook.GetRun(c.Param("run_id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": err})
			return
		}
		c.String(http.StatusOK, run.Log())
	})

	router.Run(listenAddr)
}
