package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/egesut/tools-bank-go/pkg/utils"
)

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
}

// Open opens a SQLite database at the given path.
// Creates the directory and file if they don't exist.
func Open(path string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Open database connection
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.InitSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	utils.Info("database opened", map[string]any{"path": path})
	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// InitSchema creates the required tables if they don't exist.
func (db *DB) InitSchema() error {
	schemas := []string{
		// tasks table for task management
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			metadata TEXT
		)`,

		// memory table for storing arbitrary data/memory
		`CREATE TABLE IF NOT EXISTS memory (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			value TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			tags TEXT
		)`,

		// embeddings table for vector embeddings
		`CREATE TABLE IF NOT EXISTS embeddings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			document_id TEXT NOT NULL,
			content TEXT NOT NULL,
			embedding BLOB,
			model TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			metadata TEXT
		)`,

		// Indexes for better query performance
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_created ON tasks(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_memory_key ON memory(key)`,
		`CREATE INDEX IF NOT EXISTS idx_memory_expires ON memory(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_embeddings_doc ON embeddings(document_id)`,
	}

	for _, schema := range schemas {
		if _, err := db.conn.Exec(schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	utils.Info("database schema initialized")
	return nil
}

// Exec executes a query that doesn't return rows.
// Returns the result summary and any error encountered.
func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Query executes a query that returns rows.
// Returns the rows and any error encountered.
func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow executes a query that returns at most one row.
// Returns the row and any error encountered.
func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// Begin starts a new transaction.
func (db *DB) Begin() (*sql.Tx, error) {
	return db.conn.Begin()
}

// Ping checks if the database connection is alive.
func (db *DB) Ping() error {
	return db.conn.Ping()
}
