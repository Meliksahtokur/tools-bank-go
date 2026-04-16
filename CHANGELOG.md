# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

---

## [1.1.0] - YYYY-MM-DD

### Added

- Comprehensive test suite with 95% coverage
- Input validation with UTF-8 aware length checking
- Thread-safe tool registration with mutex protection
- FTS5 fallback to LIKE-based search
- Makefile with cross-platform build support
- GitHub Actions CI pipeline

### Changed

- Improved error messages with context
- Optimized database queries with proper indexing
- Updated protocol version to 2024-11-05

### Fixed

- Race condition in concurrent tool registration
- Memory leak in scanner buffer handling

### Security

- Added input sanitization for all user-provided strings

---

## [1.0.0] - YYYY-MM-DD

### Added

- Initial stable release
- Core MCP server implementation
- Task management tools:
  - `task_create` - Create tasks with title and description
  - `task_list` - List all tasks with status filter
  - `task_update` - Update task status
  - `task_delete` - Delete tasks by ID
- Memory store tools:
  - `memory_get` - Retrieve values by key
  - `memory_set` - Store key-value pairs
- Semantic search with FTS5 full-text search:
  - `semantic_search` - Full-text search with relevance scoring
- SQLite database backend with automatic schema migration
- Configuration via environment variables
- Comprehensive documentation (README, API docs, CONTRIBUTING)
- JSON-RPC 2.0 protocol implementation
- Graceful shutdown handling
- Signal handling for clean exit

### Changed

- N/A (initial release)

### Fixed

- N/A (initial release)

---

## [0.1.0] - YYYY-MM-DD

### Added

- Alpha release for community testing
- Basic MCP protocol support
- SQLite persistence layer
- Initial task and memory tools

### Changed

- N/A (initial release)

### Fixed

- N/A (initial release)

---

## Template

Use this template for new releases:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- 

### Changed
- 

### Deprecated
- 

### Removed
- 

### Fixed
- 

### Security
- 
```

---

## Version History

| Version | Date | Status | Notes |
|---------|------|--------|-------|
| 1.1.0 | YYYY-MM-DD | Current | Enhanced testing, validation, CI |
| 1.0.0 | YYYY-MM-DD | Stable | Initial stable release |
| 0.1.0 | YYYY-MM-DD | Alpha | Alpha release |

---

## Migration Guides

### Upgrading from v0.x to v1.0

#### Breaking Changes

1. **Protocol Version**: Updated to `2024-11-05`
2. **Tool Names**: Unchanged from v0.1.0

#### Recommended Steps

```bash
# Pull latest version
git pull origin main

# Rebuild
go build -o mcp-server ./cmd/mcp-server

# Test your integration
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./mcp-server
```

### Upgrading from v1.0 to v1.1

#### New Features

1. **Improved Input Validation**: UTF-8 length checking
2. **Enhanced Error Messages**: More descriptive error context
3. **Better Test Coverage**: 95% coverage for core packages

#### No Breaking Changes

v1.1 is fully backward compatible with v1.0.

---

## Categories Explained

| Category | When to Use |
|----------|-------------|
| **Added** | New features or capabilities |
| **Changed** | Changes to existing functionality |
| **Deprecated** | Features marked for future removal |
| **Removed** | Features actually removed |
| **Fixed** | Bug fixes |
| **Security** | Vulnerability or security improvements |

---

## Release Process

### Checklist

For each release, ensure:

- [ ] Update version in code (if applicable)
- [ ] Update CHANGELOG.md with all changes
- [ ] Create release branch: `release/vX.Y.Z`
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Update documentation if needed
- [ ] Create Git tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] Create GitHub release
- [ ] Merge to main

### GitHub Release Notes Template

```markdown
## What's New

Describe major changes in user-friendly terms.

## Breaking Changes

List any breaking changes if applicable.

## Bug Fixes

- Fix 1
- Fix 2

## Contributors

Thank you to all contributors!
```

---

## Deprecation Policy

When deprecating features:

1. Add deprecation notice in changelog
2. Document migration path
3. Maintain deprecated feature for at least 2 minor versions
4. Remove in next major version

Example deprecation entry:

```markdown
### Deprecated
- `legacy_search` - Use `semantic_search` instead. Will be removed in v2.0.
```

---

## Security Advisories

For security issues:

1. **Do NOT** open a public GitHub issue
2. **Instead**: Email maintainer directly
3. **Timeline**: 
   - Acknowledgment: 24-48 hours
   - Initial assessment: 1 week
   - Fix timeline: Varies by severity

After fix is released, security issues will be documented here with CVEs if applicable.

---

## Semantic Versioning Details

### Version Format

```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
1.2.3-beta.1+build.123
│ │ │ │    │         │
│ │ │ │    │         └── Build metadata
│ │ │ │    └────────── Pre-release version
│ │ │ └─────────────── Patch version
│ │ └───────────────── Minor version
└───────────────────── Major version
```

### When to Increment

| Increment | When |
|-----------|------|
| **MAJOR** | Breaking changes to API or protocol |
| **MINOR** | New features, backward compatible |
| **PATCH** | Bug fixes, backward compatible |

### Examples

| Change | Version Bump |
|--------|-------------|
| Add new tool | MINOR |
| Change tool parameter behavior | MAJOR |
| Fix bug in existing tool | PATCH |
| Add optional parameter | MINOR |
| Remove deprecated feature | MAJOR |
| Update documentation | None |
| Improve performance | MINOR |

---

## Maintaining This File

1. Add new entries at the top under `[Unreleased]`
2. Move `[Unreleased]` to version section on release
3. Include all contributors in release notes
4. Use imperative mood ("Add" not "Added")
5. Be specific but concise
