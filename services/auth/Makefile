# ============================================================================
# Makefile for AuthService
# ============================================================================
# Common commands:
#   make help          - Show this help message
#   make run           - Run the service locally
#   make test          - Run tests
#   make build         - Build binary
#   make docker-build  - Build Docker image
#   make docker-up     - Start all services with docker-compose
#   make docker-down   - Stop all services
# ============================================================================

.PHONY: help
help:
	@echo "AuthService - Available commands:"
	@echo ""
	@echo "  Development:"
	@echo "    make run              - Run service locally"
	@echo "    make dev              - Run with live reload (requires air)"
	@echo "    make test             - Run all tests"
	@echo "    make test-unit        - Run unit tests only"
	@echo "    make test-integration - Run integration tests"
	@echo "    make test-coverage    - Run tests with coverage"
	@echo ""
	@echo "  Build:"
	@echo "    make build            - Build binary"
	@echo "    make clean            - Clean build artifacts"
	@echo "    make proto            - Generate proto files"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-build     - Build Docker image"
	@echo "    make docker-up        - Start all services"
	@echo "    make docker-down      - Stop all services"
	@echo "    make docker-logs      - View service logs"
	@echo "    make docker-restart   - Restart auth service"
	@echo ""
	@echo "  Database:"
	@echo "    make db-init          - Initialize database"
	@echo "    make db-migrate       - Run migrations"
	@echo "    make db-seed          - Seed database"
	@echo "    make db-reset         - Reset database"
	@echo ""
	@echo "  Code Quality:"
	@echo "    make lint             - Run linters"
	@echo "    make fmt              - Format code"
	@echo "    make vet              - Run go vet"
	@echo ""
	@echo "  Tools:"
	@echo "    make deps             - Download dependencies"
	@echo "    make tidy             - Tidy go.mod"
	@echo "    make grpc-test        - Test with grpcurl"
	@echo ""

# ============================================================================
# Variables
# ============================================================================

APP_NAME = auth-service
BINARY_NAME = auth-service
DOCKER_IMAGE = ikampus/$(APP_NAME)
DOCKER_TAG = latest

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOFMT = $(GOCMD) fmt
GOVET = $(GOCMD) vet

# Directories
CMD_DIR = ./cmd/server
INTERNAL_DIR = ./internal
TEST_DIR = ./test

# ============================================================================
# Development
# ============================================================================

.PHONY: run
run:
	@echo "Running AuthService..."
	$(GOCMD) run $(CMD_DIR)/main.go

.PHONY: dev
dev:
	@echo "Running with live reload (requires air)..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Error: air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

# ============================================================================
# Build
# ============================================================================

.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) \
		-ldflags="-w -s" \
		-o bin/$(BINARY_NAME) \
		$(CMD_DIR)/main.go
	@echo "Binary created at: bin/$(BINARY_NAME)"

.PHONY: build-local
build-local:
	@echo "Building $(BINARY_NAME) for local OS..."
	$(GOBUILD) -o bin/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "Binary created at: bin/$(BINARY_NAME)"

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f $(BINARY_NAME)
	$(GOCMD) clean

# ============================================================================
# Testing
# ============================================================================

.PHONY: test
test:
	@echo "Running all tests..."
	$(GOTEST) -v -race ./...

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -race ./internal/...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -race ./test/integration/...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: test-bench
test-bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# ============================================================================
# Docker
# ============================================================================

.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-up
docker-up:
	@echo "Starting all services with docker-compose..."
	docker-compose up -d
	@echo "Services started. Use 'make docker-logs' to view logs"

.PHONY: docker-up-build
docker-up-build:
	@echo "Building and starting all services..."
	docker-compose up -d --build

.PHONY: docker-down
docker-down:
	@echo "Stopping all services..."
	docker-compose down

.PHONY: docker-down-volumes
docker-down-volumes:
	@echo "Stopping all services and removing volumes..."
	docker-compose down -v

.PHONY: docker-logs
docker-logs:
	docker-compose logs -f auth-service

.PHONY: docker-restart
docker-restart:
	@echo "Restarting auth-service..."
	docker-compose restart auth-service

.PHONY: docker-shell
docker-shell:
	docker-compose exec auth-service sh

# ============================================================================
# Database
# ============================================================================

.PHONY: db-init
db-init:
	@echo "Initializing database..."
	cd ../../database && ./init_database.sh ikampus ikampus_user changeme

.PHONY: db-reset
db-reset:
	@echo "Resetting database..."
	docker-compose exec postgres psql -U ikampus_user -d ikampus -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) db-init

# ============================================================================
# Code Quality
# ============================================================================

.PHONY: lint
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# ============================================================================
# Dependencies
# ============================================================================

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

.PHONY: tidy
tidy:
	@echo "Tidying go.mod..."
	$(GOMOD) tidy

.PHONY: vendor
vendor:
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor

# ============================================================================
# Protocol Buffers
# ============================================================================

.PHONY: proto
proto:
	@echo "Generating proto files..."
	cd ../../scripts && ./generate-go.sh

# ============================================================================
# gRPC Testing
# ============================================================================

.PHONY: grpc-test
grpc-test:
	@echo "Testing gRPC endpoints with grpcurl..."
	@echo ""
	@echo "1. List services:"
	grpcurl -plaintext localhost:9000 list
	@echo ""
	@echo "2. Health check:"
	grpcurl -plaintext localhost:9000 grpc.health.v1.Health/Check
	@echo ""
	@echo "For more tests, see README.md"

.PHONY: grpc-signup
grpc-signup:
	@echo "Testing SignUp..."
	grpcurl -plaintext -d '{ \
		"email": "student@ox.ac.uk", \
		"password": "Password123", \
		"username": "john_doe", \
		"display_name": "John Doe", \
		"university_id": "uni_oxford", \
		"course": "Computer Science", \
		"year_of_study": 2, \
		"graduation_year": 2026, \
		"interests": ["coding", "startups"] \
	}' localhost:9000 ikampus.auth.AuthService/SignUp

.PHONY: grpc-login
grpc-login:
	@echo "Testing Login..."
	grpcurl -plaintext -d '{ \
		"email": "student@ox.ac.uk", \
		"password": "Password123" \
	}' localhost:9000 ikampus.auth.AuthService/Login

# ============================================================================
# Utility
# ============================================================================

.PHONY: version
version:
	@echo "Go version:"
	@$(GOCMD) version
	@echo ""
	@echo "Service version: 1.0.0"

.PHONY: status
status:
	@echo "Service status:"
	@docker-compose ps

.PHONY: info
info:
	@echo "AuthService Information:"
	@echo "  App Name:      $(APP_NAME)"
	@echo "  Binary Name:   $(BINARY_NAME)"
	@echo "  Docker Image:  $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "  gRPC Port:     9000"
	@echo "  HTTP Port:     8080"
