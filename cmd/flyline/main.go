package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/seriiserii825/go-flyline/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	gin.SetMode(cfg.GinMode)
	r := gin.Default()
	r.GET("/api/v1/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	if err := r.Run(cfg.HTTPAddr()); err != nil {
		log.Fatal(err)
	}
}
