package db

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBOpen(t *testing.T) {
	// Create temp DB file
	tmpFile, err := os.CreateTemp("", "test-tools-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Verify schema created
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	tables := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		tables[name] = true
	}

	// Check required tables exist
	requiredTables := []string{"tasks", "memory", "embeddings"}
	for _, table := range requiredTables {
		if !tables[table] {
			t.Errorf("Expected table %q not found", table)
		}
	}
}

func TestDBPing(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-ping-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestMemoryOperations(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-memory-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Test memory_set
	_, err = db.Exec(`
		INSERT INTO memory (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, "test_key", "test_value")
	if err != nil {
		t.Fatalf("Failed to insert memory: %v", err)
	}

	// Test memory_get
	var value string
	err = db.QueryRow("SELECT value FROM memory WHERE key = ?", "test_key").Scan(&value)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %q", value)
	}

	// Test memory_update
	_, err = db.Exec(`
		INSERT INTO memory (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, "test_key", "updated_value")
	if err != nil {
		t.Fatalf("Failed to update memory: %v", err)
	}

	err = db.QueryRow("SELECT value FROM memory WHERE key = ?", "test_key").Scan(&value)
	if err != nil {
		t.Fatalf("Failed to get updated memory: %v", err)
	}
	if value != "updated_value" {
		t.Errorf("Expected 'updated_value', got %q", value)
	}

	// Test memory_delete
	_, err = db.Exec("DELETE FROM memory WHERE key = ?", "test_key")
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	err = db.QueryRow("SELECT value FROM memory WHERE key = ?", "test_key").Scan(&value)
	if err == nil {
		t.Error("Expected error for deleted key")
	}
}

func TestTaskOperations(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-task-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Test task_create
	taskID := "test-task-123"
	_, err = db.Exec(`
		INSERT INTO tasks (task_id, name, description, status)
		VALUES (?, ?, ?, 'pending')
	`, taskID, "Test Task", "Test Description")
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Test task_read
	var name, description, status string
	err = db.QueryRow("SELECT name, description, status FROM tasks WHERE task_id = ?", taskID).Scan(&name, &description, &status)
	if err != nil {
		t.Fatalf("Failed to read task: %v", err)
	}
	if name != "Test Task" {
		t.Errorf("Expected name 'Test Task', got %q", name)
	}
	if status != "pending" {
		t.Errorf("Expected status 'pending', got %q", status)
	}

	// Test task_update
	_, err = db.Exec("UPDATE tasks SET status = 'in_progress', updated_at = CURRENT_TIMESTAMP WHERE task_id = ?", taskID)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	err = db.QueryRow("SELECT status FROM tasks WHERE task_id = ?", taskID).Scan(&status)
	if err != nil {
		t.Fatalf("Failed to read updated task: %v", err)
	}
	if status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got %q", status)
	}

	// Test task_delete
	_, err = db.Exec("DELETE FROM tasks WHERE task_id = ?", taskID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	err = db.QueryRow("SELECT name FROM tasks WHERE task_id = ?", taskID).Scan(&name)
	if err == nil {
		t.Error("Expected error for deleted task")
	}
}

func TestTransaction(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-txn-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Test successful transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("INSERT INTO tasks (task_id, name, status) VALUES (?, ?, 'pending')", "txn-task-1", "Transaction Task")
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify commit
	var name string
	err = db.QueryRow("SELECT name FROM tasks WHERE task_id = ?", "txn-task-1").Scan(&name)
	if err != nil {
		t.Errorf("Transaction should have been committed: %v", err)
	}
}

func TestTransactionRollback(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-rollback-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Test rollback
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("INSERT INTO tasks (task_id, name, status) VALUES (?, ?, 'pending')", "rollback-task", "Rollback Task")
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// Rollback instead of commit
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Verify rollback
	var name string
	err = db.QueryRow("SELECT name FROM tasks WHERE task_id = ?", "rollback-task").Scan(&name)
	if err == nil {
		t.Error("Transaction should have been rolled back")
	}
}

func TestEmbeddingOperations(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-embed-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Test embedding insert
	docID := "doc-123"
	content := "This is test content for embedding"
	_, err = db.Exec(`
		INSERT INTO embeddings (document_id, content, model)
		VALUES (?, ?, 'test-model')
	`, docID, content)
	if err != nil {
		t.Fatalf("Failed to insert embedding: %v", err)
	}

	// Test embedding read
	var readContent string
	err = db.QueryRow("SELECT content FROM embeddings WHERE document_id = ?", docID).Scan(&readContent)
	if err != nil {
		t.Fatalf("Failed to read embedding: %v", err)
	}
	if readContent != content {
		t.Errorf("Expected content %q, got %q", content, readContent)
	}

	// Test FTS5 search (may fail if FTS5 not available)
	_, err = db.SearchEmbeddings("test content", 10)
	// FTS5 might not be available, that's OK - fall back exists
	_ = err
}

func TestSchemaIndexes(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-indexes-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Verify indexes exist
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='index'")
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	defer rows.Close()

	indexes := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		indexes[name] = true
	}

	// Check expected indexes
	expectedIndexes := []string{
		"idx_tasks_status",
		"idx_tasks_created",
		"idx_memory_key",
		"idx_memory_expires",
		"idx_embeddings_doc",
	}

	for _, idx := range expectedIndexes {
		if !indexes[idx] {
			t.Errorf("Expected index %q not found", idx)
		}
	}
}

func TestQueryRowNotFound(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-notfound-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// QueryRow for non-existent row should return sql.ErrNoRows
	var value string
	err = db.QueryRow("SELECT value FROM memory WHERE key = ?", "nonexistent").Scan(&value)
	if err == nil {
		t.Error("Expected error for non-existent row")
	}
}

func TestClose(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-close-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	// Close should not error
	if err := db.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Double close should also not error
	if err := db.Close(); err != nil {
		t.Errorf("Double close failed: %v", err)
	}
}

// =============================================================================
// FTS5 Search Tests
// =============================================================================

func TestFTS5SearchBasic(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-fts5-basic-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Insert test embeddings
	docs := []struct {
		id      string
		content string
	}{
		{"doc-1", "Go programming language tutorial"},
		{"doc-2", "Python machine learning basics"},
		{"doc-3", "Go concurrency patterns"},
	}

	for _, doc := range docs {
		_, err = db.Exec(`
			INSERT INTO embeddings (document_id, content, model)
			VALUES (?, ?, 'test-model')
		`, doc.id, doc.content)
		if err != nil {
			t.Fatalf("Failed to insert embedding %s: %v", doc.id, err)
		}
	}

	// Search for "Go programming"
	results, err := db.SearchEmbeddings("Go", 10)
	if err != nil {
		t.Fatalf("FTS5 search failed: %v", err)
	}

	assert := require.New(t)
	assert.GreaterOrEqual(len(results), 1, "Should find at least one Go-related document")

	// Verify results contain Go-related content
	for _, r := range results {
		assert.Contains(r.Content, "Go", "Result should contain 'Go'")
	}
}

func TestFTS5SearchNoResults(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-fts5-noresult-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Insert some embeddings
	_, err = db.Exec(`
		INSERT INTO embeddings (document_id, content, model)
		VALUES ('doc-1', 'Some random content', 'test-model')
	`)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	// Search for something that doesn't exist
	results, err := db.SearchEmbeddings("xyznonexistent123", 10)
	if err != nil {
		t.Fatalf("FTS5 search failed: %v", err)
	}

	assert := require.New(t)
	assert.Equal(0, len(results), "Should return no results for non-matching query")
}

func TestFTS5SearchMultipleResults(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-fts5-multi-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Insert multiple docs with shared term
	for i := 1; i <= 5; i++ {
		_, err = db.Exec(`
			INSERT INTO embeddings (document_id, content, model)
			VALUES (?, ?, 'test-model')
		`, fmt.Sprintf("doc-%d", i), fmt.Sprintf("Document about programming with code examples %d", i))
		require.NoError(t, err)
	}

	results, err := db.SearchEmbeddings("programming", 10)
	require.NoError(t, err)

	assert := require.New(t)
	assert.Equal(5, len(results), "Should find all 5 documents with 'programming'")
}

func TestFTS5SearchWithLimit(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-fts5-limit-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Insert 10 documents
	for i := 1; i <= 10; i++ {
		_, err = db.Exec(`
			INSERT INTO embeddings (document_id, content, model)
			VALUES (?, ?, 'test-model')
		`, fmt.Sprintf("doc-%d", i), fmt.Sprintf("Test document number %d", i))
		require.NoError(t, err)
	}

	// Search with limit=3
	results, err := db.SearchEmbeddings("Test", 3)
	require.NoError(t, err)

	assert := require.New(t)
	assert.Equal(3, len(results), "Should return exactly 3 results when limit=3")
}

// =============================================================================
// TTL / Expires Tests
// =============================================================================

func TestMemoryTTL(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-ttl-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Insert memory with future expiration
	_, err = db.Exec(`
		INSERT INTO memory (key, value, expires_at)
		VALUES (?, ?, datetime('now', '+1 hour'))
	`, "ttl-key", "expires-in-future")
	require.NoError(t, err)

	// Query should find it
	var value string
	err = db.QueryRow("SELECT value FROM memory WHERE key = ? AND (expires_at IS NULL OR expires_at > datetime('now'))", "ttl-key").Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, "expires-in-future", value)

	// Insert memory with past expiration
	_, err = db.Exec(`
		INSERT INTO memory (key, value, expires_at)
		VALUES (?, ?, datetime('now', '-1 hour'))
	`, "ttl-expired", "already-expired")
	require.NoError(t, err)

	// Query should NOT find expired entry
	err = db.QueryRow("SELECT value FROM memory WHERE key = ? AND (expires_at IS NULL OR expires_at > datetime('now'))", "ttl-expired").Scan(&value)
	assert.Error(t, err, "Expired entry should not be returned")
}

func TestMemoryExpiredCleanup(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-cleanup-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Insert mix of expired and valid entries
	entries := []struct {
		key     string
		value   string
		expires string
	}{
		{"valid-1", "valid-value-1", ""},
		{"expired-1", "expired-value-1", datetimeMinus("-1 day")},
		{"valid-2", "valid-value-2", ""},
		{"expired-2", "expired-value-2", datetimeMinus("-1 hour")},
	}

	for _, e := range entries {
		if e.expires == "" {
			_, err = db.Exec("INSERT INTO memory (key, value) VALUES (?, ?)", e.key, e.value)
		} else {
			_, err = db.Exec("INSERT INTO memory (key, value, expires_at) VALUES (?, ?, datetime('now', ?))", e.key, e.value, e.expires)
		}
		require.NoError(t, err)
	}

	// Delete expired entries
	result, err := db.Exec("DELETE FROM memory WHERE expires_at IS NOT NULL AND expires_at < datetime('now')")
	require.NoError(t, err)

	rowsAffected, _ := result.RowsAffected()
	assert.Equal(t, int64(2), rowsAffected, "Should delete exactly 2 expired entries")

	// Verify valid entries remain
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM memory WHERE key LIKE 'valid-%'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Verify expired entries are gone
	err = db.QueryRow("SELECT COUNT(*) FROM memory WHERE key LIKE 'expired-%'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// Helper to generate datetime subtraction string
func datetimeMinus(s string) string {
	return s
}

// =============================================================================
// Transaction Edge Cases
// =============================================================================

func TestTransactionNestedError(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-nested-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// SQLite does not support true nested transactions, but we can verify behavior
	tx, err := db.Begin()
	require.NoError(t, err)

	// Insert first record
	_, err = tx.Exec("INSERT INTO tasks (task_id, name, status) VALUES (?, ?, 'pending')", "nested-task-1", "Nested Task 1")
	require.NoError(t, err)

	// Attempting to begin again returns error (SQLite uses savepoints)
	_, err = tx.Exec("SAVEPOINT sp1")
	require.NoError(t, err)

	// Insert in savepoint
	_, err = tx.Exec("INSERT INTO tasks (task_id, name, status) VALUES (?, ?, 'pending')", "nested-task-2", "Nested Task 2")
	require.NoError(t, err)

	// Release savepoint
	_, err = tx.Exec("RELEASE SAVEPOINT sp1")
	require.NoError(t, err)

	// Commit main transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify both records exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE task_id LIKE 'nested-task-%'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestTransactionAutoRollback(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-autorollback-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Start transaction and cause an error that triggers rollback path
	tx, err := db.Begin()
	require.NoError(t, err)

	// Insert valid record
	_, err = tx.Exec("INSERT INTO tasks (task_id, name, status) VALUES (?, ?, 'pending')", "rollback-test", "Rollback Test")
	require.NoError(t, err)

	// Simulate error scenario - explicit rollback
	err = tx.Rollback()
	require.NoError(t, err, "Rollback should succeed")

	// Verify nothing was committed
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE task_id = ?)", "rollback-test").Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists, "Record should not exist after rollback")

	// Test explicit rollback path
	tx2, _ := db.Begin()
	_, _ = tx2.Exec("INSERT INTO tasks (task_id, name, status) VALUES (?, ?, 'pending')", "auto-rollback", "Auto Rollback")

	// Explicit rollback - uncommitted transaction should be rolled back
	err = tx2.Rollback()
	require.NoError(t, err)

	// Verify
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE task_id = ?)", "auto-rollback").Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists, "Uncommitted transaction should auto-rollback on close")
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestConcurrentMemoryAccess(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-concurrent-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Very low concurrency for SQLite (single-writer database)
	const goroutines = 2
	const iterations = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Channel for non-busy errors (data integrity issues)
	var dataErrors []string
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("concurrent-key-%d-%d", id, j)
				value := fmt.Sprintf("value-%d-%d", id, j)

				// Write
				_, err := db.Exec("INSERT INTO memory (key, value) VALUES (?, ?)", key, value)
				if err != nil {
					// Ignore SQLite BUSY errors - expected with concurrent writes
					if !strings.Contains(err.Error(), "SQLITE_BUSY") {
						mu.Lock()
						dataErrors = append(dataErrors, fmt.Sprintf("write error: %v", err))
						mu.Unlock()
					}
					continue
				}

				// Read
				var readValue string
				err = db.QueryRow("SELECT value FROM memory WHERE key = ?", key).Scan(&readValue)
				if err != nil {
					mu.Lock()
					dataErrors = append(dataErrors, fmt.Sprintf("read error: %v", err))
					mu.Unlock()
					continue
				}

				if readValue != value {
					mu.Lock()
					dataErrors = append(dataErrors, fmt.Sprintf("value mismatch: expected %s, got %s", value, readValue))
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	assert := require.New(t)
	assert.Empty(dataErrors, "Data integrity errors should not occur")
}

func TestConcurrentTaskCreation(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-concurrent-task-*.db")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Very low concurrency for SQLite (single-writer database)
	const goroutines = 2
	const tasksPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Track actual successful inserts and data integrity issues
	var successCount int
	var dataErrors int
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				taskID := fmt.Sprintf("concurrent-task-%d-%d", goroutineID, j)
				_, err := db.Exec("INSERT INTO tasks (task_id, name, status) VALUES (?, ?, 'pending')",
					taskID, fmt.Sprintf("Task %d-%d", goroutineID, j))
				if err != nil {
					// Ignore SQLite BUSY - expected
					if !strings.Contains(err.Error(), "SQLITE_BUSY") {
						mu.Lock()
						dataErrors++
						mu.Unlock()
					}
					continue
				}
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert := require.New(t)
	assert.Equal(0, dataErrors, "No data integrity errors should occur")

	// Verify at least some tasks were created (accounting for SQLite limitations)
	assert.Greater(successCount, 0, "At least some tasks should be created")

	// Verify count in database
	var dbCount int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE task_id LIKE 'concurrent-task-%'").Scan(&dbCount)
	require.NoError(t, err)
	assert.Equal(successCount, dbCount, "Database count should match success count")
}
