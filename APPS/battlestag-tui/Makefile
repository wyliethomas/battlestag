# Makefile for battlestag TUI

# Binary name
BINARY_NAME=battlestag

# Installation paths
INSTALL_DIR=$(HOME)/.local/bin
CONFIG_DIR=$(HOME)/.config/battlestag-tui

# Build variables
GO=go
GOFLAGS=-ldflags="-s -w"
BUILD_DIR=.

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

# Install the binary and default config
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@mkdir -p $(CONFIG_DIR)
	@cp $(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete!"
	@echo ""
	@echo "The binary has been installed to: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "Config directory: $(CONFIG_DIR)"
	@echo ""
	@echo "Make sure $(INSTALL_DIR) is in your PATH."
	@echo "You can run '$(BINARY_NAME)' from any terminal."
	@echo ""
	@echo "On first run, you'll be prompted to configure your Agent Gateway connection."

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Binary removed from $(INSTALL_DIR)"
	@echo ""
	@echo "To remove config files, run: rm -rf $(CONFIG_DIR)"
	@echo ""

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test -v ./...

# Run the application (for development)
.PHONY: run
run: build
	@./$(BINARY_NAME)

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	@$(GO) vet ./...

# Help
.PHONY: help
help:
	@echo "Battlestag TUI - Makefile targets:"
	@echo ""
	@echo "  make build      - Build the binary"
	@echo "  make install    - Build and install to ~/.local/bin"
	@echo "  make uninstall  - Remove the installed binary"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make run        - Build and run the application"
	@echo "  make fmt        - Format code"
	@echo "  make vet        - Vet code"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Installation paths:"
	@echo "  Binary:  $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "  Config:  $(CONFIG_DIR)/"
	@echo ""
