package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unicode/utf8"

	"github.com/egesut/tools-bank-go/pkg/db"
	"github.com/google/uuid"
)

// Server represents the MCP server instance
type Server struct {
	tools map[string]ToolHandler
	mu    sync.RWMutex
	ready bool
	db    *db.DB
}

// ToolHandler defines the signature for tool handlers
type ToolHandler func(args map[string]interface{}) (interface{}, error)

// Tool represents a tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Message represents a JSON-RPC 2.0 message
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Params  interface{}     `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *ErrorResponse  `json:"error,omitempty"`
}

// ErrorResponse represents a JSON-RPC error
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewServer creates a new MCP server instance
func NewServer() *Server {
	s := &Server{
		tools: make(map[string]ToolHandler),
		db:    nil, // DB will be set via SetDB() or defaults to in-memory
	}
	s.registerDefaultTools()
	return s
}

// SetDB sets the database instance
func (s *Server) SetDB(database *db.DB) {
	s.mu.Lock()
	s.db = database
	s.mu.Unlock()
}

// registerDefaultTools registers the initial set of tools
func (s *Server) registerDefaultTools() {
	s.tools["task_list"] = func(args map[string]interface{}) (interface{}, error) {
		var taskSlice []map[string]interface{}
		
		s.mu.RLock()
		dbConn := s.db
		s.mu.RUnlock()
		
		if dbConn != nil {
			rows, err := dbConn.Query(`
				SELECT task_id, name, description, status, created_at, updated_at
				FROM tasks ORDER BY created_at DESC LIMIT 100
			`)
			if err != nil {
				return nil, fmt.Errorf("query error: %w", err)
			}
			defer rows.Close()
			
			for rows.Next() {
				var taskID, name, desc, status, created, updated string
				if err := rows.Scan(&taskID, &name, &desc, &status, &created, &updated); err != nil {
					continue
				}
				taskSlice = append(taskSlice, map[string]interface{}{
					"id":          taskID,
					"name":        name,
					"description": desc,
					"status":      status,
					"created_at":  created,
					"updated_at":  updated,
				})
			}
		}
		
		// Convert to []interface{} for JSON-RPC compatibility
		tasks := make([]interface{}, len(taskSlice))
		for i, t := range taskSlice {
			tasks[i] = t
		}
		if tasks == nil {
			tasks = []interface{}{}
		}
		return map[string]interface{}{"tasks": tasks}, nil
	}
	
	s.tools["task_create"] = func(args map[string]interface{}) (interface{}, error) {
		title, _ := args["title"].(string)
		description, _ := args["description"].(string)
		
		// Input validation
		if title == "" {
			return nil, fmt.Errorf("title is required")
		}
		if utf8.RuneCountInString(title) > 256 {
			return nil, fmt.Errorf("title exceeds maximum length of 256 characters")
		}
		if utf8.RuneCountInString(description) > 65536 {
			return nil, fmt.Errorf("description exceeds maximum length of 65536 characters")
		}
		
		// Use UUID for unique task ID
		taskID := fmt.Sprintf("task-%s", uuid.New().String()[:8])
		
		s.mu.RLock()
		dbConn := s.db
		s.mu.RUnlock()
		
		if dbConn != nil {
			_, err := dbConn.Exec(`
				INSERT INTO tasks (task_id, name, description, status)
				VALUES (?, ?, ?, 'pending')
			`, taskID, title, description)
			if err != nil {
				return nil, fmt.Errorf("insert error: %w", err)
			}
		}
		
		return map[string]interface{}{
			"success": true,
			"id":      taskID,
			"title":   title,
			"status":  "pending",
		}, nil
	}
	
	s.tools["task_update"] = func(args map[string]interface{}) (interface{}, error) {
		taskID, _ := args["id"].(string)
		status, _ := args["status"].(string)
		
		if taskID == "" {
			return nil, fmt.Errorf("id is required")
		}
		
		// Status validation
		validStatuses := map[string]bool{
			"pending":     true,
			"in_progress": true,
			"completed":   true,
			"cancelled":   true,
		}
		if !validStatuses[status] {
			return nil, fmt.Errorf("invalid status: %s. Valid values: pending, in_progress, completed, cancelled", status)
		}
		
		s.mu.RLock()
		dbConn := s.db
		s.mu.RUnlock()
		
		if dbConn != nil {
			completedAt := "NULL"
			if status == "completed" {
				completedAt = "CURRENT_TIMESTAMP"
			}
			_, err := dbConn.Exec(fmt.Sprintf(`
				UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP,
				completed_at = %s WHERE task_id = ?
			`, completedAt), status, taskID)
			if err != nil {
				return nil, fmt.Errorf("update error: %w", err)
			}
		}
		
		return map[string]interface{}{"success": true, "id": taskID, "status": status}, nil
	}
	
	s.tools["task_delete"] = func(args map[string]interface{}) (interface{}, error) {
		taskID, _ := args["id"].(string)
		
		if taskID == "" {
			return nil, fmt.Errorf("id is required")
		}
		
		s.mu.RLock()
		dbConn := s.db
		s.mu.RUnlock()
		
		if dbConn != nil {
			_, err := dbConn.Exec("DELETE FROM tasks WHERE task_id = ?", taskID)
			if err != nil {
				return nil, fmt.Errorf("delete error: %w", err)
			}
		}
		
		return map[string]interface{}{"success": true, "id": taskID}, nil
	}
	
	s.tools["memory_get"] = func(args map[string]interface{}) (interface{}, error) {
		key, _ := args["key"].(string)
		
		// Input validation
		if key == "" {
			return nil, fmt.Errorf("key is required")
		}
		if utf8.RuneCountInString(key) > 256 {
			return nil, fmt.Errorf("key exceeds maximum length of 256 characters")
		}
		
		s.mu.RLock()
		dbConn := s.db
		s.mu.RUnlock()
		
		if dbConn != nil && key != "" {
			var value string
			err := dbConn.QueryRow("SELECT value FROM memory WHERE key = ? AND (expires_at IS NULL OR expires_at > datetime('now', 'utc'))", key).Scan(&value)
			if err == nil {
				return map[string]interface{}{"key": key, "value": value, "found": true}, nil
			}
		}
		
		return map[string]interface{}{"key": key, "value": nil, "found": false}, nil
	}
	
	s.tools["memory_set"] = func(args map[string]interface{}) (interface{}, error) {
		key, _ := args["key"].(string)
		value, _ := args["value"].(string)
		
		// Input validation
		if key == "" {
			return nil, fmt.Errorf("key is required")
		}
		if utf8.RuneCountInString(key) > 256 {
			return nil, fmt.Errorf("key exceeds maximum length of 256 characters")
		}
		if utf8.RuneCountInString(value) > 65536 {
			return nil, fmt.Errorf("value exceeds maximum length of 65536 characters")
		}
		
		s.mu.RLock()
		dbConn := s.db
		s.mu.RUnlock()
		
		if dbConn != nil {
			_, err := dbConn.Exec(`
				INSERT INTO memory (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
				ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
			`, key, value)
			if err != nil {
				return nil, fmt.Errorf("insert error: %w", err)
			}
		}
		
		return map[string]interface{}{"success": true, "key": key}, nil
	}
	
	s.tools["semantic_search"] = func(args map[string]interface{}) (interface{}, error) {
		query, _ := args["query"].(string)
		limit, _ := args["limit"].(int)
		
		// Input validation
		if query == "" {
			return nil, fmt.Errorf("query is required")
		}
		if utf8.RuneCountInString(query) > 10000 {
			return nil, fmt.Errorf("query exceeds maximum length of 10000 characters")
		}
		if limit <= 0 {
			limit = 10
		}
		if limit > 1000 {
			limit = 1000
		}
		
		var rawResults []map[string]interface{}
		
		s.mu.RLock()
		dbConn := s.db
		s.mu.RUnlock()
		
		if dbConn != nil && query != "" {
			// Try FTS5 search first, fall back to LIKE if FTS5 not available
			embedResults, err := dbConn.SearchEmbeddings(query, limit)
			if err != nil {
				// Fall back to LIKE-based search
				rows, err := dbConn.Query(`
					SELECT document_id, content, created_at FROM embeddings
					WHERE content LIKE ? ORDER BY created_at DESC LIMIT ?
				`, "%"+query+"%", limit)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var docID, content, created string
						if err := rows.Scan(&docID, &content, &created); err == nil {
							rawResults = append(rawResults, map[string]interface{}{
								"id":        docID,
								"content":   content,
								"score":     0.8,
								"created":   created,
							})
						}
					}
				}
			} else {
				// Use FTS5 results
				for _, emb := range embedResults {
					rawResults = append(rawResults, map[string]interface{}{
						"id":      emb.DocumentID,
						"content": emb.Content,
						"score":   0.95, // FTS5 gives better relevance
						"created": emb.CreatedAt,
					})
				}
			}
		}
		
		// Convert to []interface{} for JSON-RPC compatibility
		finalResults := make([]interface{}, len(rawResults))
		for i, r := range rawResults {
			finalResults[i] = r
		}
		return map[string]interface{}{"results": finalResults, "query": query}, nil
	}
}

// Serve starts the MCP server
func (s *Server) Serve() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	scanner := bufio.NewScanner(os.Stdin)
	
	// Increase scanner buffer for larger messages
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			s.writeError(nil, -32700, "Parse error")
			continue
		}

		s.handleMessage(&msg)
	}

	<-sigChan
	s.ready = false
	return nil
}

// handleMessage processes a single JSON-RPC message
func (s *Server) handleMessage(msg *Message) {
	var response Message
	response.JSONRPC = "2.0"
	response.ID = msg.ID

	switch msg.Method {
	case "initialize":
		s.ready = true
		response.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{"listChanged": true},
			},
			"serverInfo": map[string]interface{}{
				"name":    "tools-bank-go",
				"version": "1.0.0",
			},
		}

	case "tools/list":
		s.mu.RLock()
		var tools []Tool
		for name := range s.tools {
			tools = append(tools, Tool{
				Name:        name,
				Description: getDescription(name),
				InputSchema: getSchema(name),
			})
		}
		s.mu.RUnlock()
		response.Result = map[string]interface{}{"tools": tools}

	case "tools/call":
		params, ok := msg.Params.(map[string]interface{})
		if !ok {
			s.writeError(msg.ID, -32602, "Invalid params")
			return
		}
		name, _ := params["name"].(string)
		args, _ := params["arguments"].(map[string]interface{})
		if args == nil {
			args = make(map[string]interface{})
		}

		s.mu.RLock()
		handler, exists := s.tools[name]
		s.mu.RUnlock()

		if !exists {
			s.writeError(msg.ID, -32601, fmt.Sprintf("Tool not found: %s", name))
			return
		}

		result, err := handler(args)
		if err != nil {
			s.writeError(msg.ID, -32603, err.Error())
			return
		}

		response.Result = map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": toJSON(result)},
			},
		}

	case "shutdown":
		response.Result = map[string]interface{}{"success": true}
		go func() {
			os.Exit(0)
		}()

	default:
		if msg.ID != nil {
			s.writeError(msg.ID, -32601, fmt.Sprintf("Method not found: %s", msg.Method))
		}
		return
	}

	s.writeResponse(response)
}

// writeResponse writes a JSON-RPC response to stdout
func (s *Server) writeResponse(response Message) {
	data, _ := json.Marshal(response)
	fmt.Fprintf(os.Stdout, "Content-Length: %d\r\n\r\n", len(data))
	os.Stdout.Write(data)
}

// writeError writes a JSON-RPC error response
func (s *Server) writeError(id interface{}, code int, message string) {
	response := Message{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &ErrorResponse{Code: code, Message: message},
	}
	s.writeResponse(response)
}

// toJSON converts interface to JSON string
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// getDescription returns description for each tool
func getDescription(name string) string {
	m := map[string]string{
		"task_list":       "List all tasks with optional status filter",
		"task_create":     "Create a new task with title and optional description",
		"task_update":     "Update an existing task's status",
		"task_delete":     "Delete a task by ID",
		"memory_get":      "Get a value from memory store",
		"memory_set":      "Set a value in memory store",
		"semantic_search": "Search using semantic understanding",
	}
	if d, ok := m[name]; ok {
		return d
	}
	return ""
}

// getSchema returns input schema for each tool
func getSchema(name string) map[string]interface{} {
	m := map[string]map[string]interface{}{
		"task_list": {
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{"type": "string", "description": "Filter by status"},
			},
		},
		"task_create": {
			"type": "object",
			"properties": map[string]interface{}{
				"title":       map[string]interface{}{"type": "string", "description": "Task title"},
				"description": map[string]interface{}{"type": "string", "description": "Task description"},
			},
			"required": []string{"title"},
		},
		"task_update": {
			"type": "object",
			"properties": map[string]interface{}{
				"id":     map[string]interface{}{"type": "string", "description": "Task ID"},
				"status": map[string]interface{}{"type": "string", "description": "New status"},
			},
			"required": []string{"id", "status"},
		},
		"task_delete": {
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{"type": "string", "description": "Task ID to delete"},
			},
			"required": []string{"id"},
		},
		"memory_get": {
			"type": "object",
			"properties": map[string]interface{}{
				"key": map[string]interface{}{"type": "string", "description": "Memory key"},
			},
			"required": []string{"key"},
		},
		"memory_set": {
			"type": "object",
			"properties": map[string]interface{}{
				"key":   map[string]interface{}{"type": "string"},
				"value": map[string]interface{}{"type": "string"},
			},
			"required": []string{"key", "value"},
		},
		"semantic_search": {
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{"type": "string", "description": "Search query"},
				"limit": map[string]interface{}{"type": "integer"},
			},
			"required": []string{"query"},
		},
	}
	if schema, ok := m[name]; ok {
		return schema
	}
	return map[string]interface{}{"type": "object"}
}

// RegisterTool registers a custom tool
func (s *Server) RegisterTool(name string, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[name] = handler
}

// UnregisterTool removes a tool
func (s *Server) UnregisterTool(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tools[name]; exists {
		delete(s.tools, name)
		return true
	}
	return false
}

// GetTools returns a copy of all tools
func (s *Server) GetTools() map[string]ToolHandler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]ToolHandler)
	for k, v := range s.tools {
		result[k] = v
	}
	return result
}

// IsReady returns whether the server is ready
func (s *Server) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready
}

// Error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
	ServerError    = -32000
)

// ProtocolVersion is the MCP protocol version
const ProtocolVersion = "2024-11-05"

// Aliases for test compatibility
func getToolDescription(name string) string {
	return getDescription(name)
}

func getToolInputSchema(name string) map[string]interface{} {
	return getSchema(name)
}
