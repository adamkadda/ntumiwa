package middleware

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/adamkadda/ntumiwa-site/shared/config"
	"github.com/google/uuid"
)

func loggingSetup(cfg config.LogConfig) {
	var handler slog.Handler

	switch cfg.Style {
	case "TEXT":
		handler = slog.NewTextHandler(os.Stdout, nil)
	case "JSON":
		handler = slog.NewJSONHandler(os.Stdout, nil)
	default:
		log.Fatalf("Invalid LogConfig.Style value: %s", cfg.Style)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.Info("Logger setup successful")
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

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.NewString()

		logger := slog.Default().With(
			slog.Group("request",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			),
		)

		ctx := context.WithValue(r.Context(), loggerKey, logger)
		r = r.WithContext(ctx)

		logger.Info("request started")

		ww := NewWrappedWriter(w)

		next.ServeHTTP(ww, r)

		logger.Info("request completed",
			slog.Group("response",
				slog.Int("status", ww.statusCode),
				slog.Int("size", ww.size),
			),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
