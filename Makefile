# Makefile for Pizzeria Project

.PHONY: all build clean test run run-main air dev help

# Configuration
GO=go
PKG=github.com/AlexTLDR/pizzeria
MAIN_PATH=./cmd/server
BIN_NAME=server
BIN_DIR=bin
STATIC_CSS_INPUT=./static/css/input.css
STATIC_CSS_OUTPUT=./static/css/output.css

# Colors
GREEN=\033[0;32m
RED=\033[0;31m
YELLOW=\033[1;33m
NC=\033[0m

all: build

help:
	@echo "Pizzeria Project Makefile"
	@echo "-------------------------"
	@echo "Available commands:"
	@echo "  make build      - Build Go binary and CSS"
	@echo "  make build-go   - Build only Go binary"
	@echo "  make build-css  - Build only CSS"
	@echo "  make run        - Build and run the application"
	@echo "  make run-main   - Run main.go directly without building"
	@echo "  make air        - Run with Air for hot reload"
	@echo "  make dev        - Run in development mode with CSS watching"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run all tests"
	@echo "  make test-v     - Run tests with verbose output"
	@echo "  make test-race  - Run tests with race detection"
	@echo "  make test-cover - Run tests with coverage report"

# Build commands
build: build-css build-go

build-go:
	@echo "Building Go application..."
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(BIN_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build successful!$(NC)"

build-css:
	@echo "Building CSS..."
	npm run build-css
	@echo "$(GREEN)CSS built successfully!$(NC)"

# Run commands
run: build
	@echo "Starting server..."
	./$(BIN_DIR)/$(BIN_NAME)

run-main:
	@echo "Running main.go directly..."
	$(GO) run $(MAIN_PATH)/main.go

air:
	@echo "Starting with Air hot reload..."
	air

dev:
	@echo "Starting development server with hot reload..."
	npm run dev

# Test commands
test:
	@echo "Running tests..."
	$(GO) test $(PKG)/...

test-v:
	@echo "Running tests with verbose output..."
	$(GO) test -v $(PKG)/...

test-race:
	@echo "Running tests with race detection..."
	$(GO) test -race $(PKG)/...

test-cover:
	@echo "Running tests with coverage..."
	$(GO) test -cover $(PKG)/...
	$(GO) test -coverprofile=coverage.out $(PKG)/...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated at coverage.html$(NC)"

# Clean command
clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR) $(STATIC_CSS_OUTPUT) coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"