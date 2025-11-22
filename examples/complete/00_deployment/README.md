# Deployment Example

This example demonstrates how to deploy Lokstra Auth framework with different deployment modes.

## Overview

The Lokstra framework supports multiple deployment strategies:
- **Monolith** - All services in a single server (default, port 8080)
- **Microservices** - Each layer as separate service (ports 8081-8085)
- **Development** - Debug mode with hot reload (port 3000)

## Quick Start

```bash
cd examples/complete/00_deployment

# Run in monolith mode (default)
go run main.go

# Run in microservices mode
SERVER=microservices go run main.go

# Run in development mode
SERVER=development go run main.go
```

## Files

- `main.go` - Application entry point with Bootstrap
- `config/deployment.yaml` - Deployment configuration

## How It Works

### 1. Bootstrap Phase

```go
lokstra.Bootstrap()
```

The Bootstrap function:
- Scans all packages for `@RouterService` annotations
- Generates `lokstra_registry/generated.go` with registration code
- Auto-wires all services

### 2. Service Registration

```go
lokstra_registry.RegisterAllServiceTypes()
```

Registers all service types from auto-generated registry:
- Core services (tenant, app, branch, user)
- Credential services (basic, apikey, oauth2, passwordless, passkey)
- Token services (JWT, simple)
- Subject services (resolver, enriched)
- Authorization services (RBAC, ABAC, ACL, policy)

### 3. Deployment Loading

```go
manager, err := deploy.LoadDeploymentFromConfig("config/deployment.yaml")
```

Loads deployment configuration:
- Reads `config/deployment.yaml`
- Selects deployment mode from `SERVER` environment variable
- Configures service instances and routing

### 4. Server Execution

```go
manager.RunAllServers()
```

Starts HTTP servers based on deployment mode:
- **Monolith**: Single server on port 8080
- **Microservices**: 5 servers (core:8081, credential:8082, token:8083, subject:8084, authz:8085)
- **Development**: Single server on port 3000 with debug logging

## Deployment Modes

### Monolith (Default)

**Use case**: Simple deployments, small-medium apps

```bash
go run main.go
# or
SERVER=monolith go run main.go
```

**Architecture**:
```
Port 8080 (All Services)
├── /api/core/*         - Core services
├── /api/cred/*         - Credential services
├── /api/token/*        - Token services
├── /api/subject/*      - Subject services
└── /api/authz/*        - Authorization services
```

### Microservices

**Use case**: Large-scale apps, independent scaling

```bash
SERVER=microservices go run main.go
```

**Architecture**:
```
Port 8081 - Core Service
└── /api/core/*

Port 8082 - Credential Service
└── /api/cred/*

Port 8083 - Token Service
└── /api/token/*

Port 8084 - Subject Service
└── /api/subject/*

Port 8085 - Authorization Service
└── /api/authz/*
```

### Development

**Use case**: Local development, debugging

```bash
SERVER=development go run main.go
```

**Features**:
- Debug logging enabled
- Port 3000 for easy testing
- Hot reload support (with external tools)

## Configuration

Edit `config/deployment.yaml` to customize:

```yaml
configs:
  jwt_secret: "your-secret-key-change-in-production"
  log_level: "info"
  environment: "production"

deployments:
  monolith:
    instances:
      - name: "all-services"
        port: 8080
        services:
          - core-service
          - credential-service
          - token-service
          - subject-service
          - authz-service
```

## Service Registry

After first run, check auto-generated code:

```bash
cat lokstra_registry/generated.go
```

You'll see registration code for all services:

```go
func RegisterAllServiceTypes() {
    registry.RegisterServiceType(&application.TenantService{})
    registry.RegisterServiceType(&application.AppService{})
    registry.RegisterServiceType(&application.BasicAuthService{})
    // ... etc
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER` | Deployment mode | `monolith` |
| `LOG_LEVEL` | Logging level | `info` |
| `CONFIG_PATH` | Config file path | `config/deployment.yaml` |
| `JWT_SECRET` | JWT signing key | (from config) |

## Testing

Test each deployment mode:

```bash
# Test monolith
go run main.go &
curl http://localhost:8080/api/core/health

# Test microservices
SERVER=microservices go run main.go &
curl http://localhost:8081/api/core/health
curl http://localhost:8082/api/cred/health

# Test development
SERVER=development go run main.go &
curl http://localhost:3000/api/core/health
```

## Production Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o lokstra-auth main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/lokstra-auth .
COPY config/ ./config/
ENV SERVER=monolith
EXPOSE 8080
CMD ["./lokstra-auth"]
```

Build and run:
```bash
docker build -t lokstra-auth .
docker run -p 8080:8080 -e SERVER=monolith lokstra-auth
```

### Kubernetes

For microservices deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: core-service
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
          value: "core-service"
        ports:
        - containerPort: 8081
```

## Monitoring

Add Prometheus metrics:

```go
import "github.com/primadi/lokstra/observability"

func main() {
    lokstra.Bootstrap()
    
    // Add metrics
    observability.EnableMetrics()
    
    // ... rest of code
}
```

## Scaling

**Monolith scaling**:
- Vertical: Increase CPU/memory
- Horizontal: Load balancer + multiple instances

**Microservices scaling**:
- Scale each service independently
- Core: 3-5 replicas
- Credential: 5-10 replicas (high traffic)
- Token: 3-5 replicas
- Subject: 2-3 replicas (cached)
- Authz: 2-3 replicas

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
netstat -ano | findstr :8080

# Kill process (Windows)
taskkill /PID <PID> /F
```

### Service Not Found

Check service registration:
```bash
grep "RegisterServiceType" lokstra_registry/generated.go
```

### Configuration Error

Validate YAML:
```bash
# Install yamllint
pip install yamllint

# Validate
yamllint config/deployment.yaml
```

## Next Steps

- See [Credential Providers](../../../docs/credential_providers.md) for authentication setup
- See [Multi-Tenant Architecture](../../../docs/multi_tenant_architecture.md) for SaaS apps
- See [Complete Examples](../) for end-to-end flows

## Related Examples

- `00_core/01_app_keys` - App API key management
- `00_core/02_credential_config` - Credential configuration
- `01_basic_flow` - Complete authentication flow
- `02_multi_auth` - Multiple auth methods

---

**This is a framework project** - `main.go` is an example, not the framework entry point.
