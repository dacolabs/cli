package app

import (
	"fmt"
	"os"
	"time"

	"github.com/dacolabs/daco/internal/env"
)

// Config holds all configuration for the API service
type Config struct {
	Server    ServerConfig
	Log       LogConfig
}

// LogFormat is the output format for application logs.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatOTel LogFormat = "otel"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Format LogFormat
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

// LoadConfig reads configuration from environment variables using os.Getenv
func LoadConfig() (*Config, error) {
	return LoadConfigWithEnv(os.Getenv)
}

// LoadConfigWithEnv reads configuration using a custom getenv function (for testability)
func LoadConfigWithEnv(getenv func(string) string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            env.ParseEnvInt(getenv, "SERVER_PORT", 8080),
			ShutdownTimeout: env.ParseEnvDuration(getenv, "SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			ReadTimeout:     env.ParseEnvDuration(getenv, "SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    env.ParseEnvDuration(getenv, "SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     env.ParseEnvDuration(getenv, "SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Log: LogConfig{
			Format: LogFormat(env.ParseEnv(getenv, "LOG_FORMAT", string(LogFormatText))),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that required configuration is present and valid
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid SERVER_PORT: %d (must be 1-65535)", c.Server.Port)
	}

	switch c.Log.Format {
	case LogFormatText, LogFormatOTel:
	default:
		return fmt.Errorf("invalid LOG_FORMAT: %q (must be %q or %q)", c.Log.Format, LogFormatText, LogFormatOTel)
	}

	return nil
}
