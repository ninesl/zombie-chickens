package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Config struct for describe configuration of the app.
type Config struct {
	Server *ServerConfig
}

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

var (
	once     sync.Once
	instance *Config
)

// NewConfig prepares config variables from environment or defaults.
func NewConfig() *Config {
	once.Do(func() {
		// host := getEnv("SERVER_HOST", "localhost")
		port := getEnvInt("SERVER_PORT", 8080)
		// readTimeout := getEnvInt("SERVER_READ_TIMEOUT", 30)
		// WriteTimeout must be 0 for SSE connections to work properly
		// SSE connections are long-lived and should not timeout
		// writeTimeout := getEnvInt("SERVER_WRITE_TIMEOUT", 0)
		// idleTimeout := getEnvInt("SERVER_IDLE_TIMEOUT", 120)

		instance = &Config{
			Server: &ServerConfig{
				Addr: fmt.Sprintf(":%d", port),
				// ReadTimeout:  time.Duration(readTimeout) * time.Second,
				// WriteTimeout: time.Duration(writeTimeout) * time.Second,
				// IdleTimeout:  time.Duration(idleTimeout) * time.Second,
			},
		}
	})

	return instance
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
