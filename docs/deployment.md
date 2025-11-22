# Lokstra Auth - Deployment Guide

## Quick Start

### 1. First Run (Generate Registration Code)

```bash
# Run with Bootstrap to auto-generate registration code
go run main.go
```

Saat pertama kali running, `lokstra.Bootstrap()` akan:
- Scan semua `@RouterService` annotations
- Generate `lokstra_registry/generated.go` dengan registrasi otomatis
- Setup dependency injection

### 2. Choose Deployment Mode

Set environment variable `SERVER` untuk memilih mode deployment:

```bash
# Monolith (default): Semua service dalam 1 server
export SERVER=monolith
go run main.go

# Microservices: Service terpisah per layer
export SERVER=microservices
go run main.go

# Development: Debug mode
export SERVER=development
go run main.go
```

## Deployment Modes

### Mode 1: Monolith (Production Default)

**Server**: `auth-server` di port `8080`

**Services**:
- `00_core`: Tenant, App, Branch, User, AppKey, CredentialConfig
- `01_credential`: Basic, APIKey, OAuth2, Passwordless, Passkey
- `02_token`: Token generation (future)
- `03_subject`: Subject resolution (future)
- `04_authz`: Authorization (future)

**Endpoints**:
```
http://localhost:8080/api/registration/*    # Core services
http://localhost:8080/api/cred/*            # Credential services
```

**Use Cases**:
- Small to medium deployments
- Simple infrastructure
- Lower latency (no network calls between services)
- Easy to debug

### Mode 2: Microservices

**Servers**:
- `core-server`: Port `8081` → Core services
- `credential-server`: Port `8082` → Authentication services

**Endpoints**:
```
http://localhost:8081/api/registration/*    # Core services
http://localhost:8082/api/cred/*            # Credential services
```

**Use Cases**:
- Large scale deployments
- Independent scaling (scale auth service separately from core)
- Team separation (different teams own different services)
- Fault isolation

### Mode 3: Development

**Server**: `dev-server` di port `3000`

**Features**:
- Hot reload friendly
- Debug logging enabled
- CORS permissive
- All services available

## Configuration File

`config/deployment.yaml`:

```yaml
configs:
  server: ${SERVER:monolith}  # Environment variable with default

deployments:
  monolith:
    servers:
      auth-server:
        base-url: "http://localhost"
        addr: ":8080"
        published-services:
          - tenant-service
          - app-service
          - basic-auth-service
          - apikey-auth-service
          # ... etc
```

## Service Registry

All services are auto-registered via annotations:

```go
// @RouterService name="basic-auth-service", prefix="/api/cred/basic"
type BasicAuthService struct {
    // @Inject "user-provider"
    UserProvider *service.Cached[UserProvider]
}

// @Route "POST /login"
func (s *BasicAuthService) Login(ctx *request.Context, req *LoginRequest) (*LoginResponse, error) {
    // ...
}
```

**Generated Registration** (`lokstra_registry/generated.go`):
```go
func RegisterAllServiceTypes() {
    lokstra.RegisterServiceType("basic-auth-service", &BasicAuthService{})
    lokstra.RegisterServiceType("apikey-auth-service", &APIKeyAuthService{})
    // ... etc
}
```

## Service Discovery

Services auto-discover dependencies via `@Inject`:

```go
type BasicAuthService struct {
    // @Inject "user-provider"
    UserProvider *service.Cached[repository.UserProvider]
    
    // @Inject "credential-config-resolver"
    ConfigResolver *service.Cached[*ConfigResolver]
}
```

Lokstra akan otomatis:
1. Resolve dependencies
2. Inject implementations
3. Handle circular dependency detection
4. Cache service instances

## Running in Production

### Docker Compose (Monolith)

```yaml
version: '3.8'
services:
  lokstra-auth:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SERVER=monolith
      - LOG_LEVEL=info
    volumes:
      - ./config:/app/config
```

### Kubernetes (Microservices)

```yaml
# core-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-core
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: core
        image: lokstra-auth:latest
        env:
        - name: SERVER
          value: "microservices"
        - name: SERVICE_NAME
          value: "core-server"
        ports:
        - containerPort: 8081

---
# credential-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-credential
spec:
  replicas: 5  # Scale auth service independently
  template:
    spec:
      containers:
      - name: credential
        image: lokstra-auth:latest
        env:
        - name: SERVER
          value: "microservices"
        - name: SERVICE_NAME
          value: "credential-server"
        ports:
        - containerPort: 8082
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER` | `monolith` | Deployment mode: monolith, microservices, development |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `CONFIG_PATH` | `config/deployment.yaml` | Path to deployment config |

## Health Checks

All servers expose health endpoints:

```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Metrics (Prometheus format)
curl http://localhost:8080/metrics
```

## Monitoring & Observability

### Structured Logging

All services use structured logging:

```json
{
  "level": "info",
  "service": "basic-auth-service",
  "method": "Login",
  "tenant_id": "tenant_abc",
  "user_id": "user_123",
  "duration_ms": 45,
  "status": "success"
}
```

### Metrics

Prometheus metrics available:

- `http_requests_total`: Total HTTP requests
- `http_request_duration_seconds`: Request latency
- `auth_attempts_total`: Authentication attempts
- `auth_failures_total`: Authentication failures
- `active_sessions_gauge`: Current active sessions

### Tracing

Distributed tracing via OpenTelemetry:

- Request ID propagation
- Service-to-service tracing
- Database query tracing
- External API call tracing

## Scaling Strategies

### Horizontal Scaling

```bash
# Scale credential service
kubectl scale deployment auth-credential --replicas=10

# Core service stays at 3 replicas
kubectl scale deployment auth-core --replicas=3
```

### Vertical Scaling

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

## Troubleshooting

### Service Not Found

```bash
# Check service registry
curl http://localhost:8080/debug/services
```

### Dependency Injection Failed

```bash
# Check dependency graph
curl http://localhost:8080/debug/dependencies
```

### Performance Issues

```bash
# Enable profiling
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze with pprof
go tool pprof cpu.prof
```

## Migration Path

### From Monolith to Microservices

1. Start with monolith in production
2. Profile to identify hot services
3. Split hot services to separate servers
4. Update `config/deployment.yaml`
5. Deploy microservices incrementally
6. Monitor and adjust

### From Other Auth Systems

1. Run Lokstra Auth in parallel
2. Gradually migrate apps to Lokstra
3. Use dual-write during transition
4. Decommission old system after full migration

## Next Steps

- [Configuration Management](./credential_configuration.md)
- [Multi-Tenant Setup](./multi_tenant_architecture.md)
- [API Documentation](./api_reference.md)
