.PHONY: all build clean install test

# Build directory
BIN_DIR := bin

# Executables
SLACK := $(BIN_DIR)/slack
GMAIL := $(BIN_DIR)/gmail
TRELLO := $(BIN_DIR)/trello
NOTEWRITER := $(BIN_DIR)/notewriter

all: build

build: $(SLACK) $(GMAIL) $(TRELLO) $(NOTEWRITER)

$(SLACK):
	@mkdir -p $(BIN_DIR)
	@echo "Building slack connector..."
	@go build -o $(SLACK) ./cmd/slack

$(GMAIL):
	@mkdir -p $(BIN_DIR)
	@echo "Building gmail connector..."
	@go build -o $(GMAIL) ./cmd/gmail

$(TRELLO):
	@mkdir -p $(BIN_DIR)
	@echo "Building trello connector..."
	@go build -o $(TRELLO) ./cmd/trello

$(NOTEWRITER):
	@mkdir -p $(BIN_DIR)
	@echo "Building notewriter..."
	@go build -o $(NOTEWRITER) ./cmd/notewriter

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)

test:
	@echo "Running tests..."
	@go test ./...

install:
	@echo "Installing executables to /usr/local/bin..."
	@cp $(SLACK) /usr/local/bin/com-observer-slack
	@cp $(GMAIL) /usr/local/bin/com-observer-gmail
	@cp $(TRELLO) /usr/local/bin/com-observer-trello
	@cp $(NOTEWRITER) /usr/local/bin/com-observer-notewriter

# Build with version info
version ?= dev
build-release:
	@mkdir -p $(BIN_DIR)
	@echo "Building release version $(version)..."
	@go build -ldflags "-X main.version=$(version)" -o $(SLACK) ./cmd/slack
	@go build -ldflags "-X main.version=$(version)" -o $(GMAIL) ./cmd/gmail
	@go build -ldflags "-X main.version=$(version)" -o $(TRELLO) ./cmd/trello
	@go build -ldflags "-X main.version=$(version)" -o $(NOTEWRITER) ./cmd/notewriter

help:
	@echo "Available targets:"
	@echo "  make build         - Build all executables"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make install       - Install to /usr/local/bin"
	@echo "  make build-release - Build with version info"
