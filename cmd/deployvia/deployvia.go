package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/3lvia/core/applications/deployvia/pkg/appconfig"
	"github.com/3lvia/core/applications/deployvia/pkg/routes"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	config, err := appconfig.New(ctx)
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(appconfig.Metrics(config))

	router.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.POST("/deployment", func(c *gin.Context) {
		routes.PostDeployment(ctx, c, config)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Info("receive interrupt signal")

		if err := server.Close(); err != nil {
			log.Fatal("Server Close:", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Info("Server closed under request")
		} else {
			log.Fatal("Server closed unexpect")
		}
	}

	log.Println("Server exiting")
}
