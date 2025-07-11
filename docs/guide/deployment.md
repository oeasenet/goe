# Deployment

This guide covers deploying GOE applications to production environments, including Docker, cloud platforms, and monitoring setup.

## Production Checklist

Before deploying to production:

- [ ] **Environment Configuration**: Set up production environment variables
- [ ] **Database**: Configure production database with proper connection pooling
- [ ] **Logging**: Set up structured logging with appropriate log levels
- [ ] **Security**: Enable HTTPS, set up authentication, validate all inputs
- [ ] **Monitoring**: Configure health checks, metrics, and alerting
- [ ] **Performance**: Set up caching, optimize database queries
- [ ] **Backup**: Implement database backup and recovery procedures

## Docker Deployment

### Dockerfile

```dockerfile
# Multi-stage build for production
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

# Change ownership
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./main"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - HTTP_PORT=8080
      - DB_HOST=db
      - CACHE_DRIVER=redis
      - REDIS_HOST=redis
    depends_on:
      - db
      - redis
    restart: unless-stopped
    networks:
      - app-network

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: appuser
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - db_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    restart: unless-stopped
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped
    networks:
      - app-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    restart: unless-stopped
    networks:
      - app-network

volumes:
  db_data:
  redis_data:

networks:
  app-network:
    driver: bridge
```

### Production Environment Variables

```bash
# .env.production
APP_ENV=production
APP_NAME=My GOE App
APP_VERSION=1.0.0

# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s

# Database
DB_DRIVER=postgres
DB_HOST=db
DB_PORT=5432
DB_NAME=myapp
DB_USER=appuser
DB_PASSWORD=secure_password_here
DB_SSL_MODE=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# Cache
CACHE_DRIVER=redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=redis_password_here

# Logging
LOG_LEVEL=warn
LOG_FORMAT=json

# Security
JWT_SECRET=super_secret_jwt_key_here
BCRYPT_COST=12
```

## Cloud Deployment

### AWS ECS

```json
{
  "family": "goe-app",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "goe-app",
      "image": "your-registry/goe-app:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "APP_ENV",
          "value": "production"
        },
        {
          "name": "HTTP_PORT",
          "value": "8080"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:db-password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/goe-app",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3
      }
    }
  ]
}
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goe-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: goe-app
  template:
    metadata:
      labels:
        app: goe-app
    spec:
      containers:
      - name: goe-app
        image: your-registry/goe-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: APP_ENV
          value: "production"
        - name: HTTP_PORT
          value: "8080"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: password
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: goe-app-service
spec:
  selector:
    app: goe-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: goe-app-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - myapp.com
    secretName: goe-app-tls
  rules:
  - host: myapp.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: goe-app-service
            port:
              number: 80
```

## Health Checks

### Health Check Handler

```go
package handler

import (
    "context"
    "time"
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/v2/contract"
)

type HealthHandler struct {
    db     contract.DB
    cache  contract.Cache
    logger contract.Logger
}

func NewHealthHandler(db contract.DB, cache contract.Cache, logger contract.Logger) *HealthHandler {
    return &HealthHandler{db: db, cache: cache, logger: logger}
}

func (h *HealthHandler) HealthCheck(c fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    checks := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().Unix(),
        "checks": map[string]interface{}{
            "database": h.checkDatabase(ctx),
            "cache":    h.checkCache(ctx),
        },
    }
    
    // Check if any service is unhealthy
    for _, check := range checks["checks"].(map[string]interface{}) {
        if status, ok := check.(map[string]interface{})["status"]; ok && status != "healthy" {
            checks["status"] = "unhealthy"
            return c.Status(503).JSON(checks)
        }
    }
    
    return c.JSON(checks)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) interface{} {
    if err := h.db.Instance().WithContext(ctx).Exec("SELECT 1").Error; err != nil {
        h.logger.Error("Database health check failed", "error", err)
        return map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    }
    
    return map[string]interface{}{
        "status": "healthy",
    }
}

func (h *HealthHandler) checkCache(ctx context.Context) interface{} {
    testKey := "health_check"
    testValue := "ok"
    
    if err := h.cache.Set(testKey, testValue, 10*time.Second); err != nil {
        h.logger.Error("Cache health check failed", "error", err)
        return map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    }
    
    if _, err := h.cache.Get(testKey); err != nil {
        h.logger.Error("Cache read health check failed", "error", err)
        return map[string]interface{}{
            "status": "unhealthy",
            "error":  err.Error(),
        }
    }
    
    return map[string]interface{}{
        "status": "healthy",
    }
}
```

## Monitoring Setup

### Prometheus Metrics

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Duration of HTTP requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    
    ActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Number of active connections",
        },
    )
)

func RecordMetrics(method, path string, status int, duration float64) {
    RequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
    RequestDuration.WithLabelValues(method, path).Observe(duration)
}
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "GOE Application Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"4..|5..\"}[5m])",
            "legendFormat": "Error Rate"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

## Security Configuration

### HTTPS/TLS

```nginx
# nginx.conf
server {
    listen 80;
    server_name myapp.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name myapp.com;
    
    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    
    location / {
        proxy_pass http://app:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Rate Limiting

```go
import "github.com/gofiber/fiber/v3/middleware/limiter"

func SetupRateLimiting(app *fiber.App) {
    // Global rate limiting
    app.Use(limiter.New(limiter.Config{
        Max:        100,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c fiber.Ctx) string {
            return c.Get("X-Forwarded-For")
        },
    }))
    
    // API rate limiting
    api := app.Group("/api")
    api.Use(limiter.New(limiter.Config{
        Max:        1000,
        Expiration: 1 * time.Hour,
        KeyGenerator: func(c fiber.Ctx) string {
            return c.Get("Authorization") // Use token for authenticated limits
        },
    }))
}
```

## Backup and Recovery

### Database Backup

```bash
#!/bin/bash
# backup.sh

DB_NAME=${DB_NAME:-myapp}
DB_USER=${DB_USER:-appuser}
DB_HOST=${DB_HOST:-localhost}
BACKUP_DIR=${BACKUP_DIR:-/backups}

# Create backup directory
mkdir -p $BACKUP_DIR

# Generate backup filename
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/db_backup_$DATE.sql"

# Create backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Remove old backups (keep last 7 days)
find $BACKUP_DIR -name "db_backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

### Recovery Script

```bash
#!/bin/bash
# restore.sh

BACKUP_FILE=$1
DB_NAME=${DB_NAME:-myapp}
DB_USER=${DB_USER:-appuser}
DB_HOST=${DB_HOST:-localhost}

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file.sql.gz>"
    exit 1
fi

# Extract backup
gunzip -c $BACKUP_FILE > temp_restore.sql

# Restore database
psql -h $DB_HOST -U $DB_USER -d $DB_NAME < temp_restore.sql

# Cleanup
rm temp_restore.sql

echo "Database restored from $BACKUP_FILE"
```

## CI/CD Pipeline

### GitHub Actions

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: go test -v -race ./...
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Build Docker image
      run: |
        docker build -t ${{ secrets.DOCKER_REGISTRY }}/goe-app:${{ github.sha }} .
        docker tag ${{ secrets.DOCKER_REGISTRY }}/goe-app:${{ github.sha }} ${{ secrets.DOCKER_REGISTRY }}/goe-app:latest
    
    - name: Push to registry
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker push ${{ secrets.DOCKER_REGISTRY }}/goe-app:${{ github.sha }}
        docker push ${{ secrets.DOCKER_REGISTRY }}/goe-app:latest

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: production
    steps:
    - name: Deploy to production
      run: |
        # Update Kubernetes deployment
        kubectl set image deployment/goe-app goe-app=${{ secrets.DOCKER_REGISTRY }}/goe-app:${{ github.sha }}
        kubectl rollout status deployment/goe-app
```

## Best Practices

1. **Zero-downtime deployments**: Use rolling updates or blue-green deployments
2. **Environment isolation**: Separate staging and production environments
3. **Secret management**: Use secure secret management systems
4. **Monitoring**: Set up comprehensive monitoring and alerting
5. **Backup**: Implement automated backup and recovery procedures
6. **Security**: Keep dependencies updated and scan for vulnerabilities
7. **Performance**: Monitor and optimize application performance
8. **Documentation**: Maintain deployment and operational documentation

## Next Steps

- [**Best Practices**](./best-practices.md) - Follow production best practices