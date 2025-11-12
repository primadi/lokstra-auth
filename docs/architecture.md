# Lokstra Auth - Architecture Overview

## Introduction

Lokstra Auth is an authentication and authorization framework built on top of [Lokstra Framework](https://github.com/primadi/lokstra). The framework is designed with a modular 4-layer architecture that provides flexibility and composability.

## 4-Layer Architecture

### Layer 1: Credential Input / Login Flow
**Folder**: `/01_credential`

This layer is responsible for:
- Receiving and validating credentials from various sources (login forms, APIs, OAuth, etc.)
- Managing different authentication methods (username/password, tokens, biometric, etc.)
- Implementing different login flows based on application requirements

**Key Components**:
- Input validators
- Authentication providers
- Login flow handlers

**Implementations**:
- `basic/` - Username/password authentication
- `oauth2/` - OAuth2 flow
- `apikey/` - API key validation
- `passwordless/` - Email/SMS OTP, magic links
- `passkey/` - WebAuthn/FIDO2 passkey
- `mtls/` - Mutual TLS (optional)

### Layer 2: Token Verification / Claim Extraction
**Folder**: `/02_token`

This layer is responsible for:
- Token verification (JWT, OAuth tokens, custom tokens)
- Claim extraction from tokens
- Token lifecycle management (generation, refresh, revocation)

**Key Components**:
- Token verifiers
- Claim extractors
- Token generators

**Implementations**:
- `jwt/` - JWT token handling
- `opaque/` - Opaque token handling
- `refresh/` - Refresh token mechanisms

### Layer 3: Subject Resolution / Identity Context
**Folder**: `/03_subject`

This layer is responsible for:
- Resolving subject/user based on claims
- Building complete identity context
- Enriching user data from various sources

**Key Components**:
- Subject resolvers
- Identity context builders
- User data enrichers

**Implementations**:
- `simple/` - Simple subject resolver
- `enriched/` - Enriched with external data sources
- `cached/` - With caching layer for performance

### Layer 4: Authorization / Policy Evaluation
**Folder**: `/04_authz`

This layer is responsible for:
- Access policy evaluation (RBAC, ABAC, etc.)
- Permission checks
- Resource-level authorization

**Key Components**:
- Policy evaluators
- Permission checkers
- Access control managers

**Implementations**:
- `rbac/` - Role-Based Access Control
- `abac/` - Attribute-Based Access Control
- `policy/` - Policy-based authorization

## Project Structure

Each layer follows a consistent structure:

```
/0X_layername/
  ├── contract.go          # Interface definitions (contracts)
  ├── implementation1/     # First implementation
  ├── implementation2/     # Second implementation
  └── ...                  # Additional implementations
```

**Contract vs Implementation**:
- **Contracts**: All interfaces are defined in `contract.go` at the root of each layer
- **Implementations**: Multiple implementations of the same contract can exist in separate folders
- **Flexibility**: This allows users to choose or create custom implementations while maintaining the same interface

## Design Principles

1. **Modularity**: Each layer is independent and can be used separately
2. **Extensibility**: Easy to add new providers or strategies
3. **Composability**: Layers can be combined based on requirements
4. **Type Safety**: Leverages Go generics for type-safe operations
5. **Lokstra Integration**: Built on top of Lokstra Framework for dependency injection and lifecycle management
6. **Contract-First**: Clear separation between interfaces and implementations

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    Request with Credentials                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: Credential Input / Login Flow (01_credential)      │
│  - Validate credentials                                      │
│  - Execute authentication flow                               │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: Token Verification / Claim Extraction (02_token)   │
│  - Verify token signature                                    │
│  - Extract claims                                            │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: Subject Resolution / Identity Context (03_subject) │
│  - Resolve user/subject                                      │
│  - Build complete identity context                           │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 4: Authorization / Policy Evaluation (04_authz)       │
│  - Check permissions                                         │
│  - Evaluate access policies                                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Allow/Deny Access to Resource                   │
└─────────────────────────────────────────────────────────────┘
```

## Use Cases

### Full Stack Authentication
Use all 4 layers for a complete authentication and authorization solution.

### API Gateway Authentication
Focus on layers 2 and 4 for token verification and authorization.

### Custom Login Flow
Use layers 1 and 2 for custom authentication implementation.

### Policy-Based Access Control
Focus on layers 3 and 4 for advanced authorization scenarios.

## Getting Started

### Quick Start with Auth Runtime

The recommended way to use Lokstra Auth is through the `Auth` runtime:

```go
auth := lokstraauth.NewBuilder().
    WithAuthenticator("basic", authenticator).
    WithTokenManager(tokenManager).
    WithSubjectResolver(resolver).
    WithIdentityContextBuilder(builder).
    WithAuthorizer(authorizer).
    Build()

// Login, verify, authorize...
```

See [Runtime Documentation](./runtime.md) for detailed information.

### Using Individual Layers

For advanced use cases, you can use individual layers directly.

See the `/examples` folder for usage examples of each layer and complete integrations.

## Documentation

- **[Runtime API](./runtime.md)** - Main entry point (Recommended)
- [Layer 1: Credential Input](./01_credential.md)
- [Layer 2: Token Verification](./02_token.md)
- [Layer 3: Subject Resolution](./03_subject.md)
- [Layer 4: Authorization](./04_authz.md)
