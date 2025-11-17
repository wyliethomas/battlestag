.PHONY: build run test clean install help

# Binary name
BINARY_NAME=agent-gateway
VERSION=1.0.0

# Build directory
BUILD_DIR=build

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

run: ## Run the application with example config
	@go run main.go --config config.yaml

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf logs/*.log
	@echo "Clean complete"

install: build ## Install the binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

install-service: ## Install systemd service
	@echo "Installing systemd service..."
	@sudo cp agent-gateway.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@echo "Service installed. Enable with: sudo systemctl enable agent-gateway"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded"

dev: ## Run in development mode
	@go run main.go --config config.yaml

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run || echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

all: clean deps fmt build test ## Clean, download deps, format, build, and test

.DEFAULT_GOAL := help
