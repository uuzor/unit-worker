# AuthService Setup Guide

Complete guide for setting up and running the AuthService locally and in production.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Protocol Buffer Code Generation](#protocol-buffer-code-generation)
4. [Local Development](#local-development)
5. [Docker Development](#docker-development)
6. [Testing](#testing)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

- **Go 1.21+**: [Download](https://go.dev/dl/)
- **Docker & Docker Compose**: [Download](https://docs.docker.com/get-docker/)
- **PostgreSQL 14+**: Required for local development (or use Docker)
- **Protocol Buffer Compiler (protoc)**: For generating gRPC code
- **grpcurl**: For testing gRPC endpoints

### Installing Protocol Buffer Compiler

#### Linux (Ubuntu/Debian)
```bash
# Install protoc
sudo apt-get update
sudo apt-get install -y protobuf-compiler

# Verify installation
protoc --version  # Should show libprotoc 3.x.x or higher
```

#### macOS
```bash
brew install protobuf
protoc --version
```

#### Manual Installation (All Platforms)
```bash
# Download latest release
VERSION=25.1
cd /tmp
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${VERSION}/protoc-${VERSION}-linux-x86_64.zip

# Extract to /usr/local
unzip protoc-${VERSION}-linux-x86_64.zip -d /usr/local

# Verify
protoc --version
```

### Installing Go Plugins for protoc

```bash
# Install protoc-gen-go (generates Go structs)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Install protoc-gen-go-grpc (generates gRPC service code)
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add Go bin to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH="$PATH:$(go env GOPATH)/bin"

# Verify
which protoc-gen-go
which protoc-gen-go-grpc
```

### Installing grpcurl

```bash
# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# macOS
brew install grpcurl

# Verify
grpcurl --version
```

## Initial Setup

### 1. Clone Repository

```bash
cd /home/user/unit-worker
```

### 2. Install Go Dependencies

```bash
cd services/auth
go mod download
go mod tidy
```

### 3. Set Up Environment Variables

Create `.env` file:

```bash
cp .env.example .env  # If example exists, or create manually
```

Edit `.env`:

```env
# Server Configuration
SERVER_PORT=8080
GRPC_PORT=9000
ENVIRONMENT=development
SHUTDOWN_TIMEOUT=10s

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=ikampus_user
DB_PASSWORD=changeme
DB_NAME=ikampus
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# JWT Configuration
JWT_ACCESS_SECRET=your-super-secret-access-key-change-in-production
JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=720h  # 30 days
JWT_ISSUER=ikampus-auth

# Email Configuration (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@ikampus.com
FROM_NAME=iKampus
VERIFICATION_EXPIRY=24h

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Protocol Buffer Code Generation

### Generate Go Code from Protos

```bash
# From repository root
cd /home/user/unit-worker

# Generate all proto files
./scripts/generate-go.sh

# Or generate only auth service
cd protos
protoc --go_out=../services/auth/pb \
       --go-grpc_out=../services/auth/pb \
       --go_opt=paths=source_relative \
       --go-grpc_opt=paths=source_relative \
       -I. \
       common/common.proto \
       auth/auth.proto
```

Expected output:
```
services/auth/pb/
├── common/
│   └── common.pb.go
└── auth/
    ├── auth.pb.go
    └── auth_grpc.pb.go
```

### Update Import Paths

After generation, update import paths in your Go files:

```go
// Change from:
import pb "github.com/ikampus/protos/auth"

// To:
import pb "github.com/ikampus/auth-service/pb/auth"
```

## Local Development

### 1. Initialize Database

```bash
# From repository root
cd database
./init_database.sh ikampus ikampus_user changeme

# Or manually with psql
createdb ikampus
psql -U ikampus_user -d ikampus -f schemas/00_extensions.sql
psql -U ikampus_user -d ikampus -f schemas/01_auth_users.sql
# ... repeat for all schema files
```

### 2. Run the Service

```bash
cd services/auth

# Load environment variables
export $(cat .env | xargs)

# Run the service
go run cmd/server/main.go
```

Expected output:
```json
{"level":"info","msg":"Starting AuthService...","time":"2024-01-15T10:00:00Z"}
{"level":"info","msg":"Configuration loaded","environment":"development","grpc_port":9000,"http_port":8080}
{"level":"info","msg":"Database connection established"}
{"level":"info","msg":"Starting gRPC server","port":9000}
{"level":"info","msg":"Starting HTTP health check server","port":8080}
{"level":"info","msg":"AuthService is running"}
```

### 3. Verify Service is Running

```bash
# Check health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","database":"up"}

# List gRPC services
grpcurl -plaintext localhost:9000 list

# Expected output:
# grpc.health.v1.Health
# grpc.reflection.v1alpha.ServerReflection
# ikampus.auth.AuthService
```

## Docker Development

### Quick Start with Docker Compose

```bash
cd services/auth

# Start all services (PostgreSQL, Redis, AuthService)
make docker-up

# Or manually:
docker-compose up -d

# View logs
make docker-logs
# Or:
docker-compose logs -f auth-service

# Stop all services
make docker-down
```

### Build Docker Image

```bash
# Build the image
make docker-build

# Or manually:
docker build -t ikampus/auth-service:latest .

# Run the container
docker run -p 8080:8080 -p 9000:9000 \
  -e DB_HOST=postgres \
  -e DB_USER=ikampus_user \
  -e DB_PASSWORD=changeme \
  ikampus/auth-service:latest
```

## Testing

### Unit Tests

```bash
cd services/auth

# Run all tests
make test
# Or:
go test -v ./...

# Run specific package tests
go test -v ./internal/utils/...

# Run with coverage
make test-coverage
# Or:
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests with grpcurl

#### 1. Sign Up

```bash
grpcurl -plaintext -d '{
  "email": "john.doe@ox.ac.uk",
  "password": "SecurePass123",
  "username": "john_doe",
  "display_name": "John Doe",
  "university_id": "uni_oxford",
  "course": "Computer Science",
  "year_of_study": 2,
  "graduation_year": 2026,
  "interests": ["coding", "startups", "AI"]
}' localhost:9000 ikampus.auth.AuthService/SignUp
```

Expected response:
```json
{
  "userId": "user_a1b2c3d4",
  "email": "john.doe@ox.ac.uk",
  "username": "john_doe",
  "emailVerified": false,
  "message": "Account created successfully. Please check your email to verify your account."
}
```

#### 2. Verify Email

```bash
# Get verification token from database or logs
grpcurl -plaintext -d '{
  "token": "your-verification-token-here"
}' localhost:9000 ikampus.auth.AuthService/VerifyEmail
```

Expected response:
```json
{
  "success": true,
  "message": "Email verified successfully",
  "tokens": {
    "accessToken": "eyJhbGc...",
    "refreshToken": "eyJhbGc...",
    "accessTokenExpiresAt": "2024-01-15T10:15:00Z",
    "refreshTokenExpiresAt": "2024-02-14T10:00:00Z"
  }
}
```

#### 3. Login

```bash
grpcurl -plaintext -d '{
  "email": "john.doe@ox.ac.uk",
  "password": "SecurePass123",
  "device_id": "web-12345",
  "device_name": "Chrome Browser",
  "device_type": "web"
}' localhost:9000 ikampus.auth.AuthService/Login
```

#### 4. Refresh Token

```bash
grpcurl -plaintext -d '{
  "refresh_token": "your-refresh-token-here"
}' localhost:9000 ikampus.auth.AuthService/RefreshToken
```

#### 5. Get Active Sessions

```bash
grpcurl -plaintext -d '{
  "user_id": "user_a1b2c3d4"
}' localhost:9000 ikampus.auth.AuthService/GetActiveSessions
```

#### 6. Logout

```bash
grpcurl -plaintext -d '{
  "refresh_token": "your-refresh-token-here"
}' localhost:9000 ikampus.auth.AuthService/Logout
```

### Using Makefile Commands

```bash
# Quick test commands
make grpc-test         # List services and health check
make grpc-signup       # Test signup endpoint
make grpc-login        # Test login endpoint
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

**Error:** `Failed to connect to database`

**Solutions:**
```bash
# Check PostgreSQL is running
sudo systemctl status postgresql  # Linux
brew services list                 # macOS

# Check connection
psql -U ikampus_user -d ikampus -h localhost

# Verify credentials in .env file
# Check DB_HOST, DB_PORT, DB_USER, DB_PASSWORD
```

#### 2. Port Already in Use

**Error:** `bind: address already in use`

**Solutions:**
```bash
# Find process using the port
lsof -i :9000  # For gRPC port
lsof -i :8080  # For HTTP port

# Kill the process
kill -9 <PID>

# Or change ports in .env
```

#### 3. Proto Generation Failed

**Error:** `protoc: command not found`

**Solution:**
```bash
# Install protoc (see Prerequisites section)
which protoc
protoc --version

# Ensure Go plugins are installed
which protoc-gen-go
which protoc-gen-go-grpc

# Check PATH includes Go bin directory
echo $PATH | grep go/bin
```

#### 4. Import Errors After Proto Generation

**Error:** `cannot find package "github.com/ikampus/protos/auth"`

**Solutions:**
```bash
# Update import paths in generated code
# Change module path in go.mod if needed

# Run go mod tidy
cd services/auth
go mod tidy

# Clear Go module cache
go clean -modcache
```

#### 5. Tests Failing

**Error:** Various test failures

**Solutions:**
```bash
# Clean build cache
go clean -testcache

# Run tests with verbose output
go test -v ./...

# Check for missing dependencies
go mod verify
go mod download
```

### Logs and Debugging

#### View Service Logs

```bash
# Docker logs
docker-compose logs -f auth-service

# Local logs (stdout)
go run cmd/server/main.go | jq  # Pretty print JSON logs
```

#### Enable Debug Logging

Add to `.env`:
```env
LOG_LEVEL=debug
```

#### Database Queries

```bash
# Connect to database
psql -U ikampus_user -d ikampus

# View users
SELECT id, email, email_verified, role, status FROM users;

# View sessions
SELECT id, user_id, device_type, created_at, last_used_at
FROM refresh_tokens
WHERE revoked = false;
```

### Performance Tuning

#### Database Connection Pool

Adjust in `.env`:
```env
DB_MAX_OPEN_CONNS=25   # Max open connections
DB_MAX_IDLE_CONNS=5    # Max idle connections
DB_CONN_MAX_LIFETIME=5m # Connection lifetime
```

#### JWT Token Expiry

```env
JWT_ACCESS_EXPIRY=15m    # Shorter = more secure, more requests
JWT_REFRESH_EXPIRY=720h  # 30 days
```

## Next Steps

1. **Set up monitoring**: Prometheus metrics, Grafana dashboards
2. **Configure CI/CD**: GitHub Actions for automated testing and deployment
3. **Add rate limiting**: Protect against brute force attacks
4. **Implement email service**: Send verification and password reset emails
5. **Add observability**: Distributed tracing with OpenTelemetry
6. **Deploy to production**: Kubernetes deployment (see DEPLOYMENT.md)

## Additional Resources

- [gRPC Go Quickstart](https://grpc.io/docs/languages/go/quickstart/)
- [Protocol Buffers Tutorial](https://protobuf.dev/getting-started/gotutorial/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [Go JWT Library](https://github.com/golang-jwt/jwt)
