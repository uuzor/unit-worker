# ============================================================================
# Multi-stage Dockerfile for AuthService
# ============================================================================
# Stage 1: Build the Go binary
# Stage 2: Create minimal runtime image
#
# Build:    docker build -t ikampus/auth-service:latest .
# Run:      docker run -p 9000:9000 ikampus/auth-service:latest
# ============================================================================

# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o auth-service \
    ./cmd/server

# Stage 2: Runtime
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from builder
COPY --from=builder /build/auth-service /auth-service

# Expose ports
EXPOSE 8080 9000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/auth-service", "healthcheck"] || exit 1

# Run as non-root user
USER 65534:65534

# Set entrypoint
ENTRYPOINT ["/auth-service"]
