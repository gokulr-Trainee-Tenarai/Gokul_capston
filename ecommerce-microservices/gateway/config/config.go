package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	ConsulURL string
}

func Load() Config {
	_ = godotenv.Load()

	// In AWS Elastic Beanstalk, EB passes PORT=5000.
	// On your local machine, it falls back to GATEWAY_PORT (8080).
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("GATEWAY_PORT")
	}
	if port == "" {
		port = "8080"
	}

	consulURL := os.Getenv("CONSUL_URL")
	if consulURL == "" {
		consulURL = "http://localhost:8500"
	}

	return Config{Port: port, ConsulURL: consulURL}
}