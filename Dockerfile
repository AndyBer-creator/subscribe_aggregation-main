# Используем официальный Go 1.25 образ в alpine для сборки
FROM golang:1.25-alpine AS builder

# Установка git и ca-certificates для go mod
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем модули для кеширования
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка бинарника с отключением CGO для кросс-платформенности
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o subscribe_aggregation ./cmd/

# Финальный минимальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/subscribe_aggregation /app/subscribe_aggregation

# Указываем порт, который слушает приложение
EXPOSE 8080

# Запуск приложения
CMD ["/app/subscribe_aggregation"]
