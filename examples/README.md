# Examples

This folder contains usage examples for the Lokstra Auth framework, organized by layer and complexity.

## 📁 Structure

```
examples/
├── complete/              # ⭐ START HERE - Complete runnable examples
│   ├── core/          # Core services (tenant, app, config)
│   ├── 01_basic_flow/    # Complete auth flow
│   └── 02_multi_auth/    # Multiple auth methods
│
├── credential/         # Layer 1: Authentication methods
├── token/             # Layer 2: Token management
├── rbac/           # Layer 3: Subject resolution
├── authz/             # Layer 4: Authorization
├── middleware/           # HTTP middleware examples
└── services/             # Service integration examples
```

## 🚀 Quick Start

### For Beginners - Start Here! 

**Complete Examples** (`/complete/`) are standalone, runnable programs:

```bash
# Understand deployment modes (RECOMMENDED FIRST!)
go run ./examples/01_deployment/main.go
```

> **Note**: The framework now uses `@RouterService` annotations with auto-wiring.  
> Examples that require manual service instantiation are being redesigned.  
> See the deployment example to understand how services are registered automatically.

## 📚 Example Categories

### ⭐ `/complete` - Complete Integration Examples

**Purpose**: Full end-to-end examples that you can run immediately

**Examples**:
- `00_deployment` - **⭐ START HERE** - Deployment modes (monolith/microservices/development)

**Best for**: Understanding the framework deployment and auto-registration

> **Note**: Additional examples being redesigned for annotation-based architecture (`@RouterService`, `@Inject`)

### 🔐 `/credential` - Layer 1: Authentication

**Status**: 🚧 Being updated to new layered architecture (application/domain/infrastructure)

**Purpose**: Individual authentication method examples

**Coming Soon**:
- Basic authentication (username/password)
- API key authentication
- OAuth2 provider integration
- Passwordless (magic link & OTP)
- Passkey/WebAuthn (FIDO2)
- User registration flow

**Best for**: Understanding specific auth methods

> For now, see credential configuration in `/complete/core/02_credential_config/`

### 🎫 `/token` - Layer 2: Token Management

**Purpose**: Token generation, verification, and storage

**Examples**:
- `01_jwt` - JWT token management
- `02_simple` - Simple token generation
- `03_store` - Token storage patterns

**Best for**: Token lifecycle management

### 👤 `/subject` - Layer 3: Subject Resolution

**Purpose**: Identity context building and enrichment

**Examples**:
- `01_simple` - Basic subject resolution
- `02_enriched` - Enriched identity context
- `03_cached_store` - Caching strategies

**Best for**: Identity management patterns

### 🔒 `/authz` - Layer 4: Authorization

**Purpose**: Permission and policy evaluation

**Examples**:
- `01_rbac` - Role-Based Access Control
- `02_abac` - Attribute-Based Access Control
- `03_acl` - Access Control Lists

**Best for**: Authorization strategies

### 🌐 `/middleware` & `/services`

**Status**: 🚧 Being updated for new architecture

**Coming Soon**:
- HTTP middleware for authentication
- Multi-tenant management service examples
- Integration patterns

**Best for**: Production integration patterns

## 📖 Learning Path

### Path 1: Quick Start (Recommended)

1. **Deployment** → `01_deployment` - ⭐ START HERE

**Learn**:
- How `lokstra.Bootstrap()` works
- How `@RouterService` annotations auto-register services  
- Deployment modes: monolith vs microservices vs development
- Configuration with deployment.yaml

### Path 2: Understanding the Architecture

1. **Framework Code** → Read `/core`, `/credential`, etc.
2. **Documentation** → See `/docs` folder
3. **Deployment** → `01_deployment`

**Topics to explore**:
- Layered architecture (domain/application/infrastructure)
- `@RouterService` annotation pattern
- Dependency injection with `@Inject`
- Multi-tenant data isolation

> New examples coming soon to demonstrate each layer

### Path 3: Production Deployment

**For understanding deployment strategies**:
1. `01_deployment` - See all 3 deployment modes
2. Read `docs/deployment.md` - Production deployment guide
3. Study `config/deployment.yaml` - Configuration structure

**Deployment modes available**:
- **Monolith** - Single server, all services (good for small-medium apps)
- **Microservices** - Separate servers per layer (good for large-scale apps)
- **Development** - Debug mode with detailed logging

> Examples for specific use cases (SaaS, API services, enterprise) coming soon.

## 🏃 Running Examples

All examples can be run directly:

```bash
# Navigate to example directory
cd examples/complete/01_basic_flow

# Run the example
go run main.go
```

Or from project root:

```bash
go run ./examples/complete/01_basic_flow/main.go
```

## 📋 Prerequisites

```bash
# Install dependencies
go mod download

# Verify build
go build ./examples/...
```

## 💡 Tips

### In-Memory vs Production

All examples use **in-memory storage** by default:
- ✅ No database setup required
- ✅ Perfect for learning
- ✅ Fast to run

For production, replace with real implementations:
```go
// Example: In-memory (development)
store := repository.NewInMemoryUserStore()

// Production: Database
store := repository.NewPostgresUserStore(db)
```

### Environment Variables

Some examples support configuration via environment:
```bash
export JWT_SECRET="your-secret-key"
export SERVER=development
go run main.go
```

### Error Handling

Examples show basic error handling. For production:
- Add proper logging
- Implement retry logic
- Add monitoring/metrics
- Use structured errors

## 🔧 Troubleshooting

### Import Errors

```bash
# Fix module dependencies
go mod tidy

# Download missing packages
go mod download
```

### Build Errors

```bash
# Clean build cache
go clean -cache

# Rebuild
go build ./examples/...
```

### Example Not Found

Make sure you're in the project root:
```bash
cd /path/to/lokstra-auth
ls examples/  # Should show folders
```

## 🆕 Adding Your Own Example

Create a new example in the appropriate folder:

```bash
mkdir -p examples/complete/my-example
cd examples/complete/my-example
```

Create `main.go`:
```go
package main

import (
    "fmt"
    coredomain "github.com/primadi/lokstra-auth/core/domain"
)

func main() {
    fmt.Println("=== My Example ===")
    
    // Your code here
    
    fmt.Println("=== Complete ===")
}
```

Run it:
```bash
go run ./examples/complete/my-example/main.go
```

## 📚 Related Documentation

- [Deployment Guide](../docs/deployment.md)
- [Credential Providers](../docs/credential_providers.md)
- [Configuration Management](../docs/credential_configuration.md)
- [Multi-Tenant Architecture](../docs/multi_tenant_architecture.md)

## 🎯 Example Index

| Example | Category | Layer | Complexity | Status |
|---------|----------|-------|------------|--------|
| 01_deployment | Deployment | - | ⭐⭐⭐ | ✅ |
| token/* | Token | 2 | ⭐⭐ | ✅ |
| rbac/* | Subject | 3 | ⭐⭐ | ✅ |
| authz/* | Authz | 4 | ⭐⭐ | ✅ |

Legend:
- ✅ Working examples
- ⭐ Complexity (1-5)

> **Note**: Token, Subject, and Authz examples exist but may need updates to match the new annotation-based pattern. Credential layer examples are being redesigned.

---

**Need help?** Start with `/complete/` examples - they're designed to be self-explanatory!
