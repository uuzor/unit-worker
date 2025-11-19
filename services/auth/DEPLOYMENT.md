# AuthService Production Deployment Guide

Complete guide for deploying AuthService to production environments.

## Table of Contents

1. [Production Checklist](#production-checklist)
2. [Kubernetes Deployment](#kubernetes-deployment)
3. [Configuration Management](#configuration-management)
4. [Database Setup](#database-setup)
5. [Monitoring & Observability](#monitoring--observability)
6. [Security Hardening](#security-hardening)
7. [Disaster Recovery](#disaster-recovery)

## Production Checklist

Before deploying to production, ensure:

### Security
- [ ] Strong JWT secrets (minimum 32 characters)
- [ ] Database uses TLS/SSL connections
- [ ] Secrets managed via Kubernetes Secrets or external vault
- [ ] HTTPS/TLS enabled for external endpoints
- [ ] Service mesh (Istio) configured for mTLS
- [ ] Network policies implemented
- [ ] Non-root container user
- [ ] Read-only root filesystem

### Performance
- [ ] Database connection pooling configured
- [ ] Redis caching enabled
- [ ] Resource limits set (CPU, memory)
- [ ] Horizontal Pod Autoscaling (HPA) configured
- [ ] Database indexes optimized
- [ ] Connection timeouts configured

### Observability
- [ ] Prometheus metrics exported
- [ ] Distributed tracing enabled (OpenTelemetry)
- [ ] Structured logging configured
- [ ] Health checks implemented
- [ ] Alerts configured
- [ ] Dashboards created (Grafana)

### Reliability
- [ ] Multiple replicas running (min 3)
- [ ] Pod disruption budgets set
- [ ] Graceful shutdown implemented
- [ ] Database backups automated
- [ ] Disaster recovery plan documented
- [ ] Load testing completed

## Kubernetes Deployment

### 1. Create Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ikampus
  labels:
    name: ikampus
    environment: production
```

```bash
kubectl apply -f namespace.yaml
```

### 2. Create Secrets

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
  namespace: ikampus
type: Opaque
stringData:
  jwt-access-secret: "your-super-secret-access-key-minimum-32-characters-long"
  jwt-refresh-secret: "your-super-secret-refresh-key-minimum-32-characters-long"
  db-password: "your-database-password"
  smtp-password: "your-smtp-password"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
  namespace: ikampus
data:
  SERVER_PORT: "8080"
  GRPC_PORT: "9000"
  ENVIRONMENT: "production"
  SHUTDOWN_TIMEOUT: "30s"
  DB_HOST: "postgres.ikampus.svc.cluster.local"
  DB_PORT: "5432"
  DB_USER: "ikampus_user"
  DB_NAME: "ikampus"
  DB_SSLMODE: "require"
  DB_MAX_OPEN_CONNS: "50"
  DB_MAX_IDLE_CONNS: "10"
  DB_CONN_MAX_LIFETIME: "5m"
  JWT_ACCESS_EXPIRY: "15m"
  JWT_REFRESH_EXPIRY: "720h"
  JWT_ISSUER: "ikampus-auth-production"
  SMTP_HOST: "smtp.gmail.com"
  SMTP_PORT: "587"
  FROM_EMAIL: "noreply@ikampus.com"
  FROM_NAME: "iKampus"
  VERIFICATION_EXPIRY: "24h"
  REDIS_HOST: "redis.ikampus.svc.cluster.local"
  REDIS_PORT: "6379"
  REDIS_DB: "0"
```

```bash
kubectl apply -f secrets.yaml
```

### 3. Deploy AuthService

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: ikampus
  labels:
    app: auth-service
    version: v1
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: auth-service
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        fsGroup: 65534
      containers:
      - name: auth-service
        image: ikampus/auth-service:v1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: grpc
          containerPort: 9000
          protocol: TCP
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: db-password
        - name: JWT_ACCESS_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: jwt-access-secret
        - name: JWT_REFRESH_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: jwt-refresh-secret
        - name: SMTP_PASSWORD
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: smtp-password
        envFrom:
        - configMapRef:
            name: auth-service-config
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 65534
          capabilities:
            drop:
            - ALL
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: auth-service
  namespace: ikampus
```

```bash
kubectl apply -f deployment.yaml
```

### 4. Create Services

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: ikampus
  labels:
    app: auth-service
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 8080
    targetPort: 8080
    protocol: TCP
  - name: grpc
    port: 9000
    targetPort: 9000
    protocol: TCP
  selector:
    app: auth-service
---
# For external access (if needed)
apiVersion: v1
kind: Service
metadata:
  name: auth-service-external
  namespace: ikampus
  labels:
    app: auth-service
spec:
  type: LoadBalancer
  ports:
  - name: grpc
    port: 9000
    targetPort: 9000
    protocol: TCP
  selector:
    app: auth-service
```

```bash
kubectl apply -f service.yaml
```

### 5. Horizontal Pod Autoscaling

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: auth-service-hpa
  namespace: ikampus
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: auth-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
      selectPolicy: Max
```

```bash
kubectl apply -f hpa.yaml
```

### 6. Pod Disruption Budget

```yaml
# pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: auth-service-pdb
  namespace: ikampus
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: auth-service
```

```bash
kubectl apply -f pdb.yaml
```

### 7. Network Policy

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: auth-service-netpol
  namespace: ikampus
spec:
  podSelector:
    matchLabels:
      app: auth-service
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ikampus
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 9000
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: ikampus
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
  - to: # DNS
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
```

```bash
kubectl apply -f network-policy.yaml
```

## Configuration Management

### Using Kustomize

Create `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: ikampus

resources:
  - namespace.yaml
  - secrets.yaml
  - deployment.yaml
  - service.yaml
  - hpa.yaml
  - pdb.yaml
  - network-policy.yaml

images:
  - name: ikampus/auth-service
    newTag: v1.0.0

commonLabels:
  app.kubernetes.io/name: auth-service
  app.kubernetes.io/part-of: ikampus
  app.kubernetes.io/managed-by: kustomize
```

Deploy with Kustomize:
```bash
kubectl apply -k .
```

### Using Helm

Create `values.yaml`:

```yaml
replicaCount: 3

image:
  repository: ikampus/auth-service
  tag: v1.0.0
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  httpPort: 8080
  grpcPort: 9000

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

config:
  environment: production
  database:
    host: postgres.ikampus.svc.cluster.local
    port: 5432
    user: ikampus_user
    name: ikampus
    sslmode: require
  jwt:
    accessExpiry: 15m
    refreshExpiry: 720h
    issuer: ikampus-auth-production
```

## Database Setup

### PostgreSQL on Kubernetes

```yaml
# postgres-statefulset.yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: ikampus
spec:
  ports:
  - port: 5432
    name: postgres
  clusterIP: None
  selector:
    app: postgres
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: ikampus
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:14-alpine
        ports:
        - containerPort: 5432
          name: postgres
        env:
        - name: POSTGRES_DB
          value: ikampus
        - name: POSTGRES_USER
          value: ikampus_user
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: db-password
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 4Gi
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
```

### Database Migrations

```bash
# Create a Job for database initialization
kubectl create job --from=cronjob/db-migrate db-migrate-manual -n ikampus

# Or use init container in deployment
```

## Monitoring & Observability

### Prometheus ServiceMonitor

```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: auth-service
  namespace: ikampus
  labels:
    app: auth-service
spec:
  selector:
    matchLabels:
      app: auth-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### Grafana Dashboard

Import dashboard ID: (Create custom dashboard)

Key metrics to monitor:
- Request rate (requests/second)
- Error rate (errors/second)
- Request latency (p50, p95, p99)
- Active connections
- JWT token generation time
- Database query duration
- Cache hit/miss ratio

### Logging with EFK Stack

Configure fluent-bit to collect logs:

```yaml
# Logs are already in JSON format from the service
# Configure index pattern in Kibana: ikampus-auth-*
```

## Security Hardening

### 1. Use Private Container Registry

```bash
# Create image pull secret
kubectl create secret docker-registry regcred \
  --docker-server=your-registry.io \
  --docker-username=your-username \
  --docker-password=your-password \
  --docker-email=your-email \
  -n ikampus

# Reference in deployment
spec:
  template:
    spec:
      imagePullSecrets:
      - name: regcred
```

### 2. External Secrets Operator

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
  namespace: ikampus
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      version: "v2"
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: auth-service-secrets
  namespace: ikampus
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: auth-service-secrets
  data:
  - secretKey: jwt-access-secret
    remoteRef:
      key: ikampus/auth
      property: jwt-access-secret
```

### 3. Pod Security Standards

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ikampus
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

## Disaster Recovery

### Database Backups

```yaml
# backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: ikampus
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:14-alpine
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h postgres -U ikampus_user ikampus | \
              gzip > /backup/ikampus-$(date +%Y%m%d-%H%M%S).sql.gz
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: auth-service-secrets
                  key: db-password
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
          restartPolicy: OnFailure
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: postgres-backup-pvc
```

### Restore Procedure

```bash
# 1. Copy backup from storage
kubectl cp ikampus/postgres-backup-pod:/backup/ikampus-20240115.sql.gz ./backup.sql.gz

# 2. Decompress
gunzip backup.sql.gz

# 3. Restore
kubectl exec -it postgres-0 -n ikampus -- psql -U ikampus_user -d ikampus < backup.sql
```

## Deployment Workflow

### CI/CD Pipeline (GitHub Actions)

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    tags:
      - 'v*'

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: |
          docker build -t ikampus/auth-service:${{ github.ref_name }} .
          docker push ikampus/auth-service:${{ github.ref_name }}

      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/auth-service \
            auth-service=ikampus/auth-service:${{ github.ref_name }} \
            -n ikampus
          kubectl rollout status deployment/auth-service -n ikampus
```

### Rollback

```bash
# View rollout history
kubectl rollout history deployment/auth-service -n ikampus

# Rollback to previous version
kubectl rollout undo deployment/auth-service -n ikampus

# Rollback to specific revision
kubectl rollout undo deployment/auth-service --to-revision=2 -n ikampus
```

## Performance Tuning

### Database Optimization

```sql
-- Create indexes
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX CONCURRENTLY idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- Analyze tables
ANALYZE users;
ANALYZE refresh_tokens;
```

### Connection Pooling

Adjust based on load:
```env
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m
```

## Health Checks

The service exposes three endpoints:

- `/health` - Overall health (database connectivity)
- `/ready` - Readiness (can accept traffic)
- `/live` - Liveness (process is alive)

## Support and Maintenance

### Regular Maintenance Tasks

- [ ] Weekly: Review logs for errors
- [ ] Weekly: Check resource usage
- [ ] Monthly: Review and optimize database queries
- [ ] Monthly: Update dependencies
- [ ] Quarterly: Load testing
- [ ] Quarterly: Disaster recovery drill

### Monitoring Alerts

Configure alerts for:
- High error rate (> 1%)
- High latency (p99 > 1s)
- Database connection pool exhaustion
- High CPU/memory usage (> 80%)
- Pod restart rate
- Failed health checks
