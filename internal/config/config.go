package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://jobqueue:jobqueue@localhost:5433/jobqueue?sslmode=disable"
	}

	return Config{
		Port: port,
	}
}
