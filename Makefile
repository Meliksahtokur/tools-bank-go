# Makefile for tools-bank-go MCP Server
# Production-ready build system

.PHONY: all build clean test test/cover test/coverage.html lint lint/fix fmt vet tidy deps install run build/all-platforms docker/build help

# ============================================================================
# Variables
# ============================================================================

BINARY_NAME       := mcp-server
MAIN_PATH         := cmd/mcp-server
BUILD_DIR         := ./bin
BUILD_FLAGS       := -ldflags "-s -w"
GO                := go
GOTEST            := $(GO) test
GOLANGCI_LINT     := golangci-lint
COVERAGE_OUT      := ./coverage.out
COVERAGE_HTML     := ./coverage.html
VERSION           ?=
BUILD_DATE        := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS_VERSION   :=
LDFLAGS_DATE      := -ldflags "-X main.buildDate=$(BUILD_DATE)"

# Build version ldflags if VERSION is set
ifeq ($(VERSION),)
	LDFLAGS := $(BUILD_FLAGS)
else
	LDFLAGS_VERSION := -X main.version=$(VERSION)
	LDFLAGS := $(BUILD_FLAGS) "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)"
endif

# Cross-platform targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# ============================================================================
# Default target
# ============================================================================

all: clean build lint test

# ============================================================================
# Build targets
# ============================================================================

## build: Build the binary
build:
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(MAIN_PATH)
	@echo "==> Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build/$(BINARY_NAME): Alias for build
build/$(BINARY_NAME): build

## clean: Remove build artifacts and test coverage
clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
	@echo "==> Clean complete"

# ============================================================================
# Testing targets
# ============================================================================

## test: Run all tests with race detector
test:
	@echo "==> Running tests with race detector..."
	@$(GOTEST) -v -race ./...

## test/cover: Run tests with coverage report
test/cover:
	@echo "==> Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=$(COVERAGE_OUT) -covermode=atomic ./...
	@echo "==> Coverage report: $(COVERAGE_OUT)"
	@$(GO) tool cover -func=$(COVERAGE_OUT)

## test/coverage.html: Generate HTML coverage report
test/coverage.html: test/cover
	@echo "==> Generating HTML coverage report..."
	@$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "==> HTML coverage report: $(COVERAGE_HTML)"

# ============================================================================
# Linting targets
# ============================================================================

## lint: Run golangci-lint
lint:
	@echo "==> Running golangci-lint..."
	@$(GOLANGCI_LINT) run ./...

## lint/fix: Run linter with auto-fix
lint/fix:
	@echo "==> Running golangci-lint with auto-fix..."
	@$(GOLANGCI_LINT) run --fix ./...

# ============================================================================
# Code quality targets
# ============================================================================

## fmt: Format code
fmt:
	@echo "==> Formatting code..."
	@$(GO) fmt ./...
	@echo "==> Format complete"

## vet: Run go vet
vet:
	@echo "==> Running go vet..."
	@$(GO) vet ./...
	@echo "==> Vet complete"

## tidy: Run go mod tidy
tidy:
	@echo "==> Running go mod tidy..."
	@$(GO) mod tidy
	@echo "==> Tidy complete"

# ============================================================================
# Dependency targets
# ============================================================================

## deps: Download dependencies
deps:
	@echo "==> Downloading dependencies..."
	@$(GO) mod download
	@echo "==> Dependencies complete"

# ============================================================================
# Installation targets
# ============================================================================

## install: Build and install to GOBIN
install: build
	@echo "==> Installing $(BINARY_NAME)..."
	@$(GO) install $(LDFLAGS) ./$(MAIN_PATH)
	@echo "==> Install complete"

# ============================================================================
# Run targets
# ============================================================================

## run: Build and run the server
run: build
	@echo "==> Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# ============================================================================
# Cross-platform build
# ============================================================================

## build/all-platforms: Cross-compile for multiple platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
build/all-platforms:
	@echo "==> Cross-compiling for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os_arch=$${platform}; \
		os=$${os_arch%/*}; \
		arch=$${os_arch#*/}; \
		output="$(BUILD_DIR)/$(BINARY_NAME)-$${os}-$${arch}"; \
		echo "==> Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch $(GO) build $(LDFLAGS) -o $$output ./$(MAIN_PATH); \
		echo "==> Created $$output"; \
	done
	@echo "==> All platforms built successfully"

# ============================================================================
# Docker targets
# ============================================================================

docker/build:
	@if [ -f Dockerfile ]; then \
		echo "==> Building Docker image..."; \
		docker build -t $(BINARY_NAME):latest .; \
		echo "==> Docker build complete"; \
	else \
		echo "==> ERROR: Dockerfile not found"; \
		exit 1; \
	fi

# ============================================================================
# Help
# ============================================================================

## help: Display this help message
help:
	@echo ""
	@echo "Makefile for $(BINARY_NAME)"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all                  Build all targets (default)"
	@echo "  build                Build the binary"
	@echo "  clean                Remove build artifacts and test coverage"
	@echo "  test                 Run all tests with race detector"
	@echo "  test/cover           Run tests with coverage report"
	@echo "  test/coverage.html   Generate HTML coverage report"
	@echo "  lint                 Run golangci-lint"
	@echo "  lint/fix             Run linter with auto-fix"
	@echo "  fmt                  Format code"
	@echo "  vet                  Run go vet"
	@echo "  tidy                 Run go mod tidy"
	@echo "  deps                 Download dependencies"
	@echo "  install              Build and install to GOBIN"
	@echo "  run                  Build and run the server"
	@echo "  build/all-platforms  Cross-compile for multiple platforms"
	@echo "  docker/build         Build Docker image (if Dockerfile exists)"
	@echo "  help                 Display this help message"
	@echo ""
	@echo "Variables:"
	@echo "  BINARY_NAME  = $(BINARY_NAME)"
	@echo "  MAIN_PATH    = $(MAIN_PATH)"
	@echo "  BUILD_DIR    = $(BUILD_DIR)"
	@echo "  VERSION      = $(VERSION)"
	@echo ""