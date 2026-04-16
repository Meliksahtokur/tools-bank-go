package utils

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Level Tests
// ============================================================================

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Level
	}{
		{"debug", "debug", LevelDebug},
		{"info", "info", LevelInfo},
		{"warn", "warn", LevelWarn},
		{"warning", "warning", LevelWarn},
		{"error", "error", LevelError},
		{"empty defaults to info", "", LevelInfo},
		{"unknown defaults to info", "random", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger("debug")
	assert.Equal(t, LevelDebug, logger.level)
	assert.NotNil(t, logger.fields)
	assert.NotNil(t, logger.output)
}

func TestLoggerWithFields(t *testing.T) {
	logger := NewLogger("info")
	
	result := logger.WithFields(map[string]any{"key": "value"})
	assert.Equal(t, "value", result.fields["key"])
	
	// Original unchanged
	assert.Nil(t, logger.fields["key"])
}

// ============================================================================
// Logger Output Tests - Using temp file
// ============================================================================

func createTempLogger(t *testing.T, level string) (*Logger, func()) {
	t.Helper()
	
	// Create temp file
	tmpFile, err := os.CreateTemp("", "logger_test*.log")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	
	// Open for writing
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_TRUNC, 0644)
	require.NoError(t, err)
	
	logger := NewLogger(level)
	origOutput := logger.output
	logger.output = f
	
	cleanup := func() {
		f.Close()
		os.Remove(tmpPath)
		logger.output = origOutput
	}
	
	return logger, cleanup
}

func readTempFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func TestLoggerInfo(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	logger.Info("test info message")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)
	
	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "test info message", entry.Message)
	assert.NotEmpty(t, entry.Timestamp)
}

func TestLoggerWarn(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	logger.Warn("test warn message")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)
	
	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	assert.Equal(t, "WARN", entry.Level)
	assert.Equal(t, "test warn message", entry.Message)
}

func TestLoggerError(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	logger.Error("test error message")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)
	
	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	assert.Equal(t, "ERROR", entry.Level)
	assert.Equal(t, "test error message", entry.Message)
}

func TestLoggerDebugBelowThreshold(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	logger.Debug("should not appear")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	assert.Empty(t, content)
}

func TestLoggerDebugWithDebugLevel(t *testing.T) {
	logger, cleanup := createTempLogger(t, "debug")
	defer cleanup()
	
	logger.Debug("debug message")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)
	
	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	assert.Equal(t, "DEBUG", entry.Level)
	assert.Equal(t, "debug message", entry.Message)
}

func TestLoggerLogRespectsMinimumLevel(t *testing.T) {
	t.Run("error logger ignores info/warn/debug", func(t *testing.T) {
		logger, cleanup := createTempLogger(t, "error")
		defer cleanup()
		
		logger.Info("info msg")
		logger.Warn("warn msg")
		logger.Debug("debug msg")
		
		content, err := readTempFile(logger.output.Name())
		require.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("error logger shows error", func(t *testing.T) {
		logger, cleanup := createTempLogger(t, "error")
		defer cleanup()
		
		logger.Error("error msg")
		
		content, err := readTempFile(logger.output.Name())
		require.NoError(t, err)
		require.NotEmpty(t, content)
	})
}

// ============================================================================
// Package-level DefaultLogger Tests
// ============================================================================

func TestPackageLevelInfo(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	original := DefaultLogger
	DefaultLogger = logger
	defer func() { DefaultLogger = original }()
	
	Info("package info message")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)
	
	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "package info message", entry.Message)
}

func TestPackageLevelWarn(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	original := DefaultLogger
	DefaultLogger = logger
	defer func() { DefaultLogger = original }()
	
	Warn("package warn message")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)
	
	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	assert.Equal(t, "WARN", entry.Level)
}

func TestPackageLevelError(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	original := DefaultLogger
	DefaultLogger = logger
	defer func() { DefaultLogger = original }()
	
	Error("package error message")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)
	
	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	assert.Equal(t, "ERROR", entry.Level)
}

func TestPackageLevelDebugShown(t *testing.T) {
	logger, cleanup := createTempLogger(t, "debug")
	defer cleanup()
	
	original := DefaultLogger
	DefaultLogger = logger
	defer func() { DefaultLogger = original }()
	
	Debug("package debug message")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)
}

func TestSetLevel(t *testing.T) {
	origLevel := DefaultLogger.level
	
	t.Run("set to debug", func(t *testing.T) {
		SetLevel("debug")
		assert.Equal(t, LevelDebug, DefaultLogger.level)
	})
	
	t.Run("set to info", func(t *testing.T) {
		SetLevel("info")
		assert.Equal(t, LevelInfo, DefaultLogger.level)
	})
	
	t.Run("set to error", func(t *testing.T) {
		SetLevel("error")
		assert.Equal(t, LevelError, DefaultLogger.level)
	})
	
	t.Run("invalid defaults to info", func(t *testing.T) {
		SetLevel("invalid")
		assert.Equal(t, LevelInfo, DefaultLogger.level)
	})
	
	DefaultLogger.level = origLevel
}

// ============================================================================
// Entry JSON Marshaling Tests
// ============================================================================

func TestEntryJSONFormat(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	logger.Info("json test", map[string]any{"count": 42, "name": "test"})
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)

	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	
	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "json test", entry.Message)
	assert.NotEmpty(t, entry.Timestamp)
	assert.NotEmpty(t, entry.Caller)
	assert.Equal(t, float64(42), entry.Fields["count"])
	assert.Equal(t, "test", entry.Fields["name"])
}

func TestLoggerWithFieldsAndInlineFields(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	logger.fields = map[string]any{"service": "test"}
	defer cleanup()
	
	logger.Info("with inline", map[string]any{"request_id": "123"})
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)

	var entry Entry
	json.Unmarshal([]byte(content), &entry)
	
	assert.Equal(t, "test", entry.Fields["service"])
	assert.Equal(t, "123", entry.Fields["request_id"])
}

func TestLoggerCallerInfo(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	logger.Info("caller test")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)

	var entry Entry
	json.Unmarshal([]byte(content), &entry)
	assert.NotEmpty(t, entry.Caller)
	assert.Contains(t, entry.Caller, "logger_test.go")
}

func TestLoggerTimestampFormat(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	logger.Info("timestamp test")
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)

	var entry Entry
	err = json.Unmarshal([]byte(content), &entry)
	require.NoError(t, err)
	
	assert.NotEmpty(t, entry.Timestamp)
	assert.Contains(t, entry.Timestamp, "T")
	assert.Contains(t, entry.Timestamp, "Z")
}

func TestMergeFields(t *testing.T) {
	logger, cleanup := createTempLogger(t, "info")
	defer cleanup()
	
	logger.Info("fields test", map[string]any{"a": 1, "b": 2})
	
	content, err := readTempFile(logger.output.Name())
	require.NoError(t, err)
	require.NotEmpty(t, content)

	var entry Entry
	json.Unmarshal([]byte(content), &entry)
	
	assert.Equal(t, float64(1), entry.Fields["a"])
	assert.Equal(t, float64(2), entry.Fields["b"])
}
