package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	// ConfigInstance - синглтон конфигурации
	ConfigInstance *Config
	once           sync.Once

	// DB - глобальное подключение к базе
	DB *sqlx.DB
)

// Config содержит настройки приложения
type Config struct {
	ServerHost  string
	ServerPort  string
	PostgresDSN string
}

// loadEnv загружает .env один раз
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, loading environment variables from system")
	}
}

// LoadConfig инициализирует Config с параметрами из env с дефолтами
func LoadConfig() *Config {
	once.Do(func() {
		loadEnv()

		sslMode := os.Getenv("SSL_MODE")
		if sslMode == "" {
			sslMode = "disable"
		}
		serverHost := os.Getenv("SERVER_HOST")
		if serverHost == "" {
			serverHost = "0.0.0.0"
		}
		serverPort := os.Getenv("SERVER_PORT")
		if serverPort == "" {
			serverPort = "8080"
		}

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"),
			os.Getenv("POSTGRES_PORT"),
			sslMode,
		)

		ConfigInstance = &Config{
			ServerHost:  serverHost,
			ServerPort:  serverPort,
			PostgresDSN: dsn,
		}
	})
	return ConfigInstance
}

// InitDB подключается к базе и устанавливает DB
func InitDB() {
	cfg := LoadConfig()
	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	// Проверим соединение
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	DB = db
}
