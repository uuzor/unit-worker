package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the UserService
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Cache    CacheConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port             int
	GRPCPort         int
	Environment      string
	ShutdownTimeout  time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	MaxRequestSize   int64
}

// DatabaseConfig holds database configuration
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

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled            bool
	ProfileTTL         time.Duration
	FollowerListTTL    time.Duration
	FollowingListTTL   time.Duration
	UserStatsTTL       time.Duration
	SearchResultsTTL   time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:             getEnvAsInt("SERVER_PORT", 8081),
			GRPCPort:         getEnvAsInt("GRPC_PORT", 9001),
			Environment:      getEnv("ENVIRONMENT", "development"),
			ShutdownTimeout:  getEnvAsDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
			ReadTimeout:      getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout:     getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
			MaxRequestSize:   getEnvAsInt64("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
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
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Cache: CacheConfig{
			Enabled:            getEnvAsBool("CACHE_ENABLED", true),
			ProfileTTL:         getEnvAsDuration("CACHE_PROFILE_TTL", 5*time.Minute),
			FollowerListTTL:    getEnvAsDuration("CACHE_FOLLOWER_LIST_TTL", 2*time.Minute),
			FollowingListTTL:   getEnvAsDuration("CACHE_FOLLOWING_LIST_TTL", 2*time.Minute),
			UserStatsTTL:       getEnvAsDuration("CACHE_USER_STATS_TTL", 1*time.Minute),
			SearchResultsTTL:   getEnvAsDuration("CACHE_SEARCH_RESULTS_TTL", 5*time.Minute),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Server.GRPCPort <= 0 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", c.Server.GRPCPort)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	return nil
}

// DatabaseDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// RedisAddr returns the Redis address
func (c *RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Helper functions to read environment variables

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

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
