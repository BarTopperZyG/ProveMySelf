# ProveMySelf Monorepo Makefile

.PHONY: dev build test test-int lint typecheck fmt openapi clean install help

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

# Default target
help:
	@echo "ProveMySelf Development Commands:"
	@echo "  $(GREEN)make dev$(NC)       - Start all services in development mode"
	@echo "  $(GREEN)make build$(NC)     - Build all services"
	@echo "  $(GREEN)make test$(NC)      - Run all tests"
	@echo "  $(GREEN)make test-int$(NC)  - Run integration tests"
	@echo "  $(GREEN)make lint$(NC)      - Run linters for all projects"
	@echo "  $(GREEN)make typecheck$(NC) - Run TypeScript type checking"
	@echo "  $(GREEN)make fmt$(NC)       - Format all code"
	@echo "  $(GREEN)make openapi$(NC)   - Generate OpenAPI client"
	@echo "  $(GREEN)make install$(NC)   - Install all dependencies"
	@echo "  $(GREEN)make clean$(NC)     - Clean build artifacts"

# Install dependencies
install:
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	pnpm install
	cd backend/go && go mod download && go mod tidy

# Development - start all services
dev:
	@echo "$(YELLOW)Starting all services in development mode...$(NC)"
	@echo "Backend: http://localhost:8080"
	@echo "Studio: http://localhost:3000"
	@echo "Player: http://localhost:3001"
	@echo ""
	@make -j3 dev-backend dev-studio dev-player

dev-backend:
	@echo "$(GREEN)[Backend]$(NC) Starting Go API server..."
	cd backend/go && make dev

dev-studio:
	@echo "$(GREEN)[Studio]$(NC) Starting Next.js studio..."
	pnpm --filter frontend/studio dev

dev-player:
	@echo "$(GREEN)[Player]$(NC) Starting Next.js player..."
	pnpm --filter frontend/player dev --port 3001

# Build all services
build:
	@echo "$(YELLOW)Building all services...$(NC)"
	cd backend/go && make build
	pnpm --filter frontend/studio build
	pnpm --filter frontend/player build

# Test all services
test:
	@echo "$(YELLOW)Running all tests...$(NC)"
	cd backend/go && make test
	pnpm --filter frontend/studio test
	pnpm --filter frontend/player test
	pnpm --filter packages/schemas test

# Integration tests
test-int:
	@echo "$(YELLOW)Running integration tests...$(NC)"
	cd backend/go && make test-int

# Lint all code
lint:
	@echo "$(YELLOW)Running linters...$(NC)"
	cd backend/go && make lint
	pnpm --filter frontend/studio lint
	pnpm --filter frontend/player lint
	pnpm --filter packages/schemas lint

# TypeScript type checking
typecheck:
	@echo "$(YELLOW)Running TypeScript type checks...$(NC)"
	pnpm --filter frontend/studio typecheck
	pnpm --filter frontend/player typecheck
	pnpm --filter packages/schemas typecheck

# Format all code
fmt:
	@echo "$(YELLOW)Formatting all code...$(NC)"
	cd backend/go && make fmt
	pnpm prettier --write "frontend/**/*.{ts,tsx,js,jsx,json,md}"
	pnpm prettier --write "packages/**/*.{ts,tsx,js,jsx,json,md}"

# Generate OpenAPI client
openapi:
	@echo "$(YELLOW)Generating OpenAPI client...$(NC)"
	cd backend/go && make openapi
	pnpm --filter packages/openapi-client generate

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	cd backend/go && make clean
	pnpm --filter frontend/studio clean || true
	pnpm --filter frontend/player clean || true
	rm -rf node_modules/.cache
	rm -rf */node_modules/.cache

# Health check - verify all services are responsive
health:
	@echo "$(YELLOW)Checking service health...$(NC)"
	@echo "Backend API:"
	@curl -s http://localhost:8080/api/v1/health || echo "❌ Backend not responding"
	@echo ""
	@echo "Studio Frontend:"
	@curl -s http://localhost:3000/health || echo "❌ Studio not responding"
	@echo ""
	@echo "Player Frontend:"
	@curl -s http://localhost:3001/health || echo "❌ Player not responding"