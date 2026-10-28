package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	Port            string
	ConsulURL       string
	SSLMode         string
	OrderServiceURL string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../../.env")
	}
	return Config{
		DBHost:          get("PRODUCT_DB_HOST", "localhost"),
		DBPort:          get("PRODUCT_DB_PORT", "5432"),
		DBUser:          get("PRODUCT_DB_USER", "postgres"),
		DBPassword:      get("PRODUCT_DB_PASSWORD", "postgres"),
		DBName:          get("PRODUCT_DB_NAME", "order_product"),
		Port:            get("PRODUCT_SERVICE_PORT", "8081"),
		ConsulURL:       get("CONSUL_URL", "http://localhost:8500"),
		SSLMode:         get("PRODUCT_DB_SSLMODE", "require"),
		OrderServiceURL: get("ORDER_SERVICE_URL", "http://localhost:8082"),
	}
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (c Config) DSN() string {
	escapedPassword := url.QueryEscape(c.DBPassword)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, escapedPassword, c.DBHost, c.DBPort, c.DBName, c.SSLMode)
}
