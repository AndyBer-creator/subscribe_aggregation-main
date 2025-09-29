package api

import (
	"net/http"
	"time"

	"log/slog"

	"subscribe_aggregation-main/internal/storage"
	"subscribe_aggregation-main/pkg/logging"
)

type Handler struct {
	storage storage.StorageInterface
}

// NewHandler создаёт новый экземпляр Handler с интерфейсом StorageInterface (обратите внимание — без указателя)
func NewHandler(store storage.StorageInterface) *Handler {
	return &Handler{storage: store}
}

// LoggingMiddleware пример middleware для логирования запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger()
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

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
			slog.Int("status", lrw.statusCode),
			slog.Int64("duration_ms", duration),
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
