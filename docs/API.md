# API Reference

Complete reference documentation for all MCP tools available in tools-bank-go.

## Table of Contents

- [Protocol](#protocol)
- [Task Management](#task-management)
  - [task_create](#task_create)
  - [task_list](#task_list)
  - [task_update](#task_update)
  - [task_delete](#task_delete)
- [Memory Store](#memory-store)
  - [memory_get](#memory_get)
  - [memory_set](#memory_set)
- [Semantic Search](#semantic-search)
  - [semantic_search](#semantic_search)
- [Error Handling](#error-handling)
- [Troubleshooting](#troubleshooting)

---

## Protocol

### Transport

- **Method**: stdin/stdout (stdio)
- **Format**: JSON-RPC 2.0
- **Protocol Version**: 2024-11-05
- **Max Message Size**: 1MB

### Session Flow

```
Client                          Server
  │                                │
  │──── initialize ───────────────▶│
  │◀─── capabilities ─────────────│
  │                                │
  │──── tools/list ───────────────▶│
  │◀─── tool list ────────────────│
  │                                │
  │──── tools/call ───────────────▶│
  │◀─── tool result ───────────────│
  │         ...                    │
  │──── shutdown ─────────────────▶│
  │◀─── shutdown ack ──────────────│
```

### Request Format

```json
{
  "jsonrpc": "2.0",
  "method": "method/name",
  "params": { ... },
  "id": 1
}
```

### Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": { ... }
}
```

### Notification Format (No Response Expected)

```json
{
  "jsonrpc": "2.0",
  "method": "method/name",
  "params": { ... }
}
```

---

## Task Management

### task_create

Create a new task with a title and optional description.

#### Input Parameters

```json
{
  "name": "task_create",
  "arguments": {
    "title": "string (required)",
    "description": "string (optional)"
  }
}
```

| Parameter | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `title` | string | ✅ | Task title | Max 256 characters (UTF-8) |
| `description` | string | ❌ | Task description | Max 65536 characters |

#### Output Format

```json
{
  "success": true,
  "id": "task-a1b2c3d4",
  "title": "Task Title",
  "status": "pending"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Operation result |
| `id` | string | Generated task ID (`task-XXXXXXXX`) |
| `title` | string | Task title (echoed) |
| `status` | string | Initial status (`pending`) |

#### Example

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "task_create",
    "arguments": {
      "title": "Implement user authentication",
      "description": "Add OAuth2 support with Google and GitHub providers"
    }
  },
  "id": 1
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"success\":true,\"id\":\"task-a1b2c3d4\",\"title\":\"Implement user authentication\",\"status\":\"pending\"}"
    }]
  }
}
```

#### Validation Errors

| Error | Cause |
|-------|-------|
| `title is required` | Missing required parameter |
| `title exceeds maximum length of 256 characters` | Title too long |

---

### task_list

Retrieve a list of all tasks with optional status filtering.

#### Input Parameters

```json
{
  "name": "task_list",
  "arguments": {
    "status": "string (optional)"
  }
}
```

| Parameter | Type | Required | Description | Valid Values |
|-----------|------|----------|-------------|--------------|
| `status` | string | ❌ | Filter by status | `pending`, `in_progress`, `completed`, `cancelled` |

#### Output Format

```json
{
  "tasks": [
    {
      "id": "task-a1b2c3d4",
      "name": "Task Name",
      "description": "Task description",
      "status": "pending",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `tasks` | array | List of task objects (max 100, newest first) |
| `id` | string | Task ID |
| `name` | string | Task title |
| `description` | string | Task description (may be empty) |
| `status` | string | Current status |
| `created_at` | string | ISO 8601 timestamp |
| `updated_at` | string | ISO 8601 timestamp |

#### Example

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "task_list",
    "arguments": {
      "status": "pending"
    }
  },
  "id": 2
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"tasks\":[{\"id\":\"task-a1b2c3d4\",\"name\":\"Task 1\",\"description\":\"Description\",\"status\":\"pending\",\"created_at\":\"2024-01-15T10:30:00Z\",\"updated_at\":\"2024-01-15T10:30:00Z\"}]}"
    }]
  }
}
```

#### Notes

- Returns up to 100 tasks ordered by creation date (newest first)
- When no status filter is provided, returns all tasks

---

### task_update

Update an existing task's status.

#### Input Parameters

```json
{
  "name": "task_update",
  "arguments": {
    "id": "string (required)",
    "status": "string (required)"
  }
}
```

| Parameter | Type | Required | Description | Valid Values |
|-----------|------|----------|-------------|--------------|
| `id` | string | ✅ | Task ID | Existing task ID |
| `status` | string | ✅ | New status | `pending`, `in_progress`, `completed`, `cancelled` |

#### Output Format

```json
{
  "success": true,
  "id": "task-a1b2c3d4",
  "status": "completed"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Operation result |
| `id` | string | Task ID (echoed) |
| `status` | string | New status |

#### Status Lifecycle

```
     ┌──────────┐
     │ pending  │
     └────┬─────┘
          │
          ▼
    ┌───────────┐      ┌────────────┐
    │in_progress│─────▶│  completed │
    └───────────┘      └────────────┘
          │                   ▲
          └───────┐   ┌───────┘
                  ▼   │
              ┌───────────┐
              │ cancelled │
              └───────────┘
```

#### Example

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "task_update",
    "arguments": {
      "id": "task-a1b2c3d4",
      "status": "in_progress"
    }
  },
  "id": 3
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"success\":true,\"id\":\"task-a1b2c3d4\",\"status\":\"in_progress\"}"
    }]
  }
}
```

#### Validation Errors

| Error | Cause |
|-------|-------|
| `id is required` | Missing task ID |
| `invalid status: X` | Invalid status value |

---

### task_delete

Delete a task by ID.

#### Input Parameters

```json
{
  "name": "task_delete",
  "arguments": {
    "id": "string (required)"
  }
}
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | string | ✅ | Task ID to delete |

#### Output Format

```json
{
  "success": true,
  "id": "task-a1b2c3d4"
}
```

#### Notes

- **Idempotent**: Deleting a non-existent task returns success
- **Permanent**: This action cannot be undone

#### Example

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "task_delete",
    "arguments": {
      "id": "task-a1b2c3d4"
    }
  },
  "id": 4
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"success\":true,\"id\":\"task-a1b2c3d4\"}"
    }]
  }
}
```

#### Validation Errors

| Error | Cause |
|-------|-------|
| `id is required` | Missing task ID |

---

## Memory Store

### memory_get

Retrieve a value from the memory store by key.

#### Input Parameters

```json
{
  "name": "memory_get",
  "arguments": {
    "key": "string (required)"
  }
}
```

| Parameter | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `key` | string | ✅ | Memory key | Max 256 characters (UTF-8) |

#### Output Format

```json
{
  "key": "config:theme",
  "value": "dark",
  "found": true
}
```

| Field | Type | Description |
|-------|------|-------------|
| `key` | string | Requested key |
| `value` | string or null | Stored value |
| `found` | boolean | Whether key exists and is valid |

#### Example

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "memory_get",
    "arguments": {
      "key": "config:theme"
    }
  },
  "id": 5
}
```

**Response (found):**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"key\":\"config:theme\",\"value\":\"dark\",\"found\":true}"
    }]
  }
}
```

**Response (not found):**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"key\":\"config:theme\",\"value\":null,\"found\":false}"
    }]
  }
}
```

#### Key Naming Conventions

Recommended patterns:
- `user:{userId}:{field}` - User data
- `config:{section}:{key}` - Configuration
- `cache:{resource}:{id}` - Cached values
- `session:{sessionId}:{key}` - Session data

#### Validation Errors

| Error | Cause |
|-------|-------|
| `key is required` | Missing key parameter |
| `key exceeds maximum length of 256 characters` | Key too long |

---

### memory_set

Store a value in the memory store.

#### Input Parameters

```json
{
  "name": "memory_set",
  "arguments": {
    "key": "string (required)",
    "value": "string (required)"
  }
}
```

| Parameter | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `key` | string | ✅ | Memory key | Max 256 characters (UTF-8) |
| `value` | string | ✅ | Value to store | Max 65536 characters |

#### Output Format

```json
{
  "success": true,
  "key": "config:theme"
}
```

#### Behavior

- **Upsert**: Creates new key or updates existing
- **No Expiration**: Values persist until deleted
- **Overwrite**: Previous value is replaced

#### Example

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "memory_set",
    "arguments": {
      "key": "config:theme",
      "value": "dark"
    }
  },
  "id": 6
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"success\":true,\"key\":\"config:theme\"}"
    }]
  }
}
```

#### Validation Errors

| Error | Cause |
|-------|-------|
| `key is required` | Missing key parameter |
| `key exceeds maximum length of 256 characters` | Key too long |
| `value exceeds maximum length of 65536 characters` | Value too long |

---

## Semantic Search

### semantic_search

Search using semantic understanding with full-text search capabilities.

#### Input Parameters

```json
{
  "name": "semantic_search",
  "arguments": {
    "query": "string (required)",
    "limit": "integer (optional)"
  }
}
```

| Parameter | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `query` | string | ✅ | Search query | Max 10000 characters (UTF-8) |
| `limit` | integer | ❌ | Max results | 1-1000, default: 10 |

#### Output Format

```json
{
  "results": [
    {
      "id": "doc-001",
      "content": "Matching content...",
      "score": 0.95,
      "created": "2024-01-15T10:30:00Z"
    }
  ],
  "query": "original query"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `results` | array | List of matching documents |
| `id` | string | Document identifier |
| `content` | string | Matching content |
| `score` | number | Relevance score (0-1) |
| `created` | string | Document creation timestamp |
| `query` | string | Original query (echoed) |

#### Search Algorithm

1. **Primary**: FTS5 (Full-Text Search 5) for SQLite
2. **Fallback**: LIKE-based search if FTS5 unavailable
3. **Scoring**: FTS5 `rank` for relevance

#### Example

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "semantic_search",
    "arguments": {
      "query": "authentication OAuth security",
      "limit": 5
    }
  },
  "id": 7
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"results\":[{\"id\":\"doc-001\",\"content\":\"OAuth2 implementation guide with security best practices...\",\"score\":0.95,\"created\":\"2024-01-15T10:30:00Z\"},{\"id\":\"doc-002\",\"content\":\"JWT authentication tutorial...\",\"score\":0.85,\"created\":\"2024-01-14T15:00:00Z\"}],\"query\":\"authentication OAuth security\"}"
    }]
  }
}
```

#### Validation Errors

| Error | Cause |
|-------|-------|
| `query is required` | Missing query parameter |
| `query exceeds maximum length of 10000 characters` | Query too long |

---

## Error Handling

### Error Codes

The MCP server follows JSON-RPC 2.0 error conventions:

| Code | Constant | Description |
|------|----------|-------------|
| `-32700` | `ParseError` | Invalid JSON received |
| `-32600` | `InvalidRequest` | Invalid request format |
| `-32601` | `MethodNotFound` | Unknown method or tool |
| `-32602` | `InvalidParams` | Invalid parameters |
| `-32603` | `InternalError` | Server-side error |
| `-32000` | `ServerError` | Application-specific error |

### Error Response Format

```json
{
  "jsonrpc": "2.0",
  "id": "original-request-id",
  "error": {
    "code": -32602,
    "message": "title is required"
  }
}
```

### Common Errors

| Code | Message | Cause | Resolution |
|------|---------|-------|------------|
| `-32602` | `title is required` | Missing required field | Include required parameters |
| `-32602` | `X exceeds maximum length of N characters` | Input too long | Shorten input |
| `-32602` | `invalid status: X` | Invalid enum value | Use valid status |
| `-32601` | `Tool not found: X` | Unknown tool name | Check tool name |
| `-32603` | `database error` | Database operation failed | Check database file |

---

## Troubleshooting

### Connection Issues

**Server doesn't respond:**
```bash
# Test basic connectivity
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./mcp-server
```

### Input Validation

**Getting validation errors:**
- Check UTF-8 encoding for non-ASCII characters
- Verify string lengths are within limits
- Ensure required fields are present

### Database Issues

**Database locked:**
```bash
# Ensure only one instance running
pkill mcp-server
./mcp-server
```

**FTS5 not working:**
```bash
# Check FTS5 status
sqlite3 ~/.tools-bank/data.db "SELECT * FROM embeddings_fts;"
```

### Performance

| Issue | Solution |
|-------|----------|
| Slow responses | Check database file size, consider cleanup |
| Large result sets | Use `limit` parameter |
| Memory issues | Reduce max message size in client |

---

## Appendix

### Database Schema

```sql
-- Tasks table
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    metadata TEXT
);

-- Memory table
CREATE TABLE memory (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    tags TEXT
);

-- Embeddings table
CREATE TABLE embeddings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    document_id TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding BLOB,
    model TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT
);
```

### Type Definitions

```typescript
// Task
interface Task {
  id: string;
  name: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

// MemoryEntry
interface MemoryEntry {
  key: string;
  value: string | null;
  found: boolean;
}

// SearchResult
interface SearchResult {
  id: string;
  content: string;
  score: number;
  created: string;
}
```
