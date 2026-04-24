package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	API        APIConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Log        LogConfig
	CORS       CORSConfig
	Migrations MigrationConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type APIConfig struct {
	Prefix string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	TimeZone        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret    string
	Issuer    string
	ExpiresIn time.Duration
}

type LogConfig struct {
	Level string
}

type CORSConfig struct {
	AllowOrigins string
	AllowHeaders string
	AllowMethods string
}

type MigrationConfig struct {
	Path string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name:            getEnv("APP_NAME", "school-enrollment-be"),
			Env:             getEnv("APP_ENV", "local"),
			Port:            getEnv("APP_PORT", "8080"),
			ReadTimeout:     getDurationEnv("APP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("APP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getDurationEnv("APP_IDLE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationEnv("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		API: APIConfig{
			Prefix: getEnv("API_PREFIX", "/api/v1"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "school_enrollment"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			TimeZone:        getEnv("DB_TIMEZONE", "Asia/Bangkok"),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 50),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "change-me"),
			Issuer:    getEnv("JWT_ISSUER", "school-enrollment-be"),
			ExpiresIn: getDurationEnv("JWT_EXPIRES_IN", 24*time.Hour),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		CORS: CORSConfig{
			AllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "*"),
			AllowHeaders: getEnv("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization"),
			AllowMethods: getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
		},
		Migrations: MigrationConfig{
			Path: getEnv("MIGRATIONS_PATH", "file://migrations"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.App.Port) == "" {
		return fmt.Errorf("APP_PORT is required")
	}

	if strings.TrimSpace(c.Database.Host) == "" {
		return fmt.Errorf("DB_HOST is required")
	}

	if strings.TrimSpace(c.Database.Name) == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return fallback
	}

	return parsed
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
