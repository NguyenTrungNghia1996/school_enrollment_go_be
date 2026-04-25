package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
	CORS     CORSConfig
}

type AppConfig struct {
	Name     string
	Env      string
	Port     string
	Timezone string
	Location *time.Location
}

type DatabaseConfig struct {
	Host              string
	Port              string
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxIdleConns      int
	MaxOpenConns      int
	ConnMaxLifetime   time.Duration
	EnableAutoMigrate bool
}

type AuthConfig struct {
	AdminJWT JWTConfig
	UserJWT  JWTConfig
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type CORSConfig struct {
	AllowOrigins string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	var validationErrors []string

	timezone := requiredEnv("APP_TIMEZONE", &validationErrors)
	location := parseLocation("APP_TIMEZONE", timezone, &validationErrors)

	cfg := &Config{
		App: AppConfig{
			Name:     requiredEnv("APP_NAME", &validationErrors),
			Env:      requiredEnv("APP_ENV", &validationErrors),
			Port:     requiredEnv("APP_PORT", &validationErrors),
			Timezone: timezone,
			Location: location,
		},
		Database: DatabaseConfig{
			Host:              requiredEnv("DB_HOST", &validationErrors),
			Port:              requiredEnv("DB_PORT", &validationErrors),
			User:              requiredEnv("DB_USER", &validationErrors),
			Password:          requiredEnv("DB_PASSWORD", &validationErrors),
			Name:              requiredEnv("DB_NAME", &validationErrors),
			SSLMode:           requiredEnv("DB_SSLMODE", &validationErrors),
			MaxIdleConns:      getIntEnv("DB_MAX_IDLE_CONNS", 10, &validationErrors),
			MaxOpenConns:      getIntEnv("DB_MAX_OPEN_CONNS", 50, &validationErrors),
			ConnMaxLifetime:   getDurationEnv("DB_CONN_MAX_LIFETIME", 30*time.Minute, &validationErrors),
			EnableAutoMigrate: getBoolEnv("DB_ENABLE_AUTO_MIGRATE", false, &validationErrors),
		},
		Auth: AuthConfig{
			AdminJWT: JWTConfig{
				Secret:    requiredEnv("ADMIN_JWT_SECRET", &validationErrors),
				ExpiresIn: requiredDuration("ADMIN_JWT_EXPIRES_IN", &validationErrors),
			},
			UserJWT: JWTConfig{
				Secret:    requiredEnv("USER_JWT_SECRET", &validationErrors),
				ExpiresIn: requiredDuration("USER_JWT_EXPIRES_IN", &validationErrors),
			},
		},
		CORS: CORSConfig{
			AllowOrigins: requiredEnv("CORS_ALLOW_ORIGINS", &validationErrors),
		},
	}

	if err := cfg.Validate(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("invalid config: %s", strings.Join(validationErrors, "; "))
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var validationErrors []string

	if _, err := fmt.Sscanf(c.App.Port, "%d", new(int)); err != nil {
		validationErrors = append(validationErrors, "APP_PORT must be a valid integer")
	}

	if _, err := fmt.Sscanf(c.Database.Port, "%d", new(int)); err != nil {
		validationErrors = append(validationErrors, "DB_PORT must be a valid integer")
	}

	if strings.TrimSpace(c.CORS.AllowOrigins) == "" {
		validationErrors = append(validationErrors, "CORS_ALLOW_ORIGINS is required")
	}

	if c.Database.MaxIdleConns < 0 {
		validationErrors = append(validationErrors, "DB_MAX_IDLE_CONNS must be greater than or equal to 0")
	}

	if c.Database.MaxOpenConns <= 0 {
		validationErrors = append(validationErrors, "DB_MAX_OPEN_CONNS must be greater than 0")
	}

	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		validationErrors = append(validationErrors, "DB_MAX_IDLE_CONNS must be less than or equal to DB_MAX_OPEN_CONNS")
	}

	if c.Database.ConnMaxLifetime < 0 {
		validationErrors = append(validationErrors, "DB_CONN_MAX_LIFETIME must be greater than or equal to 0")
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

func requiredEnv(key string, validationErrors *[]string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s is required", key))
		return ""
	}

	return value
}

func requiredDuration(key string, validationErrors *[]string) time.Duration {
	value := requiredEnv(key, validationErrors)
	if value == "" {
		return 0
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s must be a valid duration", key))
		return 0
	}

	return duration
}

func getIntEnv(key string, fallback int, validationErrors *[]string) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s must be a valid integer", key))
		return fallback
	}

	return parsed
}

func getDurationEnv(key string, fallback time.Duration, validationErrors *[]string) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s must be a valid duration", key))
		return fallback
	}

	return duration
}

func getBoolEnv(key string, fallback bool, validationErrors *[]string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s must be a valid boolean", key))
		return fallback
	}
}

func parseLocation(key, value string, validationErrors *[]string) *time.Location {
	if strings.TrimSpace(value) == "" {
		return time.UTC
	}

	location, err := time.LoadLocation(value)
	if err != nil {
		*validationErrors = append(*validationErrors, fmt.Sprintf("%s must be a valid IANA timezone", key))
		return time.UTC
	}

	return location
}
