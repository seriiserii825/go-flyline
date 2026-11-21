package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger пишет одну строку лога на каждый HTTP-запрос.
// Уровень зависит от статуса: 5xx → Error, 4xx → Warn, остальное → Info.
func RequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // выполняем остальные middleware и хендлер

		status := c.Writer.Status()

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
			slog.Int("size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		log.LogAttrs(c.Request.Context(), level, "http request", attrs...)
	}
}
