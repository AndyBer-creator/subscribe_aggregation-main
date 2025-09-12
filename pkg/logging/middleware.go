package logging

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ctxKeyRequestID — тип для ключа request ID в контексте запроса,
// чтобы избежать коллизий с ключами из других пакетов.
type ctxKeyRequestID struct{}

// RequestIDFromContext извлекает request ID из контекста запроса.
// Если в контексте нет request ID, возвращает пустую строку.
func RequestIDFromContext(ctx context.Context) string {
	id, ok := ctx.Value(ctxKeyRequestID{}).(string)
	if !ok {
		return ""
	}
	return id
}

// Middleware с логированием request ID, HTTP статуса и продолжительности
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()

		// Добавляем request ID в контекст
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, reqID)
		r = r.WithContext(ctx)

		logger := GetLogger()
		start := time.Now()

		logger.Info("Request started",
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
		)

		// Оборачиваем http.ResponseWriter, чтобы получить статус ответа
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}
		// Передаем управление следующему обработчику
		next.ServeHTTP(lrw, r)
		// Вычисляем длительность обработки запроса
		duration := time.Since(start).Milliseconds()

		logger.Info("Request completed",
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Int("status", lrw.statusCode),
			slog.Int64("duration_ms", duration),
		)
	})
}

// loggingResponseWriter — обертка над http.ResponseWriter,
// позволяющая захватить HTTP статус код ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader перехватывает установку HTTP статуса ответа,
// сохраняя его для последующего логирования.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
