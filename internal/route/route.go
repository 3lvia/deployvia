package route

import (
	"context"

	"github.com/3lvia/deployvia/internal/config"
	"github.com/3lvia/deployvia/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(conf *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(config.ConfigureMetrics(conf))

	return router
}

func RegisterRoutes(ctx context.Context, router *gin.Engine, conf *config.Config) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OK",
		})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.POST("/deployment", func(c *gin.Context) {
		handler.PostDeployment(ctx, c, conf)
	})
}
