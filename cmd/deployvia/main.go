package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/3lvia/core/applications/deployvia/internal/config"
	"github.com/3lvia/core/applications/deployvia/internal/route"
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
