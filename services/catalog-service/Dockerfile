# syntax=docker/dockerfile:1.4
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY services/catalog-service/go.mod services/catalog-service/go.sum ./

# Download dependencies with BuildKit cache mount
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download

# Copy source code
COPY services/catalog-service/cmd ./cmd
COPY services/catalog-service/internal ./internal

# Build the application with BuildKit cache mount
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /catalog-service ./cmd/main.go

# Build the seed command
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /catalog-seed ./cmd/seed/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS and stress-ng for chaos engineering
RUN apk add --no-cache ca-certificates tzdata stress-ng iproute2

# Copy binaries from builder
COPY --from=builder /catalog-service /app/catalog-service
COPY --from=builder /catalog-seed /app/catalog-seed

# Copy seed data
COPY services/catalog-service/data/seed /app/data/seed

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose ports
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/health || exit 1

# Run the application
ENTRYPOINT ["/app/catalog-service"]

