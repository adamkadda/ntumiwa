package logging

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adamkadda/ntumiwa/internal/config"
)

func Setup(cfg config.LogConfig) {
	var level slog.Level

	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		log.Fatalf("Invalid LogConfig.Style value: %s", cfg.Level)
	}

	options := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	switch cfg.Style {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, options)
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, options)
	default:
		log.Fatalf("Invalid LogConfig.Style value: %s", cfg.Style)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

type loggerContextKey struct{}

var loggerKey = loggerContextKey{}

func GetLogger(r *http.Request) *slog.Logger {
	logger, ok := r.Context().Value(loggerKey).(*slog.Logger)
	if !ok {
		panic("could not find logger in context")
	}

	return logger
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// requestID := uuid.NewString()
			logger := slog.Default().With(
				slog.Group("request",
					// slog.String("id", requestID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					// slog.String("ip", r.RemoteAddr),
				),
			)

			ctx := context.WithValue(r.Context(), loggerKey, logger)
			r = r.WithContext(ctx)

			logger.Info("Request started")

			lw := &loggingWriter{
				ResponseWriter: w,
				request:        r,
			}

			next.ServeHTTP(lw, r)

			logger.Info("Request completed",
				slog.Group("response",
					slog.Int("status", lw.statusCode),
					slog.Int("size", lw.size),
				),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
