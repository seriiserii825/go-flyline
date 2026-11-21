package logger

import (
	"log/slog"
	"os"
)

// New создаёт логгер в зависимости от режима:
//   - release: JSON, уровень Info (для сборщиков логов: Loki, ELK и т.п.)
//   - debug/test: читаемый текст, уровень Debug (для терминала)
func New(mode string) *slog.Logger {
	var handler slog.Handler

	if mode == "release" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return slog.New(handler)
}
