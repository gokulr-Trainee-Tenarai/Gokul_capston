package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	Port              string
	ConsulURL         string
	SSLMode           string
	ProductServiceURL string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../../.env")
	}
	return Config{
		DBHost:            get("ORDER_DB_HOST", "localhost"),
		DBPort:            get("ORDER_DB_PORT", "5432"),
		DBUser:            get("ORDER_DB_USER", "postgres"),
		DBPassword:        get("ORDER_DB_PASSWORD", "postgres"),
		DBName:            get("ORDER_DB_NAME", "order_product"),
		Port:              get("ORDER_SERVICE_PORT", "8082"),
		ConsulURL:         get("CONSUL_URL", "http://localhost:8500"),
		SSLMode:           get("ORDER_DB_SSLMODE", "require"),
		ProductServiceURL: get("PRODUCT_SERVICE_URL", "http://localhost:8081"),
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
