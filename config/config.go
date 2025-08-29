package config

import (
	"os"
	"strings"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	KafkaBrokers []string
	KafkaTopic   string
	HTTPPort     string
}

func Load() *Config {
	return &Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "user"),
		DBPassword:   getEnv("DB_PASSWORD", "password"),
		DBName:       getEnv("DB_NAME", "orders_db"),
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "redpanda:9092"), ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "orders"),
		HTTPPort:     getEnv("HTTP_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
