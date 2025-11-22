# Complete Examples

This directory contains complete, real-world examples demonstrating how to use all layers of Lokstra Auth together in production-like scenarios.

## Overview

These examples show the full authentication and authorization flow, from credential validation through token management, subject resolution, to final authorization decisions.

## Quick Start Guide

**Start here if you're new to Lokstra:**

1. **[Deployment](./00_deployment/)** - ⭐ How to run the framework - START HERE

This is currently the main example demonstrating:
- `lokstra.Bootstrap()` auto-registration
- `@RouterService` annotation pattern
- Deployment modes (monolith/microservices/development)
- Configuration with deployment.yaml

> **Note**: Additional complete examples are being redesigned for the annotation-based architecture.  
> The framework now uses automatic service registration instead of manual instantiation.

## Examples

### 0. [Deployment](./00_deployment/) - Framework Deployment Modes

**Demonstrates:**
- Bootstrap process with auto-registration
- Monolith deployment (single server, port 8080)
- Microservices deployment (5 separate services)
- Development mode (debug, port 3000)
- Configuration management with deployment.yaml

**Use Case:** Understanding how to deploy and run Lokstra framework

**Run:**
```bash
cd 00_deployment

# Monolith mode (default)
go run main.go

# Microservices mode
SERVER=microservices go run main.go

# Development mode
SERVER=development go run main.go
```

**What you'll learn:**
1. How Bootstrap() generates service registry
2. How to configure deployment modes
3. How to run services in different architectures
4. How to use environment variables for configuration

---

### 1. [Basic Flow](./01_basic_flow/) - Complete 4-Layer Authentication & Authorization

**Demonstrates:**
- Basic username/password authentication
- JWT token generation and verification
- Subject resolution with identity context building
- RBAC authorization with role-based permissions

**Use Case:** Standard web application login flow

**Run:**
```bash
cd 01_basic_flow
go run main.go
```

**What you'll see:**
1. User authentication with username/password
2. JWT token generation
3. Token verification
4. Subject resolution from claims
5. Identity context building with roles, permissions, groups
6. Authorization checks for different resources

---

### 2. [Multi-Authentication](./02_multi_auth/) - Multiple Credential Types

**Demonstrates:**
- **Basic Auth** (Username/Password)
- **Passwordless Auth** (Magic Link & OTP)
- JWT token management
- Subject resolution
- RBAC authorization
- Different users with different permissions

**Use Case:** Modern application supporting multiple authentication methods

**Run:**
```bash
cd 02_multi_auth
go run main.go
```

**What you'll see:**

**Scenario 1: Admin with Basic Auth**
- Login with username/password
- Full admin permissions granted
- All authorization checks pass

**Scenario 2: Passwordless Magic Link**
- Magic link initiation for email
- Mock email sender shows token and link
- Demonstrates passwordless flow (initiation only)

**Scenario 3: Passwordless OTP**
- OTP generation for email
- Mock SMS/email with 6-digit code
- Demonstrates OTP flow (initiation only)

**Scenario 4: Developer with Basic Auth**
- Login with limited permissions
- Role-based authorization
- Some actions denied based on role

**Key Features:**
- ✅ Multiple authentication methods in one app
- ✅ Mock implementations for email/SMS sending
- ✅ Different users with different permission levels
- ✅ Comprehensive RBAC testing

---

## Architecture

All examples follow the 4-layer Lokstra Auth architecture:

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 4: Authorization (RBAC, ABAC, ACL, Policy)            │
├─────────────────────────────────────────────────────────────┤
│ Layer 3: Subject Resolution (Simple, Enriched, Cached)      │
├─────────────────────────────────────────────────────────────┤
│ Layer 2: Token Management (JWT with Access + Refresh)       │
├─────────────────────────────────────────────────────────────┤
│ Layer 1: Credentials (Basic, OAuth2, Passwordless, etc.)    │
└─────────────────────────────────────────────────────────────┘
```

## Common Patterns

### 1. Authentication Flow
```go
// Authenticate user
authResult, err := authenticator.Authenticate(ctx, credentials)

// Generate JWT
token, err := tokenManager.Generate(ctx, authResult.Claims)

// Verify token
verifyResult, err := tokenManager.Verify(ctx, token.Value)

// Resolve subject
subject, err := resolver.Resolve(ctx, verifyResult.Claims)

// Build identity context
identity, err := contextBuilder.Build(ctx, subject)
```

### 2. Authorization Check
```go
// Create authorization request
request := &authz.AuthorizationRequest{
    Subject:  identity,
    Resource: &authz.Resource{Type: "posts", ID: "123"},
    Action:   authz.ActionRead,
}

// Evaluate
decision, err := evaluator.Evaluate(ctx, request)
if decision.Allowed {
    // Grant access
}
```

## Best Practices

### 1. **Use Strong Passwords**
```go
// Bad
password := "123456"

// Good
password := "MySecure@Password123"
```

The Basic Auth validator enforces:
- Minimum 8 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character

### 2. **Secure JWT Secret**
```go
// Bad (hardcoded in code)
secret := "my-secret-key"

// Good (from environment)
secret := os.Getenv("JWT_SECRET")
if secret == "" {
    log.Fatal("JWT_SECRET not set")
}
```

### 3. **Token Expiry**
```go
config := jwt.DefaultConfig(secret)
// Access tokens expire quickly
// Refresh tokens last longer for renewal
```

### 4. **Mock Implementations in Production**
```go
// Development
emailSender := &MockEmailSender{}

// Production
emailSender := smtp.NewEmailSender(smtpConfig)
```

### 5. **Error Handling**
```go
// Always check errors
authResult, err := authenticator.Authenticate(ctx, creds)
if err != nil {
    return fmt.Errorf("authentication failed: %w", err)
}

if !authResult.Success {
    return fmt.Errorf("invalid credentials: %s", authResult.Error)
}
```

## Testing

Each example can be run standalone:

```bash
# Test basic flow
cd 01_basic_flow && go run main.go

# Test multi-authentication
cd 02_multi_auth && go run main.go
```

## Integration Guide

### Adding to Your Application

1. **Choose Authentication Methods**
   - Basic (username/password)
   - Passwordless (magic link, OTP)
   - OAuth2 (Google, GitHub, etc.)
   - API Key
   - Passkey (WebAuthn)

2. **Configure Token Management**
   ```go
   jwtConfig := jwt.DefaultConfig(os.Getenv("JWT_SECRET"))
   tokenManager := jwt.NewManager(jwtConfig)
   ```

3. **Set Up Subject Resolution**
   ```go
   resolver := simple.NewResolver()
   contextBuilder := simple.NewContextBuilder(
       roleProvider,
       permProvider,
       groupProvider,
       profileProvider,
   )
   ```

4. **Choose Authorization Strategy**
   - **RBAC**: Role-based (admin, user, guest)
   - **ABAC**: Attribute-based (department, time, location)
   - **ACL**: Resource-level (per-document permissions)
   - **Policy**: Complex rules with combining algorithms

## Common Issues

### 1. Password Too Weak
```
Error: password does not meet complexity requirements
```
**Solution:** Use password with uppercase, lowercase, number, and special char (min 8 chars)

### 2. Token Expired
```
Error: token has expired
```
**Solution:** Refresh token or re-authenticate

### 3. Permission Denied
```
Decision: Allowed=false, Reason="no matching role permissions found"
```
**Solution:** Check role permissions mapping and permission format (`action:resource`)

## Next Steps

1. **Explore Individual Layers:**
   - [01_credential examples](../01_credential/) - Authentication methods
   - [02_token examples](../02_token/) - Token management
   - [03_subject examples](../03_subject/) - Subject resolution
   - [04_authz examples](../04_authz/) - Authorization strategies

2. **Build Your Own:**
   - Copy an example as starting point
   - Customize for your use case
   - Add your own authentication providers
   - Implement custom authorization rules

3. **Production Deployment:**
   - Replace mock implementations
   - Add database persistence
   - Implement proper error handling
   - Add logging and monitoring
   - Use environment variables for secrets

## Support

For more examples and documentation:
- Main README: [../../README.md](../../README.md)
- Layer 1 Examples: [../01_credential/](../01_credential/)
- Layer 2 Examples: [../02_token/](../02_token/)
- Layer 3 Examples: [../03_subject/](../03_subject/)
- Layer 4 Examples: [../04_authz/](../04_authz/)
