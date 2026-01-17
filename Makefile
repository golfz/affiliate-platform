.PHONY: init mu md seed clean-db start start-backend start-frontend stop clean test lint swagger build help

# Default target
.DEFAULT_GOAL := help

# Variables
GO_VERSION := 1.21
NODE_VERSION := 18
MIGRATE_VERSION := v4.16.2
SWAG_VERSION := latest

# Database connection parameters (can be overridden by environment variables)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= jonosize
DB_PASSWORD ?= jonosize_dev
DB_NAME ?= jonosize
DB_SSLMODE ?= disable

# Build database connection URL for migrations (golang-migrate uses postgres://, not postgresql://)
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

init: ## Initialize project (install dependencies, setup config)
	@echo "🚀 Initializing project..."
	@echo "📦 Installing Go dependencies..."
	@go mod download
	@go mod tidy
	@echo "📦 Installing Node.js dependencies..."
	@cd apps/web && npm install || echo "⚠️  Frontend not set up yet, skipping..."
	@echo "🔧 Setting up config..."
	@if [ ! -f configs/config.json ]; then \
		cp configs/config.example.json configs/config.json; \
		echo "✅ Created configs/config.json from example"; \
	fi
	@echo "🐳 Starting Docker services..."
	@docker-compose up -d
	@echo "⏳ Waiting for database to be ready..."
	@sleep 5
	@echo "✅ Project initialized! Run 'make mu' to run migrations, then 'make start' to start the project."

mu: ## Run database migrations up
	@echo "🔄 Running database migrations..."
	@if ! which migrate > /dev/null 2>&1 || ! migrate -version | grep -q "postgres" 2>/dev/null; then \
		echo "📦 Installing migrate tool with PostgreSQL driver..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION); \
	fi
	@migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up || \
		(echo "⚠️  Migration failed. Make sure database is running (docker-compose up -d)" && exit 1)
	@echo "✅ Migrations completed!"

md: ## Run database migrations down (rollback)
	@echo "🔄 Rolling back database migrations..."
	@if ! which migrate > /dev/null 2>&1 || ! migrate -version | grep -q "postgres" 2>/dev/null; then \
		echo "📦 Installing migrate tool with PostgreSQL driver..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION); \
	fi
	@migrate -path migrations -database "$(DB_URL)" down
	@echo "✅ Migrations rolled back!"

seed: ## Seed database with demo data
	@echo "🌱 Seeding database..."
	@go run cmd/seed/main.go
	@echo "✅ Database seeded!"

clean-db: ## Clean all data from database (keeps schema)
	@echo "🧹 Cleaning database..."
	@docker exec -i jonosize-postgres psql -U $(DB_USER) -d $(DB_NAME) -c "DELETE FROM clicks; DELETE FROM links; DELETE FROM campaign_products; DELETE FROM offers; DELETE FROM campaigns; DELETE FROM products;" || \
		(echo "⚠️  Failed to clean database. Make sure database is running (docker-compose up -d)" && exit 1)
	@echo "✅ Database cleaned!"

start: ## Start both frontend and backend
	@echo "🚀 Starting project..."
	@echo "📝 Make sure you've run 'make init' and 'make mu' first!"
	@echo ""
	@echo "Starting backend and frontend in parallel..."
	@trap 'kill 0' EXIT; \
		(cd apps/web && npm run dev) & \
		(cd cmd/api && go run main.go) & \
		wait

start-backend: ## Start backend only
	@echo "🚀 Starting backend..."
	@cd cmd/api && go run main.go

start-frontend: ## Start frontend only
	@echo "🚀 Starting frontend..."
	@cd apps/web && npm run dev

stop: ## Stop Docker services
	@echo "🛑 Stopping Docker services..."
	@docker-compose down
	@echo "✅ Docker services stopped!"

clean: ## Clean up (stop services, remove volumes)
	@echo "🧹 Cleaning up..."
	@docker-compose down -v
	@echo "✅ Cleanup completed!"

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test ./... -v

lint: ## Run linters
	@echo "🔍 Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run || echo "⚠️  Linter not configured yet"

swagger: ## Generate Swagger docs
	@echo "📝 Generating Swagger documentation..."
	@which swag > /dev/null || (echo "Installing swag..." && \
		go install github.com/swaggo/swag/cmd/swag@latest)
	@swag init -g cmd/api/main.go -o docs
	@echo "✅ Swagger docs generated!"

build: ## Build backend binary
	@echo "🔨 Building backend..."
	@go build -o bin/api cmd/api/main.go
	@echo "✅ Backend built! Binary: bin/api"
