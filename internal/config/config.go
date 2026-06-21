package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	WorkerCount        int
	WorkerBatchSize    int32
	WorkerPollInterval time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://jobqueue:jobqueue@localhost:5433/jobqueue?sslmode=disable"
	}

	workerCount := getEnvInt("WORKER_COUNT", 3)
	workerBatchSize := getEnvInt("WORKER_BATCH_SIZE", 10)
	workerPollInterval := getEnvDuration("WORKER_POLL_INTERVAL", 2*time.Second)

	return Config{
		Port:               port,
		DatabaseURL:        databaseURL,
		WorkerCount:        workerCount,
		WorkerBatchSize:    int32(workerBatchSize),
		WorkerPollInterval: workerPollInterval,
	}
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
