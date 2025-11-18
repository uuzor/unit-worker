# Microservices Architecture Analysis

**Based on:** Google Cloud's Online Boutique Demo
**Repository:** https://github.com/GoogleCloudPlatform/microservices-demo
**Date:** 2025-11-18

---

## Executive Summary

This document analyzes Google's production-grade microservices reference architecture to extract best practices for the iKampus platform. The Online Boutique demo showcases 11 polyglot microservices communicating via gRPC on Kubernetes, demonstrating enterprise-grade patterns for security, observability, and resilience.

### Key Learnings for iKampus

1. **gRPC for internal communication** - Type-safe, efficient, language-agnostic
2. **Polyglot architecture** - Use the right tool for each job
3. **Security by default** - Non-root containers, read-only filesystems, minimal privileges
4. **Observability-first** - OpenTelemetry, structured logging, distributed tracing
5. **Orchestration patterns** - Saga pattern for complex workflows
6. **Kubernetes-native** - Health checks, graceful shutdown, resource limits

---

## 1. Architecture Overview

### 1.1 Communication Patterns

**gRPC as Primary Protocol**

All 11 microservices communicate exclusively via gRPC:

```protobuf
// Example service definition (demo.proto)
service CartService {
    rpc AddItem(AddItemRequest) returns (Empty) {}
    rpc GetCart(GetCartRequest) returns (Cart) {}
    rpc EmptyCart(EmptyCartRequest) returns (Empty) {}
}

service ProductCatalogService {
    rpc ListProducts(Empty) returns (ListProductsResponse) {}
    rpc GetProduct(GetProductRequest) returns (Product) {}
    rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse) {}
}
```

**Benefits for iKampus:**
- Strong typing prevents integration bugs
- Efficient binary protocol (lower latency than JSON)
- Automatic client code generation for all languages
- Built-in streaming support for real-time features

**Service Discovery via Kubernetes DNS**

Services are addressed by their Kubernetes service names:

```yaml
env:
- name: PRODUCT_CATALOG_SERVICE_ADDR
  value: "productcatalogservice:3550"
- name: CART_SERVICE_ADDR
  value: "cartservice:7070"
- name: CURRENCY_SERVICE_ADDR
  value: "currencyservice:7000"
```

### 1.2 Service Architecture

| Service | Language | Port | Responsibility |
|---------|----------|------|----------------|
| **frontend** | Go | 8080 | Web UI, HTTP gateway |
| **cartservice** | C# | 7070 | Shopping cart CRUD |
| **productcatalogservice** | Go | 3550 | Product data |
| **currencyservice** | Node.js | 7000 | Currency conversion |
| **paymentservice** | Node.js | 50051 | Payment processing |
| **shippingservice** | Go | 50051 | Shipping calculations |
| **emailservice** | Python | 5000 | Email notifications |
| **checkoutservice** | Go | 5050 | **Orchestration service** |
| **recommendationservice** | Python | 8080 | Product recommendations |
| **adservice** | Java | 9555 | Contextual ads |
| **loadgenerator** | Python | - | Traffic simulation |

### 1.3 Data Persistence Strategy

**Redis for Session State:**
- CartService stores shopping carts in Redis
- Data serialized as Protocol Buffer binary
- Multiple backend support via strategy pattern:
  - RedisCartStore (default)
  - SpannerCartStore (Google Cloud Spanner)
  - AlloyDBCartStore (AlloyDB for PostgreSQL)

**Stateless Services:**
- All other services are stateless
- Product catalog loaded from JSON file (in-memory)
- No shared databases between services

---

## 2. Proposed iKampus Microservices Architecture

### 2.1 Service Decomposition

```
iKampus Platform
│
├── Authentication & User Services
│   ├── AuthService (Node.js/Go) - JWT, OAuth, verification
│   ├── UserService (Go) - User profiles, preferences
│   └── UniversityService (Go) - University data, verification
│
├── Social Features
│   ├── FeedService (Go) - Feed aggregation, ranking
│   ├── PostService (Go) - Posts, comments, reactions
│   ├── CommunityService (Go) - Groups, memberships
│   ├── MessagingService (Go) - Direct messages, chat
│   └── EventService (Go) - Campus events, RSVPs
│
├── AI & Intelligence
│   ├── AIAssistantService (Python) - Academic support AI
│   ├── RecommendationService (Python) - Content recommendations
│   └── ModerationService (Python) - Content moderation
│
├── Startup Hub
│   ├── StartupService (Go) - Startup profiles, teams
│   ├── MatchmakingService (Python) - Co-founder matching
│   └── MentorshipService (Go) - Alumni connections
│
├── Infrastructure Services
│   ├── NotificationService (Node.js) - Push, email, SMS
│   ├── SearchService (Go) - Full-text search (Elasticsearch)
│   ├── AnalyticsService (Go) - Event tracking
│   └── MediaService (Go) - Image/video upload, processing
│
└── API Gateway
    └── GatewayService (Go) - Mobile/web API, auth, rate limiting
```

### 2.2 Service Communication

**Internal (Service-to-Service):**
- gRPC for synchronous calls
- Protocol Buffers for data contracts
- Pub/Sub for asynchronous events (new post, new comment)

**External (Client-to-Service):**
- REST/GraphQL via API Gateway
- WebSocket for real-time features (chat, notifications)
- Mobile SDKs (iOS, Android)

### 2.3 Language Selection Rationale

| Language | Use Cases | Reasoning |
|----------|-----------|-----------|
| **Go** | Core services, gateway, high-throughput | Fast, low memory, excellent concurrency |
| **Python** | AI/ML services, moderation | Rich ML ecosystem, fast development |
| **Node.js** | Real-time services, notifications | Non-blocking I/O, WebSocket support |
| **Rust** | (Future) Media processing | High performance, memory safety |

---

## 3. Kubernetes Deployment Patterns

### 3.1 Security Best Practices

**Every service implements defense-in-depth:**

```yaml
securityContext:
  # Pod-level security
  fsGroup: 1000
  runAsGroup: 1000
  runAsNonRoot: true
  runAsUser: 1000

containers:
  - name: server
    securityContext:
      # Container-level security
      allowPrivilegeEscalation: false
      capabilities:
        drop:
          - ALL  # Drop all Linux capabilities
      privileged: false
      readOnlyRootFilesystem: true  # Immutable filesystem
```

**Key Security Principles:**
- ✅ Run as non-root user (UID 1000)
- ✅ Read-only root filesystem
- ✅ Drop all Linux capabilities
- ✅ No privilege escalation
- ✅ ServiceAccount for IAM integration

### 3.2 Health Checks & Probes

**gRPC Native Health Checks (Kubernetes 1.24+):**

```yaml
readinessProbe:
  initialDelaySeconds: 15
  grpc:
    port: 7070

livenessProbe:
  initialDelaySeconds: 15
  periodSeconds: 10
  grpc:
    port: 7070
```

**All services must implement:**
- `/grpc.health.v1.Health/Check` endpoint
- Readiness: "Can I accept traffic?"
- Liveness: "Am I stuck/deadlocked?"

### 3.3 Resource Management

**CPU and Memory Limits:**

```yaml
resources:
  requests:
    cpu: 100m      # Guaranteed resources
    memory: 64Mi
  limits:
    cpu: 200m      # Maximum allowed
    memory: 128Mi
```

**Recommended Resources by Service Type:**

| Service Type | CPU Request | Memory Request | Rationale |
|--------------|-------------|----------------|-----------|
| Lightweight (Auth, User) | 50m | 32Mi | Stateless CRUD |
| Medium (Feed, Post) | 100m | 64Mi | Some aggregation |
| Heavy (AI, Search) | 200m | 256Mi | Compute intensive |
| Python services | +50% | +100% | Higher overhead |

### 3.4 Graceful Shutdown

```yaml
spec:
  terminationGracePeriodSeconds: 5
```

**Service implementation:**
```go
// Listen for shutdown signals
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
<-sigCh

log.Info("received shutdown signal, draining connections...")
grpcServer.GracefulStop()  // Wait for in-flight requests
```

---

## 4. Data Architecture for iKampus

### 4.1 Database Strategy

**Primary Database: PostgreSQL**

```
Users & Auth
├── users (id, email, username, university_id, created_at)
├── profiles (user_id, course, year, bio, avatar_url)
├── university_verifications (user_id, verification_status, student_id)
└── sessions (session_id, user_id, expires_at)

Social Content
├── posts (id, user_id, content, image_url, created_at)
├── comments (id, post_id, user_id, content, created_at)
├── reactions (post_id, user_id, reaction_type)
└── communities (id, name, university_id, member_count)

Startup Hub
├── startups (id, name, description, stage, demo_url)
├── startup_members (startup_id, user_id, role)
├── investor_connections (startup_id, alumni_id, status)
└── pitches (startup_id, content, created_at)
```

**Caching Layer: Redis**

```
Session Management
├── session:{session_id} → user_data (TTL: 24h)
├── verification_token:{token} → user_id (TTL: 1h)

Feed Cache
├── feed:{user_id} → post_ids[] (TTL: 5min)
├── trending_posts → post_ids[] (TTL: 10min)

Rate Limiting
├── rate_limit:{user_id}:{endpoint} → count (TTL: 1min)

Real-time Counters
├── post_likes:{post_id} → count
├── user_followers:{user_id} → count
```

**Search Index: Elasticsearch**

```json
{
  "posts": {
    "properties": {
      "content": { "type": "text", "analyzer": "english" },
      "user_id": { "type": "keyword" },
      "university_id": { "type": "keyword" },
      "created_at": { "type": "date" },
      "tags": { "type": "keyword" }
    }
  }
}
```

### 4.2 Service Data Ownership

**Each service owns its data:**

| Service | Database | Cache | External |
|---------|----------|-------|----------|
| **UserService** | PostgreSQL (users, profiles) | Redis (sessions) | - |
| **PostService** | PostgreSQL (posts, comments) | Redis (counters) | S3 (images) |
| **FeedService** | - | Redis (feed cache) | Pulls from PostService |
| **StartupService** | PostgreSQL (startups, teams) | Redis (matches) | - |
| **SearchService** | - | - | Elasticsearch |
| **MessagingService** | PostgreSQL (messages) | Redis (presence) | - |

**No shared databases** - Services communicate via APIs only.

---

## 5. Orchestration Patterns

### 5.1 Saga Pattern (Checkout Example)

The CheckoutService orchestrates multiple services:

```go
func (cs *checkoutService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
    // Step 1: Get user's cart
    prep, err := cs.prepareOrderItemsAndShippingQuoteFromCart(ctx, req.UserId, req.UserCurrency, req.Address)
    if err != nil {
        return nil, status.Errorf(codes.Internal, err.Error())
    }

    // Step 2: Calculate total cost
    total := &pb.Money{CurrencyCode: req.UserCurrency, Units: 0, Nanos: 0}
    for _, item := range prep.orderItems {
        total = money.Must(money.Sum(total, item.Cost))
    }
    total = money.Must(money.Sum(total, prep.shippingCostLocalized))

    // Step 3: Charge credit card
    txID, err := cs.chargeCard(ctx, total, req.CreditCard)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to charge card: %+v", err)
    }

    // Step 4: Ship the order
    shippingTrackingID, err := cs.shipOrder(ctx, req.Address, prep.cartItems)
    if err != nil {
        return nil, status.Errorf(codes.Unavailable, "shipping error: %+v", err)
    }

    // Step 5: Empty the cart
    if err := cs.emptyUserCart(ctx, req.UserId); err != nil {
        log.Warnf("failed to empty cart: %+v", err)
    }

    // Step 6: Send confirmation email (fire-and-forget)
    if err := cs.sendOrderConfirmation(ctx, req.Email, orderResult); err != nil {
        log.Warnf("failed to send order confirmation: %+v", err)
    }

    return &pb.PlaceOrderResponse{Order: orderResult}, nil
}
```

### 5.2 iKampus Orchestration: Create Post with AI

```go
// PostService orchestrates:
func (ps *postService) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
    // 1. Moderate content (AI)
    moderation, err := ps.moderateContent(ctx, req.Content)
    if err != nil || !moderation.IsApproved {
        return nil, status.Error(codes.InvalidArgument, "content violates policy")
    }

    // 2. Extract hashtags and mentions
    tags := ps.extractTags(req.Content)
    mentions := ps.extractMentions(req.Content)

    // 3. Save post to database
    post, err := ps.savePost(ctx, req.UserId, req.Content, tags)
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to save post")
    }

    // 4. Update search index (async)
    go ps.indexPost(context.Background(), post)

    // 5. Notify mentioned users (async)
    go ps.notifyMentionedUsers(context.Background(), mentions, post.Id)

    // 6. Invalidate feed cache for followers
    go ps.invalidateFollowerFeeds(context.Background(), req.UserId)

    return &pb.CreatePostResponse{Post: post}, nil
}
```

### 5.3 Error Handling Strategies

**1. Graceful Degradation:**
```go
// Non-critical features don't block main flow
if err := cs.sendOrderConfirmation(ctx, email, order); err != nil {
    log.Warnf("failed to send email: %+v", err)
    // Don't return error - order was successful
}
```

**2. Timeout Pattern:**
```go
// Set timeouts for non-critical services
ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
defer cancel()

ads, err := cs.getAds(ctx, contextKeys)
if err != nil {
    log.Warnf("ad service timeout: %+v", err)
    return emptyAds  // Return empty instead of failing
}
```

**3. Retry with Exponential Backoff:**
```go
for i := 0; i < 3; i++ {
    conn, err := grpc.Dial(addr, opts...)
    if err == nil {
        return conn
    }
    time.Sleep(time.Second * time.Duration(math.Pow(2, float64(i))))
}
return nil, fmt.Errorf("failed to connect after retries")
```

**4. Circuit Breaker Pattern:**
```go
// Use proper gRPC status codes
return status.Errorf(codes.Unavailable, "service temporarily unavailable")
return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
return status.Errorf(codes.DeadlineExceeded, "request timeout")
```

---

## 6. Observability & Monitoring

### 6.1 Distributed Tracing with OpenTelemetry

**Every service implements tracing:**

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

// Initialize tracer
tp, err := initTracerProvider()
if err != nil {
    log.Fatal(err)
}
defer tp.Shutdown(context.Background())

// Set global propagator for trace context
otel.SetTextMapPropagator(
    propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ),
)

// Instrument gRPC server
srv := grpc.NewServer(
    grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
    grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
)
```

**Trace visualization:**
```
CreatePost Request (PostService)
├── ModerateContent → ModerationService [120ms]
├── SavePost → PostgreSQL [45ms]
├── IndexPost → SearchService [async]
├── NotifyUsers → NotificationService [async]
└── InvalidateCache → Redis [12ms]
Total: 177ms
```

### 6.2 Structured Logging

**JSON format for all logs:**

```go
// Go services (logrus)
log.SetFormatter(&logrus.JSONFormatter{
    FieldMap: logrus.FieldMap{
        logrus.FieldKeyTime:  "timestamp",
        logrus.FieldKeyLevel: "severity",
        logrus.FieldKeyMsg:   "message",
    },
})

log.WithFields(logrus.Fields{
    "user_id": userId,
    "post_id": postId,
    "action": "create_post",
}).Info("post created successfully")
```

**Output:**
```json
{
  "timestamp": "2025-11-18T22:30:45Z",
  "severity": "info",
  "message": "post created successfully",
  "user_id": "user_12345",
  "post_id": "post_67890",
  "action": "create_post",
  "service": "postservice"
}
```

### 6.3 Metrics (Prometheus)

**Key metrics to track:**

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    // Request counters
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total number of gRPC requests",
        },
        []string{"service", "method", "status"},
    )

    // Latency histogram
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "grpc_request_duration_seconds",
            Help: "Request duration in seconds",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"service", "method"},
    )

    // Active connections
    activeConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "grpc_active_connections",
            Help: "Number of active gRPC connections",
        },
    )
)
```

### 6.4 Health Checks

**gRPC Health Check Protocol:**

```go
import "google.golang.org/grpc/health/grpc_health_v1"

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
    // Check dependencies
    if err := checkDatabaseConnection(); err != nil {
        return &grpc_health_v1.HealthCheckResponse{
            Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
        }, nil
    }

    if err := checkRedisConnection(); err != nil {
        log.Warn("redis unavailable but service still healthy")
        // Degrade gracefully
    }

    return &grpc_health_v1.HealthCheckResponse{
        Status: grpc_health_v1.HealthCheckResponse_SERVING,
    }, nil
}

// Register health server
grpc_health_v1.RegisterHealthServer(srv, &healthServer{})
```

---

## 7. Service Mesh (Istio Integration)

### 7.1 Traffic Management

**Gateway for external traffic:**

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: ikampus-gateway
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: ikampus-tls-cert
    hosts:
    - "api.ikampus.com"
```

**VirtualService for routing:**

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: api-gateway
spec:
  hosts:
  - "api.ikampus.com"
  gateways:
  - ikampus-gateway
  http:
  - match:
    - uri:
        prefix: "/api/v1"
    route:
    - destination:
        host: gateway-service
        port:
          number: 8080
```

### 7.2 Security Policies

**Mutual TLS between services:**

```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
spec:
  mtls:
    mode: STRICT  # Require mTLS for all services
```

**Authorization policies:**

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: postservice-authz
spec:
  selector:
    matchLabels:
      app: postservice
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/default/sa/gateway-service"]
    to:
    - operation:
        methods: ["POST"]
        paths: ["/ikampus.PostService/CreatePost"]
```

### 7.3 Traffic Splitting (Canary Deployments)

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: postservice
spec:
  hosts:
  - postservice
  http:
  - match:
    - headers:
        x-beta-user:
          exact: "true"
    route:
    - destination:
        host: postservice
        subset: v2
      weight: 100
  - route:
    - destination:
        host: postservice
        subset: v1
      weight: 90
    - destination:
        host: postservice
        subset: v2
      weight: 10  # 10% canary traffic
```

---

## 8. Development Workflow

### 8.1 Local Development with Skaffold

**skaffold.yaml:**

```yaml
apiVersion: skaffold/v4beta6
kind: Config
build:
  artifacts:
  - image: userservice
    context: src/userservice
    docker:
      dockerfile: Dockerfile
  - image: postservice
    context: src/postservice
    docker:
      dockerfile: Dockerfile
  tagPolicy:
    gitCommit: {}

deploy:
  kubectl:
    manifests:
    - kubernetes-manifests/*.yaml

portForward:
- resourceType: service
  resourceName: gateway-service
  port: 8080
  localPort: 8080
```

**Development flow:**
```bash
# Start local Kubernetes cluster
minikube start

# Deploy and watch for changes
skaffold dev

# Hot reload on code changes
# Edit src/userservice/main.go
# Skaffold rebuilds and redeploys automatically
```

### 8.2 Environment Management with Kustomize

**Directory structure:**
```
kustomize/
├── base/
│   ├── kustomization.yaml
│   └── *.yaml (common manifests)
├── components/
│   ├── istio/
│   ├── monitoring/
│   └── redis-ha/
└── overlays/
    ├── dev/
    ├── staging/
    └── production/
```

**Production overlay:**
```yaml
# kustomize/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
- ../../base

components:
- ../../components/istio
- ../../components/monitoring
- ../../components/redis-ha

replicas:
- name: userservice
  count: 3
- name: postservice
  count: 5

images:
- name: userservice
  newName: gcr.io/ikampus/userservice
  newTag: v1.2.3
```

---

## 9. Production Readiness Checklist

### 9.1 Scalability

- [ ] **HorizontalPodAutoscaler** configured for all services
  ```yaml
  apiVersion: autoscaling/v2
  kind: HorizontalPodAutoscaler
  metadata:
    name: postservice-hpa
  spec:
    scaleTargetRef:
      apiVersion: apps/v1
      kind: Deployment
      name: postservice
    minReplicas: 2
    maxReplicas: 10
    metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
  ```

- [ ] **PodDisruptionBudget** for high availability
  ```yaml
  apiVersion: policy/v1
  kind: PodDisruptionBudget
  metadata:
    name: postservice-pdb
  spec:
    minAvailable: 1
    selector:
      matchLabels:
        app: postservice
  ```

- [ ] **Database connection pooling** configured
- [ ] **Redis cluster mode** for cache high availability
- [ ] **CDN** for static assets (images, videos)

### 9.2 Security

- [ ] **NetworkPolicy** for service segmentation
  ```yaml
  apiVersion: networking.k8s.io/v1
  kind: NetworkPolicy
  metadata:
    name: postservice-netpol
  spec:
    podSelector:
      matchLabels:
        app: postservice
    ingress:
    - from:
      - podSelector:
          matchLabels:
            app: gateway-service
      ports:
      - protocol: TCP
        port: 8080
  ```

- [ ] **Secret management** (not environment variables)
  - Use Google Secret Manager / AWS Secrets Manager
  - Or HashiCorp Vault
  - Mount secrets as volumes

- [ ] **TLS for all external endpoints**
- [ ] **API rate limiting** per user/IP
- [ ] **Input validation** on all endpoints
- [ ] **SQL injection prevention** (parameterized queries)
- [ ] **XSS prevention** (sanitize HTML content)

### 9.3 Reliability

- [ ] **Circuit breakers** for external dependencies
- [ ] **Retry policies** with exponential backoff
- [ ] **Timeout policies** for all outbound calls
- [ ] **Graceful degradation** for non-critical features
- [ ] **Database migrations** strategy
- [ ] **Backup and restore** procedures tested
- [ ] **Disaster recovery plan** documented

### 9.4 Observability

- [ ] **Distributed tracing** enabled (OpenTelemetry)
- [ ] **Structured logging** to centralized system
- [ ] **Prometheus metrics** exported from all services
- [ ] **Grafana dashboards** for key metrics
- [ ] **Alerting rules** configured
  - High error rate (> 5%)
  - High latency (p95 > 1s)
  - Pod crash loops
  - Database connection failures

### 9.5 Operations

- [ ] **CI/CD pipeline** automated
- [ ] **Deployment strategy** defined (blue-green, canary)
- [ ] **Rollback procedures** tested
- [ ] **On-call runbooks** created
- [ ] **Incident response plan** documented
- [ ] **Performance testing** completed
- [ ] **Load testing** passed (expected traffic × 3)

---

## 10. Cost Optimization

### 10.1 Resource Right-Sizing

**Analyze actual usage:**
```bash
# Get actual resource usage
kubectl top pods --namespace=ikampus

# Compare to requests/limits
kubectl describe pod postservice-xxx
```

**Adjust based on data:**
```yaml
# Before optimization
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi

# After profiling (actual usage: 80m CPU, 120Mi RAM)
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

### 10.2 Autoscaling Strategy

**Scale based on actual load:**
- **Off-peak hours** (2am-6am): minReplicas: 1
- **Business hours** (9am-11pm): minReplicas: 3
- **Peak hours** (6pm-9pm): maxReplicas: 10

### 10.3 Caching Strategy

**Reduce database load:**
- Cache frequently accessed data (user profiles, trending posts)
- Use TTL to balance freshness vs. performance
- Implement cache warming for predictable traffic

---

## 11. Key Takeaways for iKampus

### 11.1 Adopt from Google's Architecture

✅ **Use gRPC for internal services** - Type-safe, efficient, polyglot
✅ **Implement strong security defaults** - Non-root, read-only, minimal privileges
✅ **Build observability from day one** - Tracing, logging, metrics
✅ **Design for failure** - Retries, timeouts, graceful degradation
✅ **Automate everything** - CI/CD, testing, deployments

### 11.2 Adapt for iKampus Needs

🔧 **Add HPA for autoscaling** - Google demo doesn't include it
🔧 **Implement proper authentication** - JWT, OAuth for mobile/web
🔧 **Use managed services** - Cloud SQL, Redis, Cloud Storage
🔧 **Add real-time features** - WebSocket for chat, notifications
🔧 **Implement event-driven architecture** - Pub/Sub for async workflows

### 11.3 Avoid Common Pitfalls

❌ **Don't create too many microservices** - Start with 5-7, split later
❌ **Don't share databases** - Each service owns its data
❌ **Don't skip observability** - Debug production issues efficiently
❌ **Don't ignore security** - Implement it from the start
❌ **Don't hardcode configuration** - Use ConfigMaps, Secrets

---

## 12. Next Steps

### Phase 1: Foundation (Weeks 1-4)
1. Define Protocol Buffer contracts for core services
2. Implement UserService and AuthService
3. Set up Kubernetes cluster (GKE/EKS/AKS)
4. Configure CI/CD pipeline
5. Implement observability stack (OpenTelemetry, Prometheus, Grafana)

### Phase 2: Core Features (Weeks 5-10)
1. Implement PostService and FeedService
2. Add MessagingService
3. Build API Gateway
4. Deploy to staging environment
5. Implement caching strategy

### Phase 3: Advanced Features (Weeks 11-16)
1. Implement AIAssistantService
2. Add StartupService and MatchmakingService
3. Implement SearchService with Elasticsearch
4. Add real-time notifications
5. Performance testing and optimization

### Phase 4: Production Launch (Weeks 17-20)
1. Security audit and penetration testing
2. Load testing and capacity planning
3. Disaster recovery testing
4. Documentation and runbooks
5. Production deployment

---

## Appendix: Useful Commands

### Kubernetes

```bash
# Deploy all services
kubectl apply -f kubernetes-manifests/

# Check pod status
kubectl get pods -w

# View logs
kubectl logs -f postservice-xxx

# Port forward for local testing
kubectl port-forward svc/gateway-service 8080:8080

# Scale deployment
kubectl scale deployment postservice --replicas=5

# Restart deployment
kubectl rollout restart deployment postservice

# Check resource usage
kubectl top pods
kubectl top nodes
```

### gRPC Testing

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:8080 list

# Call method
grpcurl -plaintext -d '{"user_id": "user_123"}' \
  localhost:8080 ikampus.PostService/GetUserPosts
```

### Redis

```bash
# Connect to Redis
kubectl exec -it redis-0 -- redis-cli

# Check keys
KEYS *

# Get value
GET session:abc123

# Monitor commands
MONITOR
```

---

**Document Version:** 1.0
**Last Updated:** 2025-11-18
**Author:** Claude (AI Assistant)
**Based on:** [GoogleCloudPlatform/microservices-demo](https://github.com/GoogleCloudPlatform/microservices-demo)
