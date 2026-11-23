package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func healthHandler(db Pinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := db.Ping(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
				return
			}
		}
	}
}
