package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds environment-based configuration for the application.
type Config struct {
	SupabaseURL    string
	SupabaseKey    string
	JinaAPIKey     string
	DatabasePath   string
	LogLevel       string
}

// LoadFromEnv loads configuration from environment variables.
// It also attempts to load a .env file if present.
// Environment variables take precedence over .env file values.
func LoadFromEnv() (*Config, error) {
	// Try to load .env file silently (ignore errors if file not found)
	_ = godotenv.Load()

	cfg := &Config{
		SupabaseURL:    getEnv("SUPABASE_URL", ""),
		SupabaseKey:    getEnv("SUPABASE_KEY", ""),
		JinaAPIKey:     getEnv("JINA_API_KEY", ""),
		DatabasePath:   getEnv("DATABASE_PATH", "./data/tools_bank.db"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	// Normalize log level to lowercase
	cfg.LogLevel = strings.ToLower(cfg.LogLevel)

	return cfg, nil
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// GetDSN returns the database connection string/path.
func (c *Config) GetDSN() string {
	return c.DatabasePath
}
