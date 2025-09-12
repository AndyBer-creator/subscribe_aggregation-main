package logging

import (
	"log/slog"
	"os"
	"sync"
)

var (
	//глобальный синглтон логгера
	logger *slog.Logger
	level  = new(slog.LevelVar)
	//синхронизация инициализации логгера
	once sync.Once
)

func parseLevel(levelStr string) slog.Level {
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo // По умолчанию
	}
}

// Логгер создается один раз с JSON обработчиком,
// записывающим логи в файл app.log и добавляющим исходник вызова (source).
func GetLogger() *slog.Logger {
	once.Do(func() {
		f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("cannot open log file: " + err.Error())
		}
		// Инициализация JSON хендлера с добавлением информации об источнике вызова
		handler := slog.NewJSONHandler(f, &slog.HandlerOptions{AddSource: true})

		logger = slog.New(handler)
	})
	return logger
}
func SetLevel(levelStr string) {
	lvl := parseLevel(levelStr)
	level.Set(lvl)
}
