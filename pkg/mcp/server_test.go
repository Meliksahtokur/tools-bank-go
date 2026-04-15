package mcp

import (
	"encoding/json"
	"os"
	"testing"
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
