# ProveMySelf Backend Makefile

.PHONY: dev build test test-int lint fmt openapi clean all

# Development
dev:
	@echo "Starting backend development server..."
	go run cmd/api/main.go

# Build
build:
	@echo "Building backend..."
	go build -o bin/api cmd/api/main.go

# Testing
test:
	@echo "Running unit tests..."
	go test -race -coverprofile=coverage.out ./...

test-int:
	@echo "Running integration tests..."
	go test -race -tags=integration ./test/...

# Linting and formatting
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	goimports -w .

# OpenAPI generation
openapi:
	@echo "Generating OpenAPI spec..."
	@echo "TODO: Add swag or similar tool to generate from Go comments"

# Cleanup
clean:
	@echo "Cleaning up..."
	rm -rf bin/ coverage.out

# Dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

all: fmt lint test build