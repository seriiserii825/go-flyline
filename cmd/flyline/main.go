package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default() // = gin.New() + Logger + Recovery
	r.GET("/api/v1/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.Run(":8080")
}
