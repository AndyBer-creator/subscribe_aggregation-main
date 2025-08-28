package logging

import (
	"log/slog"
	"os"
	"sync"
)

var (
	logger *slog.Logger
	once   sync.Once
)

func GetLogger() *slog.Logger {
	once.Do(func() {
		f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("cannot open log file: " + err.Error())
		}

		handler := slog.NewJSONHandler(f, &slog.HandlerOptions{AddSource: true})

		logger = slog.New(handler)
	})
	return logger
}
