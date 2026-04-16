# tools-bank-go

> A high-performance MCP (Model Context Protocol) server for Go, providing task management, memory storage, and semantic search capabilities with SQLite persistence.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/egesut/tools-bank-go/ci.yml?branch=main&style=flat-square)](https://github.com/egesut/tools-bank-go/actions)
[![Coverage](https://img.shields.io/badge/coverage-80%25-brightgreen?style=flat-square)](#testing)

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#-features)
- [Architecture](#-architecture)
- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [API Reference](#-api-reference)
- [Development](#-development)
- [Deployment](#-deployment)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)
- [License](#-license)

---

## Overview

**tools-bank-go** is a production-ready MCP server that enables AI assistants (like Claude, Cursor) to interact with persistent storage through a standardized JSON-RPC 2.0 interface over stdin/stdout.

### What is MCP?

The [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) is an open protocol that enables seamless integration between AI assistants and data sources. It provides a standardized way for AI models to:

- Access tools and capabilities
- Exchange structured data
- Maintain context across interactions

### Why tools-bank-go?

| Feature | Benefit |
|---------|---------|
| **Pure Go** | Single binary, no runtime dependencies |
| **SQLite** | Zero-config, embedded persistence |
| **FTS5 Search** | Fast full-text search without external services |
| **Thread-Safe** | Concurrent request handling |
| **MCP Compliant** | Works with any MCP-compatible client |

---

## 🔥 Features

### Core MCP Tools

| Category | Tools | Description |
|----------|-------|-------------|
| **Task Management** | `task_create` | Create tasks with title and description |
| | `task_list` | List all tasks with optional status filter |
| | `task_update` | Update task status |
| | `task_delete` | Delete tasks by ID |
| **Memory Store** | `memory_get` | Retrieve values by key |
| | `memory_set` | Store key-value pairs |
| **Search** | `semantic_search` | Full-text search with FTS5 |

### Key Capabilities

- 🚀 **High Performance** - Built with Go for low latency and high throughput
- 💾 **Persistent Storage** - SQLite-based with FTS5 full-text search
- 🔍 **Semantic Search** - FTS5 virtual tables for fast text matching
- 📦 **Extensible** - Register custom tools at runtime
- 🔌 **MCP Compliant** - Full JSON-RPC 2.0 protocol implementation
- 🛡️ **Input Validation** - UTF-8 aware length limits and type checking
- ⚙️ **Zero Dependencies** - Standalone binary, no runtime required

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Client                               │
│                    (Claude, Cursor, etc.)                       │
└────────────────────────────┬────────────────────────────────────┘
                             │ JSON-RPC 2.0 (stdin/stdout)
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     MCP Server (tools-bank-go)                  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Server    │  │   Tools     │  │     Tool Handlers       │ │
│  │  (server.go)│  │   Registry  │  │  - task_* (CRUD)        │ │
│  │             │  │             │  │  - memory_* (K/V)       │ │
│  │             │  │             │  │  - semantic_search      │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
│                           │                                     │
│  ┌────────────────────────────────────────────────────────────┐│
│  │              Database Layer (pkg/db)                       ││
│  │  ┌──────────┐  ┌──────────┐  ┌─────────────────────────┐    ││
│  │  │  Tasks   │  │  Memory  │  │      Embeddings        │    ││
│  │  │   Table  │  │   Table  │  │  (FTS5 Virtual Table)  │    ││
│  │  └──────────┘  └──────────┘  └─────────────────────────┘    ││
│  └────────────────────────────────────────────────────────────┘│
│                           │                                     │
│  ┌────────────┐                                                  │
│  │   SQLite   │                                                  │
│  │  (local)   │                                                  │
│  └────────────┘                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Package | File | Purpose |
|---------|------|---------|
| `cmd/mcp-server` | `main.go` | Application entry point, initialization |
| `pkg/mcp` | `server.go` | MCP protocol handling, tool registry |
| `pkg/db` | `sqlite.go` | SQLite operations, schema management |
| `pkg/utils` | Various | Logging, error utilities |

---

## ⚡ Quick Start

### Prerequisites

- **Go** 1.21 or higher
- **OS**: Linux, macOS, or Windows

### 30-Second Setup

```bash
# Clone and build
git clone https://github.com/egesut/tools-bank-go.git
cd tools-bank-go

# Build the server
go build -o mcp-server ./cmd/mcp-server

# Run (stdin/stdout communication)
./mcp-server
```

### Testing the Server

Send a JSON-RPC request via stdin:

```bash
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./mcp-server
```

---

## 📥 Installation

### Option 1: Install via Go

```bash
go install github.com/egesut/tools-bank-go/cmd/mcp-server@latest
```

### Option 2: Download Binary

Download from the [releases page](https://github.com/egesut/tools-bank-go/releases):

```bash
# Linux x86_64
curl -L https://github.com/egesut/tools-bank-go/releases/latest/download/mcp-server-linux-amd64 -o mcp-server
chmod +x mcp-server
sudo mv mcp-server /usr/local/bin/

# macOS ARM64 (Apple Silicon)
curl -L https://github.com/egesut/tools-bank-go/releases/latest/download/mcp-server-darwin-arm64 -o mcp-server
chmod +x mcp-server
sudo mv mcp-server /usr/local/bin/
```

### Option 3: Build from Source

```bash
git clone https://github.com/egesut/tools-bank-go.git
cd tools-bank-go
go build -o mcp-server ./cmd/mcp-server
```

### Option 4: Use Make

```bash
# Full build with lint and test
make all

# Just build
make build

# Cross-compile for all platforms
make build/all-platforms
```

---

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_PATH` | `~/.tools-bank/data.db` | SQLite database file path |
| `LOG_LEVEL` | `info` | Logging level: `debug`, `info`, `warn`, `error` |

### .env File (Optional)

```bash
# Copy example configuration
cp .env.example .env

# Edit as needed
nano .env
```

### Database Path Resolution

1. If `$HOME/.tools-bank/data.db` exists, it's used
2. Otherwise, `./data.db` in current directory
3. Custom path can be set via `DATABASE_PATH` environment variable

---

## 📚 API Reference

### Protocol

- **Transport**: stdin/stdout (stdio)
- **Format**: JSON-RPC 2.0
- **Protocol Version**: 2024-11-05

### Methods

| Method | Description |
|--------|-------------|
| `initialize` | Initialize the MCP session |
| `tools/list` | List available tools |
| `tools/call` | Execute a tool |
| `shutdown` | Gracefully shutdown the server |

### Quick Examples

#### Initialize

```json
{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05"},"id":1}
```

#### List Tools

```json
{"jsonrpc":"2.0","method":"tools/list","id":2}
```

#### Create Task

```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"task_create","arguments":{"title":"My Task","description":"Task description"}},"id":3}
```

### Error Codes

| Code | Name | Description |
|------|------|-------------|
| `-32700` | ParseError | Invalid JSON received |
| `-32600` | InvalidRequest | Malformed request |
| `-32601` | MethodNotFound | Unknown method |
| `-32602` | InvalidParams | Invalid parameters |
| `-32603` | InternalError | Server error |

---

## 🔧 Development

### Prerequisites

- **Go** 1.21+
- **Git**
- **Make** (optional, for convenience commands)

### Setup

```bash
# Clone repository
git clone https://github.com/egesut/tools-bank-go.git
cd tools-bank-go

# Download dependencies
go mod download

# Verify build
go build -o mcp-server ./cmd/mcp-server
```

### Makefile Commands

```bash
make build              # Build binary
make test               # Run tests with race detector
make test/cover         # Run tests with coverage
make lint               # Run golangci-lint
make lint/fix           # Auto-fix lint issues
make fmt                # Format code
make tidy               # Clean go.mod
make clean              # Remove artifacts
make all                # Full CI pipeline
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run with race detector
go test -race ./...

# Run specific package
go test -v ./pkg/mcp/...
```

### Project Structure

```
tools-bank-go/
├── cmd/
│   └── mcp-server/
│       └── main.go           # Entry point
├── pkg/
│   ├── config/               # Configuration
│   ├── db/
│   │   └── sqlite.go         # SQLite layer
│   │   └── sqlite_test.go
│   ├── mcp/
│   │   ├── server.go         # MCP server
│   │   └── server_test.go    # Comprehensive tests
│   └── utils/                # Logging, errors
├── docs/
│   └── API.md                # API documentation
├── scripts/                  # Utility scripts
├── .env.example
├── go.mod
├── Makefile
└── README.md
```

---

## 🚀 Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mcp-server ./cmd/mcp-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite-libs
WORKDIR /app
COPY --from=builder /app/mcp-server .
COPY --from=builder /app/.env.example .env
CMD ["./mcp-server"]
```

```bash
# Build and run
docker build -t tools-bank-go:latest .
docker run -d --name mcp-server tools-bank-go:latest
```

### Systemd Service

```ini
[Unit]
Description=Tools Bank MCP Server
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/opt/tools-bank-go
ExecStart=/opt/tools-bank-go/mcp-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable mcp-server
sudo systemctl start mcp-server
```

### Client Configuration

#### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "tools-bank": {
      "command": "/path/to/mcp-server",
      "args": []
    }
  }
}
```

#### Cursor

Add to Cursor settings under "MCP Servers":

```json
{
  "tools-bank": {
    "command": "/path/to/mcp-server"
  }
}
```

---

## 🔍 Troubleshooting

### Common Issues

#### Server Not Responding

```bash
# Verify binary works
./mcp-server --version 2>/dev/null || echo "Binary issue"

# Check stdin/stdout works
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./mcp-server
```

#### Database Locked

```bash
# Ensure only one instance runs
pkill mcp-server
./mcp-server
```

#### FTS5 Search Not Working

```bash
# Check database integrity
sqlite3 ~/.tools-bank/data.db "PRAGMA integrity_check;"

# Rebuild FTS index
sqlite3 ~/.tools-bank/data.db "INSERT INTO embeddings_fts(embeddings_fts) VALUES('rebuild');"
```

### Debug Mode

```bash
LOG_LEVEL=debug ./mcp-server
```

### Health Check

```bash
# Test with JSON-RPC
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | ./mcp-server
```

### Performance

| Metric | Target |
|--------|--------|
| Response Time | < 10ms |
| Concurrent Requests | Thread-safe |
| Max Message Size | 1MB |

---

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for:

- Development setup instructions
- Code style guidelines
- Commit message format
- Pull request process
- Testing requirements

### Quick Contribution Workflow

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/tools-bank-go.git
cd tools-bank-go

# Create feature branch
git checkout -b feature/your-feature-name

# Make changes and test
make test lint fmt

# Commit (Conventional Commits)
git commit -m "feat(scope): add new feature"

# Push and create PR
git push origin feature/your-feature-name
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  Built with ❤️ for the MCP community<br>
  <a href="https://github.com/egesut/tools-bank-go">GitHub</a> ·
  <a href="https://github.com/egesut/tools-bank-go/issues">Issues</a> ·
  <a href="https://github.com/egesut/tools-bank-go/discussions">Discussions</a>
</p>
