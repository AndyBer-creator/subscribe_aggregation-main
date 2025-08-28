FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o subscribe_aggregation ./cmd/
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/subscribe_aggregation /subscribe_aggregation
# COPY .env /app/.env
EXPOSE 8080
CMD ["/subscribe_aggregation"]

