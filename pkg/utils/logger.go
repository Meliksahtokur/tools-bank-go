package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Level represents logging level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a string to a Level.
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Entry represents a structured log entry.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Caller    string `json:"caller,omitempty"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// Logger provides structured logging with JSON output.
type Logger struct {
	mu       sync.Mutex
	level    Level
	fields   map[string]any
	output   *os.File
}

// NewLogger creates a new Logger with the specified minimum log level.
func NewLogger(level string) *Logger {
	return &Logger{
		level:  ParseLevel(level),
		fields: make(map[string]any),
		output: os.Stdout,
	}
}

// WithFields returns a new Logger with additional default fields.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	newFields := make(map[string]any)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &Logger{
		level:  l.level,
		fields: newFields,
		output: l.output,
	}
}

// log writes a log entry with the given level and message.
func (l *Logger) log(level Level, msg string, fields map[string]any) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	caller := ""
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	// Merge fields
	mergedFields := make(map[string]any)
	for k, v := range l.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}

	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   msg,
		Caller:    caller,
		Fields:    mergedFields,
	}

	if len(mergedFields) == 0 {
		entry.Fields = nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal log entry: %v\n", err)
		return
	}

	fmt.Fprintln(l.output, string(data))
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...map[string]any) {
	f := mergeFields(fields)
	l.log(LevelInfo, msg, f)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...map[string]any) {
	f := mergeFields(fields)
	l.log(LevelWarn, msg, f)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...map[string]any) {
	f := mergeFields(fields)
	l.log(LevelError, msg, f)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...map[string]any) {
	f := mergeFields(fields)
	l.log(LevelDebug, msg, f)
}

// mergeFields merges a slice of field maps into a single map.
func mergeFields(fields []map[string]any) map[string]any {
	if len(fields) == 0 {
		return nil
	}
	result := make(map[string]any)
	for _, f := range fields {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}

// DefaultLogger is the package-level default logger.
var DefaultLogger = NewLogger("info")

// Info logs an info message using the default logger.
func Info(msg string, fields ...map[string]any) {
	DefaultLogger.Info(msg, fields...)
}

// Warn logs a warning message using the default logger.
func Warn(msg string, fields ...map[string]any) {
	DefaultLogger.Warn(msg, fields...)
}

// Error logs an error message using the default logger.
func Error(msg string, fields ...map[string]any) {
	DefaultLogger.Error(msg, fields...)
}

// Debug logs a debug message using the default logger.
func Debug(msg string, fields ...map[string]any) {
	DefaultLogger.Debug(msg, fields...)
}

// SetLevel sets the minimum log level for the default logger.
func SetLevel(level string) {
	DefaultLogger.level = ParseLevel(level)
}
