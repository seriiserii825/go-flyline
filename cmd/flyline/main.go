package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/seriiserii825/go-flyline/internal/platform/config"
	"github.com/seriiserii825/go-flyline/internal/platform/httpserver"
	"github.com/seriiserii825/go-flyline/internal/platform/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("app stopped with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.GinMode)
	slog.SetDefault(log) // чтобы slog.Error(...) в main тоже шёл через этот логгер

	srv := httpserver.New(cfg.HTTPAddr(), cfg.GinMode, log)

	srv.API().GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return srv.Run(ctx)
}
