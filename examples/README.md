# Examples

This folder contains usage examples for the Lokstra Auth framework.

## Structure

### `/01_credential`
Examples for Layer 1 - Credential Input/Login Flow:
- Basic username/password authentication
- OAuth2 flow implementation
- Passwordless authentication
- Passkey/WebAuthn implementation
- Custom authentication methods

### `/02_token`
Examples for Layer 2 - Token Verification/Claim Extraction:
- JWT generation and verification
- Token refresh flow
- Opaque token handling
- Custom token formats

### `/03_subject`
Examples for Layer 3 - Subject Resolution/Identity Context:
- Subject resolution from claims
- Identity context enrichment
- Multi-source data loading
- Caching strategies

### `/04_authz`
Examples for Layer 4 - Authorization/Policy Evaluation:
- RBAC implementation
- ABAC with custom attributes
- Policy-based authorization
- Permission checking patterns

### `/complete`
Complete integration examples using all layers:
- End-to-end authentication flow
- Complete authorization pipeline
- Real-world use cases
- Best practices

## Running Examples

Each example can be run independently:

```bash
cd examples/01_credential
go run main.go
```

## Prerequisites

Make sure you have installed all dependencies:

```bash
go mod download
```
