# Contributing to tools-bank-go

Thank you for your interest in contributing to tools-bank-go! This document provides comprehensive guidelines and instructions for contributing to this project.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Git Guidelines](#git-guidelines)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)
- [Release Process](#release-process)
- [Questions and Support](#questions-and-support)

---

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

**Expected Behavior:**
- Be respectful and inclusive
- Focus on constructive feedback
- Show empathy towards other contributors

**Unacceptable Behavior:**
- Harassment or discrimination
- Personal attacks
- Publishing private information

---

## Getting Started

### Before You Begin

1. **Check existing issues** - Your issue or feature may already be tracked
2. **Open a discussion** - For significant changes, start a GitHub Discussion
3. **Fork the repository** - Create your own fork to work in

### First-Time Setup

```bash
# 1. Fork via GitHub UI or CLI
gh repo fork egut/tools-bank-go

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/tools-bank-go.git
cd tools-bank-go

# 3. Add upstream remote
git remote add upstream https://github.com/egesut/tools-bank-go.git

# 4. Verify remotes
git remote -v
```

---

## Development Setup

### Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | Required |
| Git | Any recent | For version control |
| Make | Optional | For convenience commands |

### Installation

```bash
# Install dependencies
go mod download

# Verify setup builds
go build -o mcp-server ./cmd/mcp-server

# Run the server briefly
timeout 1 ./mcp-server || true

# Run tests
go test -v ./...
```

### Recommended IDE Setup

**VS Code / Cursor:**
```json
{
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "go.useLanguageServer": true,
  "editor.formatOnSave": true
}
```

**Goland:**
- Enable "Format on save"
- Install golangci-lint plugin
- Configure Go SDK 1.21+

---

## Code Style

### Formatting Rules

All code must be formatted with `gofmt`:

```bash
# Format all files (in-place)
gofmt -w .

# Check formatting without modifying
gofmt -d .
```

### Linting

We use `golangci-lint` for comprehensive linting:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run ./...

# Auto-fix issues
golangci-lint run --fix ./...
```

### Style Guidelines

| Guideline | Description | Example |
|-----------|-------------|---------|
| **Naming** | Use descriptive names | `userID` not `uid` |
| **Error handling** | Always handle errors explicitly | Wrap with context |
| **Comments** | Document public APIs | `// FetchUser retrieves...` |
| **Imports** | Group: stdlib, external, internal | See import organization |
| **Line length** | Keep under 100 characters | Break long lines |

### Import Organization

```go
import (
    // Standard library (alphabetical)
    "encoding/json"
    "fmt"
    "os"

    // External packages (alphabetical)
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    // Internal packages (alphabetical)
    "github.com/egesut/tools-bank-go/pkg/db"
)
```

### Error Wrapping

```go
// ✅ Good - Descriptive context
if err != nil {
    return nil, fmt.Errorf("failed to open database: %w", err)
}

// ❌ Bad - Lost error context
if err != nil {
    return nil, err
}

// ✅ Good - Structured errors
if err != nil {
    return nil, fmt.Errorf("task_create: %w", err)
}
```

### Documentation Comments

```go
// Server represents the MCP server instance.
// It handles JSON-RPC requests via stdin/stdout and manages
// tool registration and execution.
type Server struct {
    tools map[string]ToolHandler
    mu    sync.RWMutex
    db    *db.DB
}

// NewServer creates a new MCP server instance with default tools registered.
func NewServer() *Server {
    // ...
}
```

---

## Git Guidelines

### Branch Naming

```
feature/<short-description>    # New features
fix/<short-description>        # Bug fixes
docs/<short-description>       # Documentation
refactor/<short-description>  # Code refactoring
test/<short-description>       # Test improvements
chore/<short-description>      # Maintenance tasks
```

Examples:
- `feature/add-user-preferences`
- `fix/memory-leak-on-restart`
- `docs/update-api-reference`

### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

#### Types

| Type | Description | When to Use |
|------|-------------|-------------|
| `feat` | New feature | Adding new functionality |
| `fix` | Bug fix | Fixing existing bugs |
| `docs` | Documentation | README, comments, etc. |
| `style` | Formatting | No logic changes |
| `refactor` | Refactoring | Code restructuring |
| `perf` | Performance | Performance improvements |
| `test` | Tests | Adding or updating tests |
| `chore` | Maintenance | Build, tooling, CI |

#### Examples

**Feature:**
```
feat(memory): add TTL support for memory keys

Added optional expires_at parameter to memory_set.
When expires_at is set, keys automatically expire.

Closes #123
```

**Bug Fix:**
```
fix(task): handle empty status in task_update

Prevented panic when status field is missing in request.
Now returns proper validation error instead.

Fixes #456
```

**Documentation:**
```
docs: update API reference for semantic_search

Added FTS5 fallback behavior to documentation.
Included example responses for edge cases.
```

### Commit Guidelines

1. **Atomic commits** - One logical change per commit
2. **Meaningful messages** - Explain *why*, not just *what*
3. **Imperative mood** - "Add feature" not "Added feature"
4. **First line limit** - Keep subject under 50 characters
5. **Reference issues** - Include issue numbers in footer

---

## Testing Requirements

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific package
go test -v ./pkg/mcp/...

# Run specific test
go test -v -run TestTaskCreate ./pkg/mcp/...
```

### Test Structure

```
pkg/
├── mcp/
│   ├── server.go           # Implementation
│   └── server_test.go      # Tests (required)
├── db/
│   ├── sqlite.go           # Implementation
│   └── sqlite_test.go      # Tests (required)
└── utils/
    ├── logger.go           # Implementation
    └── logger_test.go      # Tests (if applicable)
```

### Writing Tests

Use `testify` for assertions:

```go
package mcp

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTaskCreate(t *testing.T) {
    server := NewServer()

    result, err := server.tools["task_create"](map[string]interface{}{
        "title": "Test Task",
    })

    require.NoError(t, err)
    assert.True(t, result.(map[string]interface{})["success"].(bool))
}

func TestTaskCreateValidation(t *testing.T) {
    tests := []struct {
        name    string
        args    map[string]interface{}
        wantErr string
    }{
        {
            name:    "empty title",
            args:    map[string]interface{}{},
            wantErr: "title is required",
        },
        {
            name:    "title too long",
            args:    map[string]interface{}{"title": strings.Repeat("x", 300)},
            wantErr: "exceeds maximum length",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := NewServer()
            _, err := server.tools["task_create"](tt.args)
            assert.ErrorContains(t, err, tt.wantErr)
        })
    }
}
```

### When to Use assert vs require

| Function | Behavior | Use When |
|----------|----------|----------|
| `assert` | Continue on failure | Multiple independent checks |
| `require` | Fail fast on failure | Prerequisites, setup |

### Coverage Requirements

| Package | Minimum Coverage |
|---------|-----------------|
| `pkg/mcp` | 80% |
| `pkg/db` | 80% |
| `pkg/utils` | 70% |

### Test Naming Conventions

```
Test<Unit>                  # Basic unit test
Test<Unit>_<Scenario>       # Scenario-specific test
Test<Unit>_<Scenario>_<Expected> # Expected outcome
```

---

## Pull Request Process

### Before Submitting

1. **Run all tests**
   ```bash
   go test -race ./...
   ```

2. **Run linter**
   ```bash
   golangci-lint run ./...
   ```

3. **Format code**
   ```bash
   gofmt -w .
   ```

4. **Add tests** for new functionality

5. **Update documentation** if needed

6. **Sync with upstream**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

### PR Description Template

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Motivation and Context
Why is this change needed? What problem does it solve?

## How Has This Been Tested?
Describe how the change was tested.

## Screenshots (if applicable)
Include screenshots for UI changes.

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-reviewed
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No warnings or errors
- [ ] Commit messages follow conventions
```

### Review Process

1. **Automated checks** - CI must pass
2. **Code review** - Maintainer review
3. **Address feedback** - Make requested changes
4. **Approval** - At least one maintainer approval
5. **Merge** - Maintainer merges

### After Merging

```bash
# Sync your fork
git checkout main
git fetch upstream
git merge upstream/main
git push origin main
```

---

## Project Structure

```
tools-bank-go/
├── cmd/
│   └── mcp-server/
│       └── main.go           # Entry point, initialization
├── pkg/
│   ├── config/               # Configuration management
│   │   └── config.go
│   ├── db/                   # Database layer
│   │   ├── sqlite.go         # SQLite operations
│   │   └── sqlite_test.go
│   ├── mcp/                  # MCP protocol
│   │   ├── server.go         # Server implementation
│   │   └── server_test.go    # Comprehensive tests
│   └── utils/                # Utilities
│       ├── errors.go
│       └── logger.go
├── docs/                     # Documentation
│   └── API.md
├── scripts/                  # Utility scripts
├── .env.example              # Environment template
├── .gitignore
├── .golangci.yml             # Linter config
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Package Responsibilities

| Package | Responsibility | Dependencies |
|---------|----------------|--------------|
| `cmd/mcp-server` | Application entry, DI | All pkg/* |
| `pkg/config` | Environment config | None |
| `pkg/db` | SQLite operations | modernc.org/sqlite |
| `pkg/mcp` | Protocol handling | pkg/db |
| `pkg/utils` | Logging, errors | None |

---

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH
1.2.3
│  │  │
│  │  └── Patch: Bug fixes
│  └───── Minor: New features (backward compatible)
└──────── Major: Breaking changes
```

### Release Checklist

1. [ ] Update version in code (if applicable)
2. [ ] Update CHANGELOG.md
3. [ ] Create release branch: `git checkout -b release/v1.2.0`
4. [ ] Run full test suite
5. [ ] Update documentation
6. [ ] Create GitHub release with tag
7. [ ] Merge to main

### Git Tags

```bash
# Create tag
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

---

## Questions and Support

- **Bug Reports**: [GitHub Issues](https://github.com/egesut/tools-bank-go/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/egesut/tools-bank-go/discussions)
- **General Questions**: [GitHub Discussions](https://github.com/egesut/tools-bank-go/discussions)

### Response Times

| Type | Response Time |
|------|--------------|
| Bug reports | Within 48 hours |
| Feature requests | Within 1 week |
| Pull requests | Within 1 week |

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🎉
