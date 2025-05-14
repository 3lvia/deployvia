package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/3lvia/deployvia/internal/config"
	"github.com/3lvia/deployvia/internal/route"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	config, err := config.New(ctx)
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	router := route.SetupRouter(config)
	route.RegisterRoutes(ctx, router, config)

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Info("receive interrupt signal")

		if err := server.Close(); err != nil {
			log.Error("failed to close server:", err)
		} else {
			log.Info("server closed")
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Info("server closed under request")
		} else {
			log.Fatal("server closed unexpectedly:", err)
		}
	}

	log.Info("Server exiting")
}
