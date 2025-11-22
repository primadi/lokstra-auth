# Layer 1: Credential Input / Login Flow

## Overview

The Credential Input layer is the first layer in Lokstra's authentication architecture. It handles receiving, validating, and processing various types of credentials from users.

## Core Philosophy

This layer follows a **contract-first pattern** with a **minimal general interface**:
- Only 3 core interfaces: `Credentials`, `Authenticator`, and `AuthenticationResult`
- Implementation-specific contracts are defined in their respective packages
- All authenticators implement the same `Authenticator` interface for consistency

## Core Interfaces

### 1. Credentials Interface

Represents user credentials input for any authentication method.

```go
type Credentials interface {
    Type() string     // Returns the credential type (e.g., "basic", "oauth2")
    Validate() error  // Validates credential format
}
```

### 2. Authenticator Interface

**The only general contract** - all authentication implementations must implement this.

```go
type Authenticator interface {
    Authenticate(ctx context.Context, creds Credentials) (*AuthenticationResult, error)
    Type() string // Must match Credentials.Type()
}
```

### 3. AuthenticationResult Struct

Standardized result returned by all authenticators.

```go
type AuthenticationResult struct {
    Success  bool               // Whether authentication succeeded
    Subject  string             // Authenticated user ID
    Claims   map[string]any     // Additional claims for token
    Metadata map[string]any     // Authentication metadata
    Error    error              // Error if authentication failed
}
```

## Implementation-Specific Contracts

Each authenticator may define its own specific contracts:

### Basic Authenticator
- **`CredentialValidator`** - Validates username/password complexity requirements
- **`UserProvider`** - Interface for retrieving user data from storage

Located in: `01_credential/basic/contract.go`

### Passkey Authenticator
- **`CredentialStore`** - Manages WebAuthn credentials and user data

Located in: `01_credential/passkey/store.go`

### Passwordless Authenticator
- **`TokenStore`** - Manages OTP and Magic Link tokens
- **`UserResolver`** - Resolves user ID from email address
- **`TokenSender`** - Sends tokens via email or SMS

### API Key Authenticator
- **`KeyStore`** - Stores and validates API keys

## Credential Types

The layer supports multiple credential types:

- **Username/Password** - Traditional authentication (`basic`)
- **OAuth2** - Third-party provider authentication (`oauth2`)
- **API Keys** - Service-to-service authentication (`apikey`)
- **Passwordless** - OTP and Magic Link (`passwordless`)
- **Passkey** - WebAuthn/FIDO2 biometric/security key (`passkey`)
- **Custom** - Extensible for custom credential types

## Components

### Core Components

1. **Authenticators** - Verify credentials and return authentication results
2. **Credentials** - Type-safe credential input structures
3. **Validators** - Validate credential format and requirements (implementation-specific)
4. **Stores** - Persist and retrieve authentication data (implementation-specific)

### Multi-Authenticator Support

The layer supports multiple authentication methods simultaneously through `MultiAuthenticator`:

```go
multi := credential.NewMultiAuthenticator()
multi.AddAuthenticator(basicAuth)
multi.AddAuthenticator(oauth2Auth)
multi.AddAuthenticator(apikeyAuth)

// Automatically routes to the correct authenticator based on credential type
result, err := multi.Authenticate(ctx, credentials)
```

## Implementations

### Basic (`01_credential/basic/`)

Traditional username/password authentication with configurable password policies.

**Key Features:**
- Bcrypt password hashing
- Username or email login
- Password complexity validation via `CredentialValidator`
- Extensible `UserProvider` interface

**Specific Contracts:**
- `CredentialValidator` - Validates password complexity
- `UserProvider` - Retrieves user data

### OAuth2 (`01_credential/oauth2/`)

OAuth2 flow implementation supporting multiple providers.

**Supported Providers:**
- Google (OpenID Connect)
- GitHub
- Facebook

**Key Features:**
- Authorization code flow
- Token validation with providers
- User info fetching
- Email verification

### API Key (`01_credential/apikey/`)

API key validation for service-to-service authentication.

**Key Features:**
- SHA3-256 key hashing
- Key expiration support
- Scope-based permissions
- Usage tracking
- Key revocation

**Specific Contracts:**
- `KeyStore` - Stores and validates API keys

### Passwordless (`01_credential/passwordless/`)

Email-based OTP and magic link authentication flows.

**Authentication Methods:**
- Magic Link (15-minute expiry)
- OTP - 6-digit code (5-minute expiry)

**Key Features:**
- Token lifecycle management
- One-time use enforcement
- Configurable token generation
- Email sender abstraction

**Specific Contracts:**
- `TokenStore` - Manages token lifecycle
- `UserResolver` - Resolves user ID from email
- `TokenGenerator` - Generates random tokens
- `TokenSender` - Sends tokens via email

### Passkey (`01_credential/passkey/`)

WebAuthn/FIDO2 implementation for passwordless authentication using biometrics or security keys.

**Key Features:**
- WebAuthn protocol support
- Resident key support
- User verification levels
- Challenge-response authentication
- Registration and login ceremonies

**Specific Contracts:**
- `CredentialStore` - Manages WebAuthn credentials and users

## Architecture

```
01_credential/
├── contract.go              # Core general interfaces (Credentials, Authenticator, AuthenticationResult)
├── basic/
│   ├── contract.go          # Basic-specific: CredentialValidator
│   ├── authenticator.go
│   ├── credential.go
│   └── provider_inmemory.go
├── oauth2/
│   ├── authenticator.go
│   └── credential.go
├── apikey/
│   ├── authenticator.go
│   └── store.go             # KeyStore interface + implementation
├── passwordless/
│   ├── authenticator.go
│   └── store.go             # TokenStore interface + implementation
└── passkey/
    ├── authenticator.go
    └── store.go             # CredentialStore interface + implementation (passkey-specific)
```

## Contract Philosophy

The credential layer maintains a **clean separation of concerns**:

### General Contracts (`01_credential/contract.go`)
Contains only 3 interfaces that ALL authenticators must implement:
- `Credentials` - Generic credential input
- `Authenticator` - Authentication logic
- `AuthenticationResult` - Standardized output

### Implementation-Specific Contracts
Each authenticator package defines its own specific interfaces:
- **Basic**: `CredentialValidator`, `UserProvider`
- **Passkey**: `CredentialStore` (WebAuthn-specific)
- **Passwordless**: `TokenStore`, `UserResolver`, `TokenSender`
- **API Key**: `KeyStore`

This design ensures:
✅ Clean, focused general contract
✅ No unused interfaces in general contract
✅ Implementation-specific needs stay in their packages
✅ Easy to extend with new authenticators

## Usage Examples

See `/examples/01_credential/` folder for complete implementation examples:

- `01_basic/` - Username/password authentication
- `02_multi_auth/` - Multiple authentication methods
- `03_oauth2/` - OAuth2 provider integration
- `04_passwordless/` - Magic link and OTP
- `05_apikey/` - API key authentication
- `06_passkey/` - WebAuthn/FIDO2 authentication
