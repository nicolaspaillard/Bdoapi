package config

import (
	"fmt"
	"os"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

var ConnectionString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
	GetEnv("DB_USER", "postgres"),
	GetEnv("DB_PASSWORD", "password"),
	GetEnv("DB_HOST", "localhost"),
	GetEnv("DB_PORT", "5432"),
	GetEnv("DB_NAME", "bdoapi"))
