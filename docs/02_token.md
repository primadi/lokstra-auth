# Layer 2: Token Verification / Claim Extraction

## Overview

The Token Verification layer is responsible for token generation, verification, and lifecycle management.

## Concepts

### Token Types
- JWT (JSON Web Tokens)
- Opaque tokens
- Refresh tokens
- Custom tokens

### Claims
Data contained within tokens (user ID, roles, permissions, metadata).

### Token Lifecycle
- **Generation**: Creating new tokens
- **Verification**: Verifying token validity
- **Refresh**: Renewing expired tokens
- **Revocation**: Revoking active tokens

## Components

### Token Generators
Components for creating tokens with specific claims.

### Token Verifiers
Components for verifying token signature and validity.

### Claim Extractors
Components for extracting information from tokens.

## Implementations

### JWT (`/jwt`)
JSON Web Token implementation with support for various signing algorithms (HS256, RS256, ES256).

### Opaque (`/opaque`)
Opaque token handling with server-side storage and validation.

### Refresh (`/refresh`)
Refresh token mechanisms for token rotation and renewal.

## Contract

All implementations must adhere to the contracts defined in `contract.go`:

```go
// See /token/contract.go for interface definitions
```

## Usage Examples

See `/examples/token` folder for implementation examples.

## API Reference

(To be documented)
