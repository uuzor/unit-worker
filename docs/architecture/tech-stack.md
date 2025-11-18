# Technology Stack for iKampus

**Microservices-based university platform**

---

## Overview

This document outlines the complete technology stack for the iKampus platform, based on Google Cloud's microservices best practices and modern cloud-native architecture.

---

## Programming Languages

### Go (Primary Backend Language)
**Use for:** Core services, API Gateway, high-throughput services

**Services:**
- GatewayService
- UserService
- PostService
- FeedService
- MessagingService
- CommunityService
- StartupService
- SearchService
- MediaService
- AnalyticsService

**Why Go:**
- ✅ Excellent performance and low latency
- ✅ Low memory footprint
- ✅ Built-in concurrency (goroutines)
- ✅ Fast compilation
- ✅ Strong standard library
- ✅ Native gRPC support
- ✅ Easy deployment (single binary)

**Key Libraries:**
```go
// gRPC and Protocol Buffers
google.golang.org/grpc
google.golang.org/protobuf

// HTTP framework
github.com/gin-gonic/gin
github.com/labstack/echo/v4

// Database
github.com/lib/pq                    // PostgreSQL
github.com/go-redis/redis/v8         // Redis
gorm.io/gorm                         // ORM

// Observability
go.opentelemetry.io/otel
github.com/sirupsen/logrus

// Authentication
github.com/golang-jwt/jwt/v5
golang.org/x/crypto/bcrypt
```

---

### Python (AI & ML Services)
**Use for:** AI/ML features, data processing, content moderation

**Services:**
- AIAssistantService
- RecommendationService
- ModerationService

**Why Python:**
- ✅ Rich AI/ML ecosystem
- ✅ Easy integration with OpenAI, Anthropic
- ✅ Fast prototyping
- ✅ Strong data processing libraries
- ✅ Extensive NLP support

**Key Libraries:**
```python
# gRPC
grpcio
grpcio-tools

# AI/ML
openai
anthropic
langchain
transformers
torch
scikit-learn

# Data processing
pandas
numpy

# Web framework
fastapi
uvicorn

# Database
psycopg2
redis

# Observability
opentelemetry-api
opentelemetry-sdk
```

---

### Node.js (Real-time Services)
**Use for:** Real-time features, high concurrency I/O

**Services:**
- NotificationService (alternative to Go)
- Real-time WebSocket handlers

**Why Node.js:**
- ✅ Non-blocking I/O
- ✅ Excellent WebSocket support
- ✅ Large ecosystem (npm)
- ✅ Easy async programming
- ✅ Good for I/O-bound tasks

**Key Libraries:**
```javascript
// gRPC
@grpc/grpc-js
@grpc/proto-loader

// Web framework
express
fastify
socket.io              // WebSocket

// Database
pg                     // PostgreSQL
ioredis                // Redis

// Authentication
jsonwebtoken
bcrypt

// Observability
@opentelemetry/sdk-node
@opentelemetry/auto-instrumentations-node
winston                // Logging

// Push notifications
firebase-admin
```

---

## Communication & APIs

### gRPC + Protocol Buffers
**Purpose:** Internal service-to-service communication

**Why gRPC:**
- ✅ Type-safe contracts
- ✅ Efficient binary serialization
- ✅ Built-in code generation
- ✅ Streaming support
- ✅ Language-agnostic
- ✅ HTTP/2 based

**Example Proto Definition:**
```protobuf
syntax = "proto3";

package ikampus.user;

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc UpdateProfile(UpdateProfileRequest) returns (User);
  rpc FollowUser(FollowUserRequest) returns (Empty);
}

message User {
  string id = 1;
  string username = 2;
  string email = 3;
  string course = 4;
  int32 year = 5;
  string university_id = 6;
  string avatar_url = 7;
  repeated string interests = 8;
  int64 created_at = 9;
}
```

---

### REST API (External)
**Purpose:** Mobile/web client communication

**Framework:** Go (Gin/Echo)

**API Design:**
```
GET    /api/v1/users/:id
POST   /api/v1/posts
GET    /api/v1/feed
POST   /api/v1/auth/login
GET    /api/v1/startups
```

**Authentication:** JWT (Bearer tokens)

---

### WebSocket
**Purpose:** Real-time features (chat, notifications)

**Protocol:** WebSocket over HTTPS

**Events:**
```javascript
// Client → Server
{
  "type": "message.send",
  "data": {
    "conversation_id": "conv_123",
    "content": "Hello!"
  }
}

// Server → Client
{
  "type": "message.received",
  "data": {
    "message_id": "msg_456",
    "sender_id": "user_789",
    "content": "Hello!",
    "timestamp": 1700000000
  }
}
```

---

## Databases

### PostgreSQL (Primary Database)
**Version:** 14+
**Purpose:** Primary data store for all persistent data

**Cloud Options:**
- Google Cloud SQL
- AWS RDS
- Azure Database for PostgreSQL
- Self-hosted on Kubernetes (development)

**Schema Overview:**
```sql
-- Users & Authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- User Profiles
CREATE TABLE profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    username VARCHAR(50) UNIQUE NOT NULL,
    course VARCHAR(100),
    year INT,
    university_id UUID REFERENCES universities(id),
    bio TEXT,
    avatar_url VARCHAR(500),
    interests TEXT[],
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Posts
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    community_id UUID REFERENCES communities(id),
    content TEXT NOT NULL,
    image_urls TEXT[],
    is_anonymous BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_community_id ON posts(community_id);

-- Comments
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    content TEXT NOT NULL,
    parent_comment_id UUID REFERENCES comments(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Reactions
CREATE TABLE reactions (
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    reaction_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

-- Startups
CREATE TABLE startups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    stage VARCHAR(50),
    industry VARCHAR(100),
    demo_url VARCHAR(500),
    founded_at DATE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Extensions:**
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- Full-text search
```

---

### Redis (Cache & Session Store)
**Version:** 7+
**Purpose:** Caching, session management, rate limiting, real-time counters

**Cloud Options:**
- Google Memorystore
- AWS ElastiCache
- Azure Cache for Redis
- Self-hosted on Kubernetes

**Data Structures:**
```redis
# Sessions (Hash)
HSET session:abc123 user_id "user_456" email "john@uni.ac.uk" expires_at 1700000000
EXPIRE session:abc123 86400

# Feed cache (List)
LPUSH feed:user_123 post_1 post_2 post_3
EXPIRE feed:user_123 300

# Counters (String)
INCR post_likes:post_456
GET post_likes:post_456

# Rate limiting (String)
INCR rate_limit:user_123:api
EXPIRE rate_limit:user_123:api 60

# Presence (Set)
SADD online_users user_123
EXPIRE online_users 300

# Trending posts (Sorted Set)
ZADD trending_posts 150 post_1 120 post_2 95 post_3
ZREVRANGE trending_posts 0 9
```

**Redis Cluster:**
- 3 master nodes
- 3 replica nodes
- Automatic failover

---

### Elasticsearch (Search Engine)
**Version:** 8+
**Purpose:** Full-text search across users, posts, communities, startups

**Cloud Options:**
- Elastic Cloud
- AWS OpenSearch
- Self-hosted on Kubernetes

**Indices:**
```json
// Users index
{
  "mappings": {
    "properties": {
      "user_id": { "type": "keyword" },
      "username": { "type": "text", "analyzer": "standard" },
      "bio": { "type": "text", "analyzer": "english" },
      "course": { "type": "keyword" },
      "university_id": { "type": "keyword" },
      "interests": { "type": "keyword" }
    }
  }
}

// Posts index
{
  "mappings": {
    "properties": {
      "post_id": { "type": "keyword" },
      "content": { "type": "text", "analyzer": "english" },
      "user_id": { "type": "keyword" },
      "community_id": { "type": "keyword" },
      "hashtags": { "type": "keyword" },
      "created_at": { "type": "date" }
    }
  }
}
```

---

## Message Queue & Event Streaming

### Google Cloud Pub/Sub / AWS SQS
**Purpose:** Async event processing, decoupling services

**Topics:**
- `post.created` - New post notifications
- `user.followed` - Follower notifications
- `message.sent` - Message delivery
- `startup.created` - Startup notifications
- `analytics.event` - Analytics events

**Example Flow:**
```
PostService creates post
    ↓
Publishes to post.created topic
    ↓
    ├→ NotificationService (notify followers)
    ├→ FeedService (invalidate caches)
    ├→ SearchService (index post)
    └→ AnalyticsService (track event)
```

---

## Storage

### Object Storage (S3 / Google Cloud Storage)
**Purpose:** Images, videos, documents, attachments

**Buckets:**
- `ikampus-avatars` - Profile pictures
- `ikampus-posts` - Post images/videos
- `ikampus-documents` - Student verification docs
- `ikampus-pitches` - Startup pitch decks

**CDN:** CloudFront / Cloud CDN for faster delivery

**Upload Flow:**
```
Client → MediaService (generates signed URL)
Client → S3 (direct upload)
Client → PostService (saves URL to database)
```

---

## Container Orchestration

### Kubernetes
**Version:** 1.28+
**Managed Options:**
- Google Kubernetes Engine (GKE)
- Amazon Elastic Kubernetes Service (EKS)
- Azure Kubernetes Service (AKS)

**Resources:**
- **Nodes:** Auto-scaling node pools
- **Ingress:** NGINX Ingress Controller / Istio Gateway
- **Service Mesh:** Istio
- **Cert Manager:** Let's Encrypt for TLS

**Example Deployment:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: userservice
spec:
  replicas: 3
  selector:
    matchLabels:
      app: userservice
  template:
    metadata:
      labels:
        app: userservice
        version: v1
    spec:
      serviceAccountName: userservice
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      containers:
      - name: server
        image: gcr.io/ikampus/userservice:v1.2.3
        ports:
        - containerPort: 9001
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: userservice-secrets
              key: database-url
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 128Mi
        livenessProbe:
          grpc:
            port: 9001
        readinessProbe:
          grpc:
            port: 9001
---
apiVersion: v1
kind: Service
metadata:
  name: userservice
spec:
  selector:
    app: userservice
  ports:
  - port: 9001
    targetPort: 9001
```

---

## Service Mesh

### Istio
**Version:** 1.20+
**Purpose:** Traffic management, security, observability

**Features:**
- Mutual TLS between services
- Traffic splitting (canary deployments)
- Circuit breaking
- Retry policies
- Distributed tracing

**Example VirtualService:**
```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: userservice
spec:
  hosts:
  - userservice
  http:
  - route:
    - destination:
        host: userservice
        subset: v1
      weight: 90
    - destination:
        host: userservice
        subset: v2
      weight: 10
    timeout: 5s
    retries:
      attempts: 3
      perTryTimeout: 2s
```

---

## Observability

### Distributed Tracing
**Tool:** OpenTelemetry + Jaeger
**Purpose:** Trace requests across services

**Example Trace:**
```
API Gateway: GET /api/v1/feed [200ms]
├─ AuthService: ValidateToken [10ms]
├─ FeedService: GetHomeFeed [150ms]
│  ├─ UserService: GetFollowing [30ms]
│  ├─ PostService: GetPostsByIds [80ms]
│  └─ Redis: Get feed:user_123 [5ms]
└─ UserService: GetUsersByIds [40ms]
```

---

### Metrics
**Tool:** Prometheus + Grafana
**Purpose:** Monitor service health and performance

**Key Metrics:**
```prometheus
# Request rate
rate(grpc_requests_total[5m])

# Error rate
rate(grpc_requests_total{status="error"}[5m])

# Latency (p95)
histogram_quantile(0.95, grpc_request_duration_seconds_bucket)

# Active connections
grpc_active_connections

# Database connections
database_connections_active
database_connections_idle
```

---

### Logging
**Tool:** Fluentd → Elasticsearch → Kibana
**Format:** Structured JSON logs

**Log Entry:**
```json
{
  "timestamp": "2025-11-18T22:30:45.123Z",
  "severity": "info",
  "service": "postservice",
  "trace_id": "a1b2c3d4e5f6",
  "span_id": "1234567890ab",
  "message": "post created successfully",
  "user_id": "user_123",
  "post_id": "post_456",
  "duration_ms": 45
}
```

---

## CI/CD

### GitHub Actions
**Purpose:** Build, test, and deploy

**Workflow:**
```yaml
name: Deploy UserService

on:
  push:
    branches: [main]
    paths:
      - 'src/userservice/**'

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Build Docker image
      run: |
        docker build -t gcr.io/ikampus/userservice:${{ github.sha }} \
          src/userservice

    - name: Push to registry
      run: |
        docker push gcr.io/ikampus/userservice:${{ github.sha }}

    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/userservice \
          server=gcr.io/ikampus/userservice:${{ github.sha }}
```

---

### ArgoCD
**Purpose:** GitOps continuous deployment

**Features:**
- Automatic sync from Git
- Rollback capabilities
- Multi-environment support

---

## Security

### Authentication & Authorization

**JWT (JSON Web Tokens):**
```json
{
  "sub": "user_123",
  "email": "john@uni.ac.uk",
  "role": "student",
  "university_id": "uni_456",
  "exp": 1700000000,
  "iat": 1699900000
}
```

**OAuth 2.0:** For third-party integrations

**RBAC (Role-Based Access Control):**
- Student
- Alumni
- Moderator
- Admin

---

### Secrets Management

**Google Secret Manager / AWS Secrets Manager**

**Never in code/config:**
- Database passwords
- API keys
- JWT signing keys
- Third-party credentials

**Kubernetes Secrets:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: userservice-secrets
type: Opaque
data:
  database-url: <base64-encoded>
  jwt-secret: <base64-encoded>
```

---

## Development Tools

### Local Development

**Docker Compose:**
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: ikampus
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  elasticsearch:
    image: elasticsearch:8.10.0
    environment:
      discovery.type: single-node
    ports:
      - "9200:9200"
```

---

### Testing

**Unit Tests:**
- Go: `testing` package
- Python: `pytest`
- Node.js: `jest`

**Integration Tests:**
- gRPC: `grpc-test` library
- Database: Testcontainers

**Load Tests:**
- Locust (Python)
- k6 (Go-based)

---

## Cost Estimates

### Development (3 engineers)
- GKE cluster: $200/month
- Cloud SQL: $50/month
- Redis: $30/month
- **Total: ~$300/month**

### Production Launch (1,000 users)
- GKE cluster: $800/month
- Cloud SQL: $200/month
- Redis: $100/month
- Storage: $50/month
- CDN: $50/month
- **Total: ~$1,200/month**

### Production Growth (10,000 users)
- GKE cluster: $2,000/month
- Cloud SQL: $500/month
- Redis: $200/month
- Storage: $200/month
- CDN: $150/month
- **Total: ~$3,050/month**

---

**Document Version:** 1.0
**Last Updated:** 2025-11-18
