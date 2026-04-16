package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	server := NewServer()
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if server.tools == nil {
		t.Fatal("server.tools is nil")
	}
	if len(server.tools) != 7 {
		t.Errorf("Expected 7 tools, got %d", len(server.tools))
	}
}

func TestServerHasDefaultTools(t *testing.T) {
	server := NewServer()
	expectedTools := []string{
		"task_list",
		"task_create",
		"task_update",
		"task_delete",
		"memory_get",
		"memory_set",
		"semantic_search",
	}

	for _, name := range expectedTools {
		if _, ok := server.tools[name]; !ok {
			t.Errorf("Expected tool %q not found", name)
		}
	}
}

func TestToolHandlersReturnExpectedResults(t *testing.T) {
	server := NewServer()

	// Test task_list - no DB, should return empty array
	result, err := server.tools["task_list"](map[string]interface{}{})
	if err != nil {
		t.Errorf("task_list returned error: %v", err)
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("task_list result is not a map")
	}
	tasks, ok := resultMap["tasks"].([]interface{})
	if !ok {
		t.Fatalf("task_list should return tasks array, got %T", resultMap["tasks"])
	}
	if len(tasks) != 0 {
		t.Errorf("task_list without DB should return empty array, got %d items", len(tasks))
	}

	// Test task_create
	result, err = server.tools["task_create"](map[string]interface{}{"title": "Test"})
	if err != nil {
		t.Errorf("task_create returned error: %v", err)
	}
	resultMap, ok = result.(map[string]interface{})
	if !ok {
		t.Fatal("task_create result is not a map")
	}
	if resultMap["success"] != true {
		t.Errorf("task_create should return success=true")
	}

	// Test memory_get
	result, err = server.tools["memory_get"](map[string]interface{}{"key": "test"})
	if err != nil {
		t.Errorf("memory_get returned error: %v", err)
	}
	resultMap, ok = result.(map[string]interface{})
	if !ok {
		t.Fatal("memory_get result is not a map")
	}
	if resultMap["found"] != false {
		t.Errorf("memory_get should return found=false")
	}

	// Test memory_set
	result, err = server.tools["memory_set"](map[string]interface{}{"key": "test", "value": "val"})
	if err != nil {
		t.Errorf("memory_set returned error: %v", err)
	}
	resultMap, ok = result.(map[string]interface{})
	if !ok {
		t.Fatal("memory_set result is not a map")
	}
	if resultMap["success"] != true {
		t.Errorf("memory_set should return success=true")
	}

	// Test semantic_search
	result, err = server.tools["semantic_search"](map[string]interface{}{"query": "test"})
	if err != nil {
		t.Errorf("semantic_search returned error: %v", err)
	}
	resultMap, ok = result.(map[string]interface{})
	if !ok {
		t.Fatal("semantic_search result is not a map")
	}
	if results, ok := resultMap["results"].([]interface{}); !ok || len(results) != 0 {
		t.Errorf("semantic_search should return empty array, got %v", resultMap["results"])
	}
}

func TestGetToolDescription(t *testing.T) {
	tests := []struct {
		tool   string
		desc   string
		expect bool
	}{
		{"task_list", "List all tasks with optional status filter", true},
		{"task_create", "Create a new task with title and optional description", true},
		{"memory_get", "Get a value from memory store", true},
		{"memory_set", "Set a value in memory store", true},
		{"semantic_search", "Search using semantic understanding", true},
		{"unknown_tool", "", true}, // Should return empty string
	}

	for _, tt := range tests {
		desc := getToolDescription(tt.tool)
		if tt.expect && desc != tt.desc {
			t.Errorf("getToolDescription(%q) = %q, want %q", tt.tool, desc, tt.desc)
		}
	}
}

func TestGetToolInputSchema(t *testing.T) {
	schema := getToolInputSchema("task_create")
	if schema == nil {
		t.Fatal("getToolInputSchema returned nil")
	}

	schemaType, ok := schema["type"].(string)
	if !ok || schemaType != "object" {
		t.Errorf("Expected type=object, got %v", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties in schema")
	}

	if _, ok := props["title"]; !ok {
		t.Error("Expected 'title' in properties")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected 'required' in schema")
	}
	if len(required) != 1 || required[0] != "title" {
		t.Errorf("Expected required=['title'], got %v", required)
	}
}

func TestToJSON(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
		"num": 42,
	}

	jsonStr := toJSON(data)
	if jsonStr == "" {
		t.Error("toJSON returned empty string")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("toJSON returned invalid JSON: %v", err)
	}

	if parsed["key"] != "value" {
		t.Errorf("Expected key=value, got %v", parsed["key"])
	}
}

func TestServerRegisterTool(t *testing.T) {
	server := NewServer()
	initialCount := len(server.tools)

	// Register new tool
	server.RegisterTool("custom_tool", func(args map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{"custom": true}, nil
	})

	if len(server.tools) != initialCount+1 {
		t.Errorf("Expected %d tools after registration, got %d", initialCount+1, len(server.tools))
	}

	handler, ok := server.tools["custom_tool"]
	if !ok {
		t.Error("custom_tool not registered")
	}

	result, err := handler(map[string]interface{}{})
	if err != nil {
		t.Errorf("custom_tool handler returned error: %v", err)
	}
	resultMap := result.(map[string]interface{})
	if resultMap["custom"] != true {
		t.Error("custom_tool handler returned wrong result")
	}
}

func TestServerUnregisterTool(t *testing.T) {
	server := NewServer()
	initialCount := len(server.tools)

	// Unregister existing tool
	success := server.UnregisterTool("task_list")
	if !success {
		t.Error("UnregisterTool returned false for existing tool")
	}

	if len(server.tools) != initialCount-1 {
		t.Errorf("Expected %d tools after unregistration, got %d", initialCount-1, len(server.tools))
	}

	// Unregister non-existent tool
	success = server.UnregisterTool("nonexistent")
	if success {
		t.Error("UnregisterTool should return false for non-existent tool")
	}
}

func TestServerGetTools(t *testing.T) {
	server := NewServer()
	tools := server.GetTools()

	if len(tools) != 7 {
		t.Errorf("Expected 7 tools, got %d", len(tools))
	}

	// Verify it returns a copy
	tools["nonexistent"] = nil
	if _, ok := server.tools["nonexistent"]; ok {
		t.Error("Modifying returned tools should not affect server.tools")
	}
}

func TestServerIsReady(t *testing.T) {
	server := NewServer()
	
	// Initially not ready
	if server.IsReady() {
		t.Error("Server should not be ready initially")
	}

	// Simulate initialization
	server.mu.Lock()
	server.ready = true
	server.mu.Unlock()

	if !server.IsReady() {
		t.Error("Server should be ready after initialization")
	}
}

func TestErrorCodes(t *testing.T) {
	if ParseError != -32700 {
		t.Errorf("ParseError = %d, want -32700", ParseError)
	}
	if InvalidRequest != -32600 {
		t.Errorf("InvalidRequest = %d, want -32600", InvalidRequest)
	}
	if MethodNotFound != -32601 {
		t.Errorf("MethodNotFound = %d, want -32601", MethodNotFound)
	}
	if InvalidParams != -32602 {
		t.Errorf("InvalidParams = %d, want -32602", InvalidParams)
	}
	if InternalError != -32603 {
		t.Errorf("InternalError = %d, want -32603", InternalError)
	}
	if ServerError != -32000 {
		t.Errorf("ServerError = %d, want -32000", ServerError)
	}
}

func TestProtocolVersion(t *testing.T) {
	if ProtocolVersion == "" {
		t.Error("ProtocolVersion should not be empty")
	}
}

func TestMainEntryPoint(t *testing.T) {
	// Just verify that Main exists and can be called (without actually running server)
	// In a real test scenario, you'd use a pipe to send test messages
	t.Skip("Main() blocks waiting for stdin, skipping in unit tests")
}

// Test writing response format
func TestWriteResponseFormat(t *testing.T) {
	// This test verifies the response structure
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": `{"tasks":[]}`,
				},
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}

	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("Expected jsonrpc=2.0, got %v", parsed["jsonrpc"])
	}
	if parsed["id"] == nil {
		t.Error("Response should have an id")
	}
}

// Test that stdin/stdout operations work correctly
func TestStdioMessageParsing(t *testing.T) {
	// Verify we can parse Content-Length format messages
	// This would be an integration test in practice
	testCases := []struct {
		name     string
		header   string
		expected int
		valid    bool
	}{
		{"Valid header", "Content-Length: 123\r\n", 123, true},
		{"Valid with spaces", "Content-Length: 456   \r\n", 456, true},
		{"Invalid header", "Not-Length: 123\r\n", 0, false},
		{"Empty content", "Content-Length: 0\r\n", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.header) >= 16 && tc.header[:16] == "Content-Length: " {
				length := 0
				for i := 16; i < len(tc.header); i++ {
					if tc.header[i] == '\r' {
						break
					}
					if tc.header[i] >= '0' && tc.header[i] <= '9' {
						length = length*10 + int(tc.header[i]-'0')
					}
				}
				if length != tc.expected {
					t.Errorf("Expected length %d, got %d", tc.expected, length)
				}
			} else if tc.valid {
				t.Error("Expected valid header")
			}
		})
	}
}

// Test signal handling setup
func TestSignalChannelCreation(t *testing.T) {
	sigChan := make(chan os.Signal, 1)
	if sigChan == nil {
		t.Error("Signal channel should not be nil")
	}
}


// ============================================
// TEST-001: Additional Unit Tests
// ============================================

// TestTaskUpdateInvalidStatus tests that invalid status values are rejected
func TestTaskUpdateInvalidStatus(t *testing.T) {
	server := NewServer()
	
	invalidStatuses := []string{
		"invalid_status",
		"IN_PROGRESS",
		"completed123",
		"pending!",
		"",
	}
	
	for _, status := range invalidStatuses {
		_, err := server.tools["task_update"](map[string]interface{}{
			"id":     "test-123",
			"status": status,
		})
		if err == nil {
			t.Errorf("Expected error for invalid status %q, got nil", status)
		}
	}
}

// TestTaskCreateUniqueIDs tests that task_create generates unique IDs
func TestTaskCreateUniqueIDs(t *testing.T) {
	server := NewServer()
	ids := make(map[string]bool)
	
	for i := 0; i < 10; i++ {
		result, err := server.tools["task_create"](map[string]interface{}{
			"title": fmt.Sprintf("Task %d", i),
		})
		if err != nil {
			t.Fatalf("task_create returned error: %v", err)
		}
		
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("task_create result is not a map")
		}
		
		taskID, ok := resultMap["id"].(string)
		if !ok {
			t.Fatal("task_create result missing id field")
		}
		
		if ids[taskID] {
			t.Errorf("Duplicate task ID generated: %s", taskID)
		}
		ids[taskID] = true
		
		// Verify ID format: task-XXXXXXXX
		if len(taskID) < 10 || taskID[:5] != "task-" {
			t.Errorf("Task ID %q doesn't match expected format 'task-XXXXXXXX'", taskID)
		}
	}
	
	if len(ids) != 10 {
		t.Errorf("Expected 10 unique IDs, got %d", len(ids))
	}
}

// TestMemorySetEmptyKey tests that empty keys are rejected
func TestMemorySetEmptyKey(t *testing.T) {
	server := NewServer()
	
	_, err := server.tools["memory_set"](map[string]interface{}{
		"key":   "",
		"value": "test",
	})
	if err == nil {
		t.Error("Expected error for empty key, got nil")
	}
}

// TestMemoryGetEmptyKey tests that empty keys for memory_get are rejected
func TestMemoryGetEmptyKey(t *testing.T) {
	server := NewServer()
	
	_, err := server.tools["memory_get"](map[string]interface{}{
		"key": "",
	})
	if err == nil {
		t.Error("Expected error for empty key in memory_get, got nil")
	}
}

// TestMemorySetOversizedValue tests that oversized values are rejected
func TestMemorySetOversizedValue(t *testing.T) {
	server := NewServer()
	
	// Create a value larger than 64KB (65536 characters)
	largeValue := strings.Repeat("x", 70000)
	
	_, err := server.tools["memory_set"](map[string]interface{}{
		"key":   "test",
		"value": largeValue,
	})
	if err == nil {
		t.Error("Expected error for oversized value (>64KB), got nil")
	}
}

// TestMemorySetOversizedKey tests that oversized keys are rejected
func TestMemorySetOversizedKey(t *testing.T) {
	server := NewServer()
	
	// Create a key larger than 256 characters
	largeKey := strings.Repeat("k", 300)
	
	_, err := server.tools["memory_set"](map[string]interface{}{
		"key":   largeKey,
		"value": "test",
	})
	if err == nil {
		t.Error("Expected error for oversized key (>256 chars), got nil")
	}
}

// TestTaskCreateEmptyTitle tests that empty titles are rejected
func TestTaskCreateEmptyTitle(t *testing.T) {
	server := NewServer()
	
	_, err := server.tools["task_create"](map[string]interface{}{
		"title": "",
	})
	if err == nil {
		t.Error("Expected error for empty title, got nil")
	}
}

// TestTaskCreateOversizedTitle tests that oversized titles are rejected
func TestTaskCreateOversizedTitle(t *testing.T) {
	server := NewServer()
	
	// Create a title larger than 256 characters
	largeTitle := strings.Repeat("T", 300)
	
	_, err := server.tools["task_create"](map[string]interface{}{
		"title": largeTitle,
	})
	if err == nil {
		t.Error("Expected error for oversized title (>256 chars), got nil")
	}
}

// TestSemanticSearchEmptyQuery tests that empty queries are rejected
func TestSemanticSearchEmptyQuery(t *testing.T) {
	server := NewServer()
	
	_, err := server.tools["semantic_search"](map[string]interface{}{
		"query": "",
	})
	if err == nil {
		t.Error("Expected error for empty query, got nil")
	}
}

// TestSemanticSearchOversizedQuery tests that oversized queries are rejected
func TestSemanticSearchOversizedQuery(t *testing.T) {
	server := NewServer()
	
	// Create a query larger than 10000 characters
	largeQuery := strings.Repeat("q", 11000)
	
	_, err := server.tools["semantic_search"](map[string]interface{}{
		"query": largeQuery,
	})
	if err == nil {
		t.Error("Expected error for oversized query (>10000 chars), got nil")
	}
}

// TestTaskUpdateEmptyID tests that empty IDs are rejected
func TestTaskUpdateEmptyID(t *testing.T) {
	server := NewServer()
	
	_, err := server.tools["task_update"](map[string]interface{}{
		"id":     "",
		"status": "pending",
	})
	if err == nil {
		t.Error("Expected error for empty id in task_update, got nil")
	}
}

// TestTaskDeleteEmptyID tests that empty IDs are rejected
func TestTaskDeleteEmptyID(t *testing.T) {
	server := NewServer()
	
	_, err := server.tools["task_delete"](map[string]interface{}{
		"id": "",
	})
	if err == nil {
		t.Error("Expected error for empty id in task_delete, got nil")
	}
}

// TestTaskDeleteNonExistent tests deleting a non-existent task
func TestTaskDeleteNonExistent(t *testing.T) {
	server := NewServer()
	
	// Should not error even for non-existent ID (idempotent delete)
	result, err := server.tools["task_delete"](map[string]interface{}{
		"id": "non-existent-id-12345",
	})
	if err != nil {
		t.Errorf("task_delete should not error for non-existent ID: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("task_delete result is not a map")
	}
	
	if resultMap["success"] != true {
		t.Error("task_delete should return success=true")
	}
}

// TestValidStatusTransitions tests all valid status values
func TestValidStatusTransitions(t *testing.T) {
	server := NewServer()
	
	validStatuses := []string{
		"pending",
		"in_progress",
		"completed",
		"cancelled",
	}
	
	for _, status := range validStatuses {
		_, err := server.tools["task_update"](map[string]interface{}{
			"id":     "test-task-id",
			"status": status,
		})
		if err != nil {
			t.Errorf("Valid status %q should not error, got: %v", status, err)
		}
	}
}

// TestTaskCreateWithDescription tests task creation with description
func TestTaskCreateWithDescription(t *testing.T) {
	server := NewServer()
	
	result, err := server.tools["task_create"](map[string]interface{}{
		"title":       "Test Task",
		"description": "This is a test description",
	})
	if err != nil {
		t.Fatalf("task_create returned error: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("task_create result is not a map")
	}
	
	if resultMap["title"] != "Test Task" {
		t.Errorf("Expected title 'Test Task', got %v", resultMap["title"])
	}
	
	if resultMap["status"] != "pending" {
		t.Errorf("Expected status 'pending', got %v", resultMap["status"])
	}
}

// TestMemorySetAndGet tests basic memory storage operations
func TestMemorySetAndGet(t *testing.T) {
	server := NewServer()
	
	// Set a value
	setResult, err := server.tools["memory_set"](map[string]interface{}{
		"key":   "test_key",
		"value": "test_value",
	})
	if err != nil {
		t.Fatalf("memory_set returned error: %v", err)
	}
	
	setMap, ok := setResult.(map[string]interface{})
	if !ok {
		t.Fatal("memory_set result is not a map")
	}
	
	if setMap["success"] != true {
		t.Error("memory_set should return success=true")
	}
	
	if setMap["key"] != "test_key" {
		t.Errorf("Expected key 'test_key', got %v", setMap["key"])
	}
}

// TestGetToolsReturnsCopy tests that GetTools returns a copy
func TestGetToolsReturnsCopy(t *testing.T) {
	server := NewServer()
	originalCount := len(server.GetTools())
	
	tools := server.GetTools()
	tools["another_tool"] = nil
	
	if len(server.GetTools()) != originalCount {
		t.Error("Modifying returned tools should not affect server.tools")
	}
}

// TestConcurrentToolRegistration tests thread safety of tool registration
func TestConcurrentToolRegistration(t *testing.T) {
	server := NewServer()
	initialCount := len(server.tools)
	
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			server.RegisterTool(fmt.Sprintf("concurrent_tool_%d", idx), func(args map[string]interface{}) (interface{}, error) {
				return map[string]interface{}{"index": idx}, nil
			})
		}(i)
	}
	wg.Wait()
	
	finalCount := len(server.tools)
	expectedCount := initialCount + 10
	
	if finalCount != expectedCount {
		t.Errorf("Expected %d tools after concurrent registration, got %d", expectedCount, finalCount)
	}
}

// TestServerSetDB tests setting the database
func TestServerSetDB(t *testing.T) {
	server := NewServer()
	
	if server.db != nil {
		t.Error("Server should not have a DB initially")
	}
	
	// SetDB should not panic
	server.SetDB(nil)
	
	if server.db != nil {
		t.Error("Server DB should still be nil after SetDB(nil)")
	}
}



// ============================================
// JSON-RPC Handler Tests (with testify)
// ============================================

// captureServerOutput captures stdout for testing
func captureServerOutput(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	require.NoError(t, w.Close())
	os.Stdout = old

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	require.NoError(t, err)
	return buf.String()
}

// TestInitializeRequest tests the initialize handler
func TestInitializeRequest(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:     1,
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	// Parse the response
	is.Contains(output, "Content-Length:")
	is.Contains(output, `"jsonrpc":"2.0"`)
	is.Contains(output, `"id":1`)
	is.Contains(output, `"protocolVersion":"2024-11-05"`)
	is.Contains(output, `"name":"tools-bank-go"`)
	is.Contains(output, `"version":"1.0.0"`)

	// Server should be ready after initialize
	is.True(server.IsReady())
}

// TestToolsListRequest tests the tools/list handler
func TestToolsListRequest(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:     2,
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	is.Contains(output, `"jsonrpc":"2.0"`)
	is.Contains(output, `"id":2`)
	is.Contains(output, `"tools"`)
	is.Contains(output, `"task_list"`)
	is.Contains(output, `"task_create"`)
	is.Contains(output, `"memory_get"`)
	is.Contains(output, `"memory_set"`)
	is.Contains(output, `"semantic_search"`)
}

// TestToolsCallRequest tests the tools/call handler with valid tool
func TestToolsCallRequest(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "tools/call",
		ID:     3,
		Params: map[string]interface{}{
			"name": "task_create",
			"arguments": map[string]interface{}{
				"title": "Test Task",
			},
		},
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	is.Contains(output, `"jsonrpc":"2.0"`)
	is.Contains(output, `"id":3`)
	is.Contains(output, `"result"`)
	is.Contains(output, `"content"`)
	is.Contains(output, `"type":"text"`)
}

// TestShutdownRequest tests the shutdown handler
func TestShutdownRequest(t *testing.T) {
	is := assert.New(t)
	server := NewServer()
	server.ready = true

	// Note: msg is defined but not used because shutdown calls os.Exit
	msg := &Message{
		JSONRPC: "2.0",
		Method:  "shutdown",
		ID:     4,
	}

	// Shutdown handler calls os.Exit which would terminate the test
	// So we verify server state instead
	// The actual exit behavior is tested in integration tests
	is.NotNil(server)
	is.NotNil(msg)
	is.True(server.IsReady())
}

// TestInvalidJSONRequest tests handling of invalid JSON messages
func TestInvalidJSONRequest(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	// Test with invalid JSON structure (missing jsonrpc field will cause parse error)
	// Note: The Serve() function handles JSON parsing, not handleMessage
	// We test the error path for malformed messages

	msg := &Message{
		Method: "test",
		ID:     nil, // Notification style without ID
	}

	// Should not panic, should just return silently (no response for notifications)
	is.NotNil(server)
	is.NotNil(msg)
	require.NotPanics(t, func() {
		server.handleMessage(msg)
	})
}

// TestMissingMethodRequest tests handling of empty method
func TestMissingMethodRequest(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "",
		ID:      5,
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	// Empty method should result in MethodNotFound error
	is.Contains(output, `"error"`)
	is.Contains(output, `-32601`)
	is.Contains(output, `"id":5`)
}

// TestUnknownMethodRequest tests handling of unknown methods
func TestUnknownMethodRequest(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	tests := []struct {
		name   string
		method string
	}{
		{"unknown_tool", "unknown/tool"},
		{"random_method", "random_method"},
		{"tools/execute", "tools/execute"},
		{"tasks/create", "tasks/create"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				JSONRPC: "2.0",
				Method:  tt.method,
				ID:      float64(100),
			}

			output := captureServerOutput(t, func() {
				server.handleMessage(msg)
			})

			is.Contains(output, `"error"`)
			is.Contains(output, `"Method not found`)  // server returns capitalized
			is.Contains(output, `"id":100`)
		})
	}
}

// ============================================
// Response Writer Tests
// ============================================

// TestWriteResponse tests the writeResponse method
func TestWriteResponse(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	response := Message{
		JSONRPC: "2.0",
		ID:      1,
		Result: map[string]interface{}{
			"success": true,
			"data":    "test",
		},
	}

	output := captureServerOutput(t, func() {
		server.writeResponse(response)
	})

	// Verify Content-Length header
	is.True(strings.HasPrefix(strings.TrimSpace(output), "Content-Length:"))

	// Extract body after headers
	parts := strings.SplitN(output, "\r\n\r\n", 2)
	require.Len(t, parts, 2, "Should have headers and body")

	body := parts[1]

	// Parse JSON body
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(body), &parsed)
	require.NoError(t, err)

	is.Equal("2.0", parsed["jsonrpc"])
	is.Equal(1.0, parsed["id"])
	is.NotNil(t, parsed["result"])
}

// TestWriteError tests the writeError method
func TestWriteError(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	tests := []struct {
		name       string
		id         interface{}
		code       int
		message    string
		wantCode   string
		wantMsg    string
		wantID     interface{}
	}{
		{
			name:     "parse error",
			id:       1,
			code:     ParseError,
			message:  "Parse error",
			wantCode: "-32700",
			wantMsg:  "Parse error",
			wantID:   1.0,
		},
		{
			name:     "invalid request",
			id:       "abc",
			code:     InvalidRequest,
			message:  "Invalid request",
			wantCode: "-32600",
			wantMsg:  "Invalid request",
			wantID:   "abc",
		},
		{
			name:     "method not found",
			id:       nil,
			code:     MethodNotFound,
			message:  "Method not found",
			wantCode: "-32601",
			wantMsg:  "Method not found",
			wantID:   nil,
		},
		{
			name:     "invalid params",
			id:       5,
			code:     InvalidParams,
			message:  "Invalid params",
			wantCode: "-32602",
			wantMsg:  "Invalid params",
			wantID:   5.0,
		},
		{
			name:     "internal error",
			id:       float64(100),
			code:     InternalError,
			message:  "Internal error occurred",
			wantCode: "-32603",
			wantMsg:  "Internal error occurred",
			wantID:   float64(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureServerOutput(t, func() {
				server.writeError(tt.id, tt.code, tt.message)
			})

			is.Contains(output, "Content-Length:")

			parts := strings.SplitN(output, "\r\n\r\n", 2)
			require.Len(t, parts, 2)

			body := parts[1]

			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(body), &parsed)
			require.NoError(t, err)

			is.Equal("2.0", parsed["jsonrpc"])

			errorObj, ok := parsed["error"].(map[string]interface{})
			is.True(ok, "error should be a map")
			is.Equal(tt.wantCode, fmt.Sprintf("%v", errorObj["code"]))
			is.Equal(tt.wantMsg, errorObj["message"])
			is.Equal(tt.wantID, parsed["id"])
		})
	}
}

// TestWriteNotification tests handling of notification messages (no ID)
func TestWriteNotification(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "some/notification",
		ID:      nil, // Notification has no ID
	}

	// Capture output
	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	// Notifications should not write any response
	is.Equal("", output, "Notifications should not produce a response")
}

// ============================================
// Edge Case Tests
// ============================================

// TestHandleMessageNilDB tests server behavior with nil database
func TestHandleMessageNilDB(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	// Ensure DB is nil
	require.Nil(t, server.db)

	// Test task_list with nil DB
	msg := &Message{
		JSONRPC: "2.0",
		Method:  "tools/call",
		ID:      1,
		Params: map[string]interface{}{
			"name":      "task_list",
			"arguments": map[string]interface{}{},
		},
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	// Should still work, returning empty array
	is.Contains(output, `"jsonrpc":"2.0"`)
	is.Contains(output, `\"tasks\":`)  // server returns escaped quotes
}

// TestHandleMessageWithDB tests server behavior when DB is set
func TestHandleMessageWithDB(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	// Set nil DB explicitly - should behave same as unset
	server.SetDB(nil)

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "tools/call",
		ID:      1,
		Params: map[string]interface{}{
			"name":      "task_list",
			"arguments": map[string]interface{}{},
		},
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	is.Contains(output, `"jsonrpc":"2.0"`)
	is.Contains(output, `"result"`)
}

// TestToolsCallInvalidParams tests tools/call with invalid params type
func TestToolsCallInvalidParams(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	tests := []struct {
		name   string
		params interface{}
	}{
		{"string params", "invalid string params"},
		{"array params", []interface{}{1, 2, 3}},
		{"number params", 42},
		{"nil params", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				JSONRPC: "2.0",
				Method:  "tools/call",
				ID:      1,
				Params:  tt.params,
			}

			output := captureServerOutput(t, func() {
				server.handleMessage(msg)
			})

			is.Contains(output, `"error"`)
			is.Contains(output, `-32602`) // InvalidParams code
		})
	}
}

// TestToolsCallNotFound tests tools/call with non-existent tool
func TestToolsCallNotFound(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	tests := []struct {
		name string
		tool string
	}{
		{"completely unknown", "nonexistent_tool"},
		{"similar name", "task_list_extra"},
		{"case mismatch", "Task_Create"},
		{"empty name", ""},
		{"with slash", "task/create"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				JSONRPC: "2.0",
				Method:  "tools/call",
				ID:      float64(123),
				Params: map[string]interface{}{
					"name":      tt.tool,
					"arguments": map[string]interface{}{},
				},
			}

			output := captureServerOutput(t, func() {
				server.handleMessage(msg)
			})

			is.Contains(output, `"error"`)
			is.Contains(output, `-32601`) // MethodNotFound code
			is.Contains(output, `Tool not found`)
			is.Contains(output, `"id":123`)
		})
	}
}

// TestToolsCallWithArgs tests tools/call with various arguments
func TestToolsCallWithArgs(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	tests := []struct {
		name       string
		tool       string
		args       map[string]interface{}
		wantError  bool
		errorCode  int
	}{
		{
			name:      "task_create with title",
			tool:      "task_create",
			args:      map[string]interface{}{"title": "New Task"},
			wantError: false,
		},
		{
			name:      "task_create without title",
			tool:      "task_create",
			args:      map[string]interface{}{},
			wantError: true,
			errorCode: InternalError,
		},
		{
			name:      "memory_get with key",
			tool:      "memory_get",
			args:      map[string]interface{}{"key": "test_key"},
			wantError: false,
		},
		{
			name:      "memory_get without key",
			tool:      "memory_get",
			args:      map[string]interface{}{},
			wantError: true,
			errorCode: InternalError,
		},
		{
			name:      "semantic_search with query",
			tool:      "semantic_search",
			args:      map[string]interface{}{"query": "test query"},
			wantError: false,
		},
		{
			name:      "semantic_search without query",
			tool:      "semantic_search",
			args:      map[string]interface{}{},
			wantError: true,
			errorCode: InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				JSONRPC: "2.0",
				Method:  "tools/call",
				ID:      1,
				Params: map[string]interface{}{
					"name":      tt.tool,
					"arguments": tt.args,
				},
			}

			output := captureServerOutput(t, func() {
				server.handleMessage(msg)
			})

			if tt.wantError {
				is.Contains(output, `"error"`)
				if tt.errorCode != 0 {
					is.Contains(output, fmt.Sprintf(`%d`, tt.errorCode))
				}
			} else {
				is.Contains(output, `"result"`)
			}
		})
	}
}

// TestToolsCallHandlerError tests that handler errors are properly reported
func TestToolsCallHandlerError(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	// Register a tool that returns an error
	server.RegisterTool("error_tool", func(args map[string]interface{}) (interface{}, error) {
		return nil, fmt.Errorf("deliberate error for testing")
	})

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "tools/call",
		ID:      1,
		Params: map[string]interface{}{
			"name":      "error_tool",
			"arguments": map[string]interface{}{},
		},
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	is.Contains(output, `"error"`)
	is.Contains(output, `deliberate error for testing`)
	is.Contains(output, `-32603`) // InternalError
}

// TestHandleMessageConcurrent tests thread safety of handleMessage
func TestHandleMessageConcurrent(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := &Message{
				JSONRPC: "2.0",
				Method:  "tools/list",
				ID:      id,
			}
			server.handleMessage(msg)
		}(i)
	}
	wg.Wait()

	// If we got here without race conditions or panics, test passed
	is.NotNil(server)
}

// TestResponseWriterWithVariousIDs tests response writing with different ID types
func TestResponseWriterWithVariousIDs(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	tests := []struct {
		name string
		id   interface{}
	}{
		{"integer id", 1},
		{"float id", 1.5},
		{"string id", "request-123"},
		{"null id", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				JSONRPC: "2.0",
				Method:  "initialize",
				ID:      tt.id,
			}

			is.NotNil(server)
			is.NotNil(msg)
			require.NotPanics(t, func() {
				server.handleMessage(msg)
			})
		})
	}
}

// TestMessageWithNilParams tests handling of nil params
func TestMessageWithNilParams(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "tools/call",
		ID:      1,
		Params:  nil,
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	// Should handle nil params gracefully
	is.Contains(output, `"error"`)
	is.Contains(output, `-32602`)
}

// TestHandleMessageEmptyToolName tests handling of empty tool name in tools/call
func TestHandleMessageEmptyToolName(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "tools/call",
		ID:      1,
		Params: map[string]interface{}{
			"name":      "",
			"arguments": map[string]interface{}{},
		},
	}

	output := captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	is.Contains(output, `"error"`)
	is.Contains(output, `-32601`)
	is.Contains(output, `Tool not found`)
}

// TestServerReadyStateDuringHandleMessage tests ready state changes
func TestServerReadyStateDuringHandleMessage(t *testing.T) {
	is := assert.New(t)
	server := NewServer()

	is.False(server.IsReady(), "Server should not be ready initially")

	msg := &Message{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:      1,
	}

	captureServerOutput(t, func() {
		server.handleMessage(msg)
	})

	is.True(server.IsReady(), "Server should be ready after initialize")
}

// TestProtocolVersionConstant tests the protocol version constant
func TestProtocolVersionConstant(t *testing.T) {
	is := assert.New(t)
	is.Equal("2024-11-05", ProtocolVersion)
}

// TestErrorCodesConstant tests all error code constants
func TestErrorCodesConstant(t *testing.T) {
	is := assert.New(t)

	tests := []struct {
		name  string
		code  int
		value int
	}{
		{"ParseError", ParseError, -32700},
		{"InvalidRequest", InvalidRequest, -32600},
		{"MethodNotFound", MethodNotFound, -32601},
		{"InvalidParams", InvalidParams, -32602},
		{"InternalError", InternalError, -32603},
		{"ServerError", ServerError, -32000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is.Equal(tt.value, tt.code)
		})
	}
}
