package api

import (
	"net/http"
	"time"

	"log/slog"

	"subscribe_aggregation-main/internal/storage"
)

type Handler struct {
	Storage storage.StorageInterface
}

// NewHandler создаёт новый экземпляр Handler с интерфейсом StorageInterface (обратите внимание — без указателя)
func NewHandler(store storage.StorageInterface) *Handler {
	return &Handler{Storage: store}
}

var globalLogger *slog.Logger = slog.Default()

func SetLogger(l *slog.Logger) {
	globalLogger = l
}

func GetLogger() *slog.Logger {
	return globalLogger
}

// LoggingMiddleware пример middleware для логирования запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := GetLogger()
		start := time.Now()
		lrw := &LoggingResponseWriter{ResponseWriter: w, StatusCode: 200}

		logger.Info("Request started",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
		)

		next.ServeHTTP(lrw, r)

		duration := time.Since(start).Milliseconds()

		logger.Info("Request completed",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Int("status", lrw.StatusCode),
			slog.Int64("duration_ms", duration),
		)
	})
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
