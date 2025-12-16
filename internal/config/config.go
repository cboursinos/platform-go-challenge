package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	NewRelic  NewRelicConfig
	JWT       JWTConfig
	App       AppConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// MySQLConfig holds MySQL database configuration
type MySQLConfig struct {
	DSN string
}

// RedisConfig holds Redis cache configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
}

// NewRelicConfig holds New Relic APM configuration
type NewRelicConfig struct {
	LicenseKey string
	AppName    string
	Enabled    bool
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret string
}

// AppConfig holds general application configuration
type AppConfig struct {
	Environment string
	Debug       bool
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		MySQL: MySQLConfig{
			DSN: getEnv("MYSQL_DSN", "root:password@tcp(localhost:3306)/favorites_db?charset=utf8mb4&parseTime=True&loc=Local"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TTL:      15 * time.Minute,
		},
		NewRelic: NewRelicConfig{
			LicenseKey: getEnv("NEW_RELIC_LICENSE_KEY", ""),
			AppName:    getEnv("NEW_RELIC_APP_NAME", "Favorites API"),
			Enabled:    getEnv("NEW_RELIC_LICENSE_KEY", "") != "" && getEnv("NEW_RELIC_APP_NAME", "") != "",
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		App: AppConfig{
			Environment: getEnv("ENVIRONMENT", "development"),
			Debug:       getEnvAsBool("DEBUG", false),
		},
	}

	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
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

// getEnvAsBool gets an environment variable as boolean or returns a default value
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

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if c.MySQL.DSN == "" {
		return fmt.Errorf("MYSQL_DSN is required")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("REDIS_ADDR is required")
	}
	return nil
}

// String returns a string representation of the config (without sensitive data)
func (c *Config) String() string {
	return fmt.Sprintf(`
Configuration:
  Server:
    Port: %s
  MySQL:
    DSN: %s (masked)
  Redis:
    Addr: %s
    DB: %d
  New Relic:
    Enabled: %v
    App Name: %s
  App:
    Environment: %s
    Debug: %v
`,
		c.Server.Port,
		maskDSN(c.MySQL.DSN),
		c.Redis.Addr,
		c.Redis.DB,
		c.NewRelic.Enabled,
		c.NewRelic.AppName,
		c.App.Environment,
		c.App.Debug,
	)
}

// maskDSN masks password in DSN for logging
func maskDSN(dsn string) string {
	// Simple masking - replace password with ***
	// This is a basic implementation
	if len(dsn) > 20 {
		return dsn[:10] + "***" + dsn[len(dsn)-20:]
	}
	return "***"
}



