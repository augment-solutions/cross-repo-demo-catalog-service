.PHONY: build run test test-unit test-integration lint clean docker-build docker-run

# Variables
APP_NAME := catalog-service
DOCKER_IMAGE := ecomm-stack/$(APP_NAME)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -ldflags="-X main.version=$(VERSION)" -o bin/$(APP_NAME) ./cmd/main.go

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	go run ./cmd/main.go

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v -race -cover ./tests/unit/...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -race -cover ./tests/integration/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint the code
lint:
	@echo "Linting..."
	golangci-lint run ./...

# Format the code
fmt:
	@echo "Formatting..."
	go fmt ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 \
		-e DATABASE_URL=postgres://postgres:postgres@host.docker.internal:5432/catalog \
		-e REDIS_URL=redis://host.docker.internal:6379 \
		-e OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317 \
		$(DOCKER_IMAGE):latest

# Docker compose up
compose-up:
	docker-compose up -d

# Docker compose down
compose-down:
	docker-compose down

# Generate mocks
mocks:
	@echo "Generating mocks..."
	mockgen -source=internal/domain/repositories/interfaces.go -destination=tests/mocks/repositories.go

# Run database migrations
migrate-up:
	@echo "Running migrations..."
	go run ./cmd/migrate up

# Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	go run ./cmd/migrate down

# Help
help:
	@echo "Available targets:"
	@echo "  build            - Build the application"
	@echo "  run              - Run the application"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  lint             - Lint the code"
	@echo "  fmt              - Format the code"
	@echo "  tidy             - Tidy dependencies"
	@echo "  clean            - Clean build artifacts"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  compose-up       - Start Docker Compose services"
	@echo "  compose-down     - Stop Docker Compose services"
	@echo "  mocks            - Generate mocks"
	@echo "  migrate-up       - Run database migrations"
	@echo "  migrate-down     - Rollback database migrations"

