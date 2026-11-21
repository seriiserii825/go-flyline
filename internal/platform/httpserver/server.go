package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seriiserii825/go-flyline/internal/platform/middleware"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// Server — gin-роутер + http.Server с корректной остановкой.
type Server struct {
	engine *gin.Engine
	api    *gin.RouterGroup
	http   *http.Server
	log    *slog.Logger
}

// New создаёт сервер. addr — например ":8080", mode — debug|release|test.
func New(addr, mode string, log *slog.Logger) *Server {
	gin.SetMode(mode)

	engine := gin.New()
	engine.Use(
		middleware.RequestLogger(log), // снаружи: видит итоговый статус, в т.ч. 500 после паники
		gin.Recovery(),                // внутри: ловит панику в хендлере → 500
	)

	return &Server{
		engine: engine,
		api:    engine.Group("/api/v1"),
		log:    log,
		http: &http.Server{
			Addr:              addr,
			Handler:           engine,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
}

// API — группа /api/v1, сюда фичи регистрируют свои роуты.
func (s *Server) API() *gin.RouterGroup {
	return s.api
}

// Engine — корневой роутер (для /swagger и того, что вне /api/v1).
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Run запускает сервер и блокируется, пока не отменят ctx (Ctrl+C / SIGTERM)
// или сервер не упадёт. При отмене ждёт завершения текущих запросов.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http server started", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	s.log.Info("http server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	s.log.Info("http server stopped")
	return nil
}
