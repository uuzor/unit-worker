package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the auth service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port            int
	GRPCPort        int
	Environment     string // development, staging, production
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	AccessTokenSecret   string
	RefreshTokenSecret  string
	AccessTokenExpiry   time.Duration // 15 minutes
	RefreshTokenExpiry  time.Duration // 30 days
	Issuer              string
}

type EmailConfig struct {
	SMTPHost            string
	SMTPPort            int
	SMTPUser            string
	SMTPPassword        string
	FromEmail           string
	FromName            string
	VerificationExpiry  time.Duration // 24 hours
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			GRPCPort:        getEnvAsInt("GRPC_PORT", 9000),
			Environment:     getEnv("ENVIRONMENT", "development"),
			ShutdownTimeout: getEnvAsDuration("SHUTDOWN_TIMEOUT", "10s"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "ikampus_user"),
			Password:        getEnv("DB_PASSWORD", "changeme"),
			DBName:          getEnv("DB_NAME", "ikampus"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
		},
		JWT: JWTConfig{
			AccessTokenSecret:  getEnv("JWT_ACCESS_SECRET", "change-me-in-production-access"),
			RefreshTokenSecret: getEnv("JWT_REFRESH_SECRET", "change-me-in-production-refresh"),
			AccessTokenExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY", "15m"),
			RefreshTokenExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", "720h"), // 30 days
			Issuer:             getEnv("JWT_ISSUER", "ikampus-auth"),
		},
		Email: EmailConfig{
			SMTPHost:           getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:           getEnvAsInt("SMTP_PORT", 587),
			SMTPUser:           getEnv("SMTP_USER", ""),
			SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
			FromEmail:          getEnv("FROM_EMAIL", "noreply@ikampus.com"),
			FromName:           getEnv("FROM_NAME", "iKampus"),
			VerificationExpiry: getEnvAsDuration("VERIFICATION_EXPIRY", "24h"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.JWT.AccessTokenSecret == "" || c.JWT.AccessTokenSecret == "change-me-in-production-access" {
		if c.Server.Environment == "production" {
			return fmt.Errorf("JWT_ACCESS_SECRET must be set in production")
		}
	}
	if c.JWT.RefreshTokenSecret == "" || c.JWT.RefreshTokenSecret == "change-me-in-production-refresh" {
		if c.Server.Environment == "production" {
			return fmt.Errorf("JWT_REFRESH_SECRET must be set in production")
		}
	}
	return nil
}

// DatabaseDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		valueStr = defaultValue
	}
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}
