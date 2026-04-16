package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Clean up any .env files created during tests
	m.Run()
	// Clean up test .env files after tests
	_ = os.Remove(".env")
	_ = os.Remove(".env.test")
}

func TestGetDSN(t *testing.T) {
	tests := []struct {
		name          string
		databasePath  string
		expectedDSN   string
	}{
		{
			name:         "default database path",
			databasePath: "./data/tools_bank.db",
			expectedDSN:  "./data/tools_bank.db",
		},
		{
			name:         "custom database path",
			databasePath: "/var/lib/app/database.db",
			expectedDSN:  "/var/lib/app/database.db",
		},
		{
			name:         "in-memory database path",
			databasePath: ":memory:",
			expectedDSN:  ":memory:",
		},
		{
			name:         "empty database path",
			databasePath: "",
			expectedDSN:  "",
		},
		{
			name:         "relative path with spaces",
			databasePath: "./my database/tools.db",
			expectedDSN:  "./my database/tools.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DatabasePath: tt.databasePath,
			}
			assert.Equal(t, tt.expectedDSN, cfg.GetDSN())
		})
	}
}

func TestLoadFromEnv_AllVariablesSet(t *testing.T) {
	// Setup environment variables
	envVars := map[string]string{
		"SUPABASE_URL":    "https://example.supabase.co",
		"SUPABASE_KEY":    "test-key-123",
		"JINA_API_KEY":    "jina-api-key-456",
		"DATABASE_PATH":   "/custom/path/tools.db",
		"LOG_LEVEL":       "debug",
	}

	// Set environment variables
	for key, value := range envVars {
		_ = os.Unsetenv(key)
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	cfg, err := LoadFromEnv()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "https://example.supabase.co", cfg.SupabaseURL)
	assert.Equal(t, "test-key-123", cfg.SupabaseKey)
	assert.Equal(t, "jina-api-key-456", cfg.JinaAPIKey)
	assert.Equal(t, "/custom/path/tools.db", cfg.DatabasePath)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoadFromEnv_PartialVariables(t *testing.T) {
	// Only set some environment variables
	_ = os.Unsetenv("SUPABASE_URL")
	_ = os.Unsetenv("SUPABASE_KEY")
	_ = os.Setenv("JINA_API_KEY", "partial-key")
	_ = os.Unsetenv("DATABASE_PATH")
	_ = os.Setenv("LOG_LEVEL", "WARN")
	defer func() {
		_ = os.Unsetenv("JINA_API_KEY")
		_ = os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := LoadFromEnv()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.SupabaseURL)
	assert.Empty(t, cfg.SupabaseKey)
	assert.Equal(t, "partial-key", cfg.JinaAPIKey)
	assert.Equal(t, "./data/tools_bank.db", cfg.DatabasePath) // Default
	assert.Equal(t, "warn", cfg.LogLevel) // Normalized to lowercase
}

func TestLoadFromEnv_NoVariablesSet(t *testing.T) {
	// Unset all environment variables
	envVars := []string{"SUPABASE_URL", "SUPABASE_KEY", "JINA_API_KEY", "DATABASE_PATH", "LOG_LEVEL"}
	for _, key := range envVars {
		_ = os.Unsetenv(key)
	}

	cfg, err := LoadFromEnv()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.SupabaseURL)
	assert.Empty(t, cfg.SupabaseKey)
	assert.Empty(t, cfg.JinaAPIKey)
	assert.Equal(t, "./data/tools_bank.db", cfg.DatabasePath) // Default
	assert.Equal(t, "info", cfg.LogLevel) // Default
}

func TestLoadFromEnv_EmptyStringValues(t *testing.T) {
	// Set environment variables to empty strings
	envVars := map[string]string{
		"SUPABASE_URL":  "",
		"SUPABASE_KEY":  "",
		"JINA_API_KEY":  "",
		"DATABASE_PATH": "",
		"LOG_LEVEL":     "",
	}
	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	cfg, err := LoadFromEnv()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	// Empty strings should fall back to defaults for DATABASE_PATH and LOG_LEVEL
	assert.Equal(t, "./data/tools_bank.db", cfg.DatabasePath)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLogLevelNormalization(t *testing.T) {
	tests := []struct {
		name          string
		inputLogLevel string
		expected      string
	}{
		{
			name:          "uppercase DEBUG",
			inputLogLevel: "DEBUG",
			expected:      "debug",
		},
		{
			name:          "uppercase INFO",
			inputLogLevel: "INFO",
			expected:      "info",
		},
		{
			name:          "uppercase WARN",
			inputLogLevel: "WARN",
			expected:      "warn",
		},
		{
			name:          "uppercase ERROR",
			inputLogLevel: "ERROR",
			expected:      "error",
		},
		{
			name:          "mixed case Warn",
			inputLogLevel: "Warn",
			expected:      "warn",
		},
		{
			name:          "mixed case Info",
			inputLogLevel: "Info",
			expected:      "info",
		},
		{
			name:          "lowercase debug",
			inputLogLevel: "debug",
			expected:      "debug",
		},
		{
			name:          "lowercase trace",
			inputLogLevel: "trace",
			expected:      "trace",
		},
		{
			name:          "empty string falls back to default",
			inputLogLevel: "",
			expected:      "info",
		},
		{
			name:          "fatal uppercase",
			inputLogLevel: "FATAL",
			expected:      "fatal",
		},
		{
			name:          "warning mixed",
			inputLogLevel: "Warning",
			expected:      "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unset all env vars first
			_ = os.Unsetenv("SUPABASE_URL")
			_ = os.Unsetenv("SUPABASE_KEY")
			_ = os.Unsetenv("JINA_API_KEY")
			_ = os.Unsetenv("DATABASE_PATH")

			if tt.inputLogLevel != "" {
				_ = os.Setenv("LOG_LEVEL", tt.inputLogLevel)
			} else {
				_ = os.Unsetenv("LOG_LEVEL")
			}
			defer func() {
				_ = os.Unsetenv("LOG_LEVEL")
			}()

			cfg, err := LoadFromEnv()

			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, tt.expected, cfg.LogLevel)
		})
	}
}

func TestLoadFromEnv_DotEnvFile(t *testing.T) {
	// Create a temporary .env file
	envContent := `SUPABASE_URL=https://env.supabase.co
SUPABASE_KEY=env-file-key
JINA_API_KEY=env-file-jina
DATABASE_PATH=/env/file/path.db
LOG_LEVEL=trace
`

	tempDir := t.TempDir()
	envFilePath := filepath.Join(tempDir, ".env")
	err := os.WriteFile(envFilePath, []byte(envContent), 0o644)
	assert.NoError(t, err)

	// Verify the file was created
	assert.FileExists(t, envFilePath)

	// Now try to load the .env from the temp directory
	// We need to change to that directory or use godotenv with explicit path
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	_ = os.Chdir(tempDir)

	cfg, err := LoadFromEnv()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "https://env.supabase.co", cfg.SupabaseURL)
	assert.Equal(t, "env-file-key", cfg.SupabaseKey)
	assert.Equal(t, "env-file-jina", cfg.JinaAPIKey)
	assert.Equal(t, "/env/file/path.db", cfg.DatabasePath)
	assert.Equal(t, "trace", cfg.LogLevel)
}

func TestLoadFromEnv_EnvVarsOverrideDotEnvFile(t *testing.T) {
	// Create a temporary .env file
	envContent := `SUPABASE_URL=https://env.supabase.co
SUPABASE_KEY=env-file-key
JINA_API_KEY=env-file-jina
DATABASE_PATH=/env/file/path.db
LOG_LEVEL=trace
`

	tempDir := t.TempDir()
	envFilePath := filepath.Join(tempDir, ".env")
	err := os.WriteFile(envFilePath, []byte(envContent), 0o644)
	assert.NoError(t, err)

	// Change to temp directory where .env exists
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	_ = os.Chdir(tempDir)

	// Set environment variables that should override .env file
	_ = os.Setenv("SUPABASE_URL", "https://env-override.supabase.co")
	_ = os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		_ = os.Unsetenv("SUPABASE_URL")
		_ = os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := LoadFromEnv()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	// Env vars should take precedence
	assert.Equal(t, "https://env-override.supabase.co", cfg.SupabaseURL)
	assert.Equal(t, "env-file-key", cfg.SupabaseKey) // From .env file
	assert.Equal(t, "env-file-jina", cfg.JinaAPIKey)  // From .env file
	assert.Equal(t, "/env/file/path.db", cfg.DatabasePath) // From .env file
	assert.Equal(t, "debug", cfg.LogLevel) // From env var, normalized
}

func TestLoadFromEnv_MissingDotEnvFile(t *testing.T) {
	// Ensure no .env file exists in temp directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	_ = os.Chdir(tempDir)

	// Unset all environment variables
	_ = os.Unsetenv("SUPABASE_URL")
	_ = os.Unsetenv("SUPABASE_KEY")
	_ = os.Unsetenv("JINA_API_KEY")
	_ = os.Unsetenv("DATABASE_PATH")
	_ = os.Unsetenv("LOG_LEVEL")

	cfg, err := LoadFromEnv()

	// Should not error even when .env file is missing
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	// Should use defaults
	assert.Empty(t, cfg.SupabaseURL)
	assert.Empty(t, cfg.SupabaseKey)
	assert.Empty(t, cfg.JinaAPIKey)
	assert.Equal(t, "./data/tools_bank.db", cfg.DatabasePath)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoadFromEnv_ConfigStructFields(t *testing.T) {
	tests := []struct {
		name          string
		envSetup      func()
		envTeardown   func()
		expectedCfg   *Config
	}{
		{
			name: "all fields populated",
			envSetup: func() {
				_ = os.Setenv("SUPABASE_URL", "https://all.supabase.co")
				_ = os.Setenv("SUPABASE_KEY", "all-key")
				_ = os.Setenv("JINA_API_KEY", "all-jina")
				_ = os.Setenv("DATABASE_PATH", "/all/path.db")
				_ = os.Setenv("LOG_LEVEL", "error")
			},
			envTeardown: func() {
				_ = os.Unsetenv("SUPABASE_URL")
				_ = os.Unsetenv("SUPABASE_KEY")
				_ = os.Unsetenv("JINA_API_KEY")
				_ = os.Unsetenv("DATABASE_PATH")
				_ = os.Unsetenv("LOG_LEVEL")
			},
			expectedCfg: &Config{
				SupabaseURL:  "https://all.supabase.co",
				SupabaseKey:  "all-key",
				JinaAPIKey:   "all-jina",
				DatabasePath: "/all/path.db",
				LogLevel:     "error",
			},
		},
		{
			name: "all fields empty",
			envSetup: func() {
				// All unset
			},
			envTeardown: func() {
				_ = os.Unsetenv("SUPABASE_URL")
				_ = os.Unsetenv("SUPABASE_KEY")
				_ = os.Unsetenv("JINA_API_KEY")
				_ = os.Unsetenv("DATABASE_PATH")
				_ = os.Unsetenv("LOG_LEVEL")
			},
			expectedCfg: &Config{
				SupabaseURL:  "",
				SupabaseKey:  "",
				JinaAPIKey:   "",
				DatabasePath: "./data/tools_bank.db",
				LogLevel:     "info",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment first
			_ = os.Unsetenv("SUPABASE_URL")
			_ = os.Unsetenv("SUPABASE_KEY")
			_ = os.Unsetenv("JINA_API_KEY")
			_ = os.Unsetenv("DATABASE_PATH")
			_ = os.Unsetenv("LOG_LEVEL")

			// Setup test environment
			tt.envSetup()
			defer tt.envTeardown()

			cfg, err := LoadFromEnv()

			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, tt.expectedCfg.SupabaseURL, cfg.SupabaseURL)
			assert.Equal(t, tt.expectedCfg.SupabaseKey, cfg.SupabaseKey)
			assert.Equal(t, tt.expectedCfg.JinaAPIKey, cfg.JinaAPIKey)
			assert.Equal(t, tt.expectedCfg.DatabasePath, cfg.DatabasePath)
			assert.Equal(t, tt.expectedCfg.LogLevel, cfg.LogLevel)
		})
	}
}

func TestGetEnv_WithValue(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		defaultVal  string
		expected    string
	}{
		{
			name:       "returns environment value",
			key:        "TEST_GETENV_VALUE",
			value:      "my-test-value",
			defaultVal: "default-value",
			expected:   "my-test-value",
		},
		{
			name:       "returns environment value even if default is same",
			key:        "TEST_GETENV_SAME",
			value:      "value",
			defaultVal: "value",
			expected:   "value",
		},
		{
			name:       "returns environment value with spaces",
			key:        "TEST_GETENV_SPACES",
			value:      "value with spaces",
			defaultVal: "default",
			expected:   "value with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv(tt.key, tt.value)
			defer func() {
				_ = os.Unsetenv(tt.key)
			}()

			result := getEnv(tt.key, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnv_WithDefault(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal string
		expected   string
	}{
		{
			name:       "returns default for unset key",
			key:        "UNSET_TEST_KEY_12345",
			defaultVal: "default-value",
			expected:   "default-value",
		},
		{
			name:       "returns empty default for unset key",
			key:        "UNSET_TEST_KEY_67890",
			defaultVal: "",
			expected:   "",
		},
		{
			name:       "returns default for non-existent key",
			key:        "NONEXISTENT_KEY_ABCDE",
			defaultVal: "fallback",
			expected:   "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv(tt.key) // Ensure key is not set

			result := getEnv(tt.key, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnv_EmptyStringValue(t *testing.T) {
	key := "EMPTY_VALUE_TEST_KEY"
	_ = os.Setenv(key, "")
	defer func() {
		_ = os.Unsetenv(key)
	}()

	// Empty string should be treated as "not set" and return default
	result := getEnv(key, "my-default")
	assert.Equal(t, "my-default", result)
}

func TestLoadFromEnv_DefaultValues(t *testing.T) {
	// Ensure no environment variables are set
	_ = os.Unsetenv("SUPABASE_URL")
	_ = os.Unsetenv("SUPABASE_KEY")
	_ = os.Unsetenv("JINA_API_KEY")
	_ = os.Unsetenv("DATABASE_PATH")
	_ = os.Unsetenv("LOG_LEVEL")

	cfg, err := LoadFromEnv()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify default values
	assert.Equal(t, "./data/tools_bank.db", cfg.DatabasePath, "DATABASE_PATH should default to ./data/tools_bank.db")
	assert.Equal(t, "info", cfg.LogLevel, "LOG_LEVEL should default to info")
	assert.Empty(t, cfg.SupabaseURL, "SUPABASE_URL should default to empty")
	assert.Empty(t, cfg.SupabaseKey, "SUPABASE_KEY should default to empty")
	assert.Empty(t, cfg.JinaAPIKey, "JINA_API_KEY should default to empty")
}

func TestLoadFromEnv_LogLevelCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected string
	}{
		{"DEBUG", "DEBUG", "debug"},
		{"Info", "Info", "info"},
		{"WARN", "WARN", "warn"},
		{"Error", "Error", "error"},
		{"Trace", "Trace", "trace"},
		{"Debug", "Debug", "debug"},
		{"INFO", "INFO", "info"},
		{"warn", "warn", "warn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			_ = os.Unsetenv("SUPABASE_URL")
			_ = os.Unsetenv("SUPABASE_KEY")
			_ = os.Unsetenv("JINA_API_KEY")
			_ = os.Unsetenv("DATABASE_PATH")

			_ = os.Setenv("LOG_LEVEL", tt.logLevel)
			defer func() {
				_ = os.Unsetenv("LOG_LEVEL")
			}()

			cfg, err := LoadFromEnv()

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.LogLevel)
		})
	}
}

func TestLoadFromEnv_SpecialCharactersInValues(t *testing.T) {
	// Clean environment
	_ = os.Unsetenv("SUPABASE_URL")
	_ = os.Unsetenv("SUPABASE_KEY")
	_ = os.Unsetenv("JINA_API_KEY")
	_ = os.Unsetenv("DATABASE_PATH")
	_ = os.Unsetenv("LOG_LEVEL")

	// Set values with special characters
	_ = os.Setenv("SUPABASE_URL", "https://example.com?key=value&other=123")
	_ = os.Setenv("SUPABASE_KEY", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	_ = os.Setenv("JINA_API_KEY", "jina-api-key-with-special=chars")
	defer func() {
		_ = os.Unsetenv("SUPABASE_URL")
		_ = os.Unsetenv("SUPABASE_KEY")
		_ = os.Unsetenv("JINA_API_KEY")
	}()

	cfg, err := LoadFromEnv()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "https://example.com?key=value&other=123", cfg.SupabaseURL)
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", cfg.SupabaseKey)
	assert.Equal(t, "jina-api-key-with-special=chars", cfg.JinaAPIKey)
}

func TestLoadFromEnv_WhitespaceHandling(t *testing.T) {
	// Test that whitespace is preserved (not trimmed by getEnv)
	tests := []struct {
		name     string
		logLevel string
		expected string
	}{
		{
			name:     "lowercase preserved",
			logLevel: "info",
			expected: "info",
		},
		{
			name:     "leading whitespace preserved",
			logLevel: " debug",
			expected: " debug", // whitespace is NOT trimmed, only lowercased
		},
		{
			name:     "trailing whitespace preserved",
			logLevel: "debug ",
			expected: "debug ", // whitespace is NOT trimmed, only lowercased
		},
		{
			name:     "both leading and trailing whitespace preserved",
			logLevel: " debug ",
			expected: " debug ", // whitespace is NOT trimmed, only lowercased
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv("SUPABASE_URL")
			_ = os.Unsetenv("SUPABASE_KEY")
			_ = os.Unsetenv("JINA_API_KEY")
			_ = os.Unsetenv("DATABASE_PATH")

			_ = os.Setenv("LOG_LEVEL", tt.logLevel)
			defer func() {
				_ = os.Unsetenv("LOG_LEVEL")
			}()

			cfg, err := LoadFromEnv()

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.LogLevel)
		})
	}
}

func TestLoadFromEnv_NilErrorReturn(t *testing.T) {
	// Ensure LoadFromEnv always returns nil error (current implementation)
	_ = os.Unsetenv("SUPABASE_URL")
	_ = os.Unsetenv("SUPABASE_KEY")
	_ = os.Unsetenv("JINA_API_KEY")
	_ = os.Unsetenv("DATABASE_PATH")
	_ = os.Unsetenv("LOG_LEVEL")

	cfg, err := LoadFromEnv()

	assert.NoError(t, err, "LoadFromEnv should not return an error")
	assert.NotNil(t, cfg, "LoadFromEnv should always return a valid Config pointer")
}

// Test getEnv indirectly through LoadFromEnv
func TestLoadFromEnv_GetEnvIntegration(t *testing.T) {
	// Test the getEnv helper indirectly by verifying LoadFromEnv behavior
	// when environment variables are set vs unset

	tests := []struct {
		name            string
		envKey          string
		envValue        string
		configField     string
		expectedValue   string
	}{
		{
			name:            "SUPABASE_URL set",
			envKey:          "SUPABASE_URL",
			envValue:        "https://set.supabase.co",
			configField:     "SupabaseURL",
			expectedValue:   "https://set.supabase.co",
		},
		{
			name:            "SUPABASE_KEY set",
			envKey:          "SUPABASE_KEY",
			envValue:        "key-set",
			configField:     "SupabaseKey",
			expectedValue:   "key-set",
		},
		{
			name:            "JINA_API_KEY set",
			envKey:          "JINA_API_KEY",
			envValue:        "jina-set",
			configField:     "JinaAPIKey",
			expectedValue:   "jina-set",
		},
		{
			name:            "DATABASE_PATH set",
			envKey:          "DATABASE_PATH",
			envValue:        "/custom/set/path.db",
			configField:     "DatabasePath",
			expectedValue:   "/custom/set/path.db",
		},
		{
			name:            "LOG_LEVEL set",
			envKey:          "LOG_LEVEL",
			envValue:        "warning",
			configField:     "LogLevel",
			expectedValue:   "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			_ = os.Unsetenv("SUPABASE_URL")
			_ = os.Unsetenv("SUPABASE_KEY")
			_ = os.Unsetenv("JINA_API_KEY")
			_ = os.Unsetenv("DATABASE_PATH")
			_ = os.Unsetenv("LOG_LEVEL")

			_ = os.Setenv(tt.envKey, tt.envValue)
			defer func() {
				_ = os.Unsetenv(tt.envKey)
			}()

			cfg, err := LoadFromEnv()
			assert.NoError(t, err)

			// Use reflection-like approach with strings
			var got string
			switch tt.configField {
			case "SupabaseURL":
				got = cfg.SupabaseURL
			case "SupabaseKey":
				got = cfg.SupabaseKey
			case "JinaAPIKey":
				got = cfg.JinaAPIKey
			case "DatabasePath":
				got = cfg.DatabasePath
			case "LogLevel":
				got = cfg.LogLevel
			}

			assert.Equal(t, tt.expectedValue, got)
		})
	}
}

func TestConfigStruct_NoInitialization(t *testing.T) {
	// Verify that an uninitialized Config struct has empty string fields
	cfg := Config{}

	assert.Empty(t, cfg.SupabaseURL)
	assert.Empty(t, cfg.SupabaseKey)
	assert.Empty(t, cfg.JinaAPIKey)
	assert.Empty(t, cfg.DatabasePath)
	assert.Empty(t, cfg.LogLevel)
}

func TestConfigStruct_WithInitialization(t *testing.T) {
	cfg := Config{
		SupabaseURL:  "https://test.supabase.co",
		SupabaseKey:  "test-key",
		JinaAPIKey:   "test-jina",
		DatabasePath: "/test/path.db",
		LogLevel:     "debug",
	}

	assert.Equal(t, "https://test.supabase.co", cfg.SupabaseURL)
	assert.Equal(t, "test-key", cfg.SupabaseKey)
	assert.Equal(t, "test-jina", cfg.JinaAPIKey)
	assert.Equal(t, "/test/path.db", cfg.DatabasePath)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoadFromEnv_ConcurrentSafety(t *testing.T) {
	// Test that LoadFromEnv can be called concurrently without issues
	_ = os.Setenv("SUPABASE_URL", "concurrent-test")
	_ = os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		_ = os.Unsetenv("SUPABASE_URL")
		_ = os.Unsetenv("LOG_LEVEL")
	}()

	const goroutineCount = 10
	done := make(chan bool, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func() {
			cfg, err := LoadFromEnv()
			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, "concurrent-test", cfg.SupabaseURL)
			assert.Equal(t, "debug", cfg.LogLevel)
			done <- true
		}()
	}

	for i := 0; i < goroutineCount; i++ {
		<-done
	}
}

// Helper function to check if a string contains substrings (for debugging)
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
