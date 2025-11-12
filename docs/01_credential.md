# Layer 1: Credential Input / Login Flow

## Overview

The Credential Input layer is responsible for receiving, validating, and processing various types of credentials from users.

## Concepts

### Authentication Provider
Interface that defines how credentials are verified.

### Login Flow
Complete process from receiving credentials to generating tokens or sessions.

### Credential Types
- Username/Password
- Email/Password
- OAuth2 tokens
- API Keys
- Biometric data
- Passwordless (OTP, Magic Links)
- Passkey (WebAuthn/FIDO2)
- Custom credentials

## Components

### Validators
Components for validating credential format and integrity.

### Authenticators
Components for verifying credentials against data stores.

### Flow Handlers
Components for managing various login flows (standard, OAuth2, 2FA, etc.).

## Implementations

### Basic (`/basic`)
Traditional username/password authentication with configurable password policies.

### OAuth2 (`/oauth2`)
OAuth2 flow implementation supporting multiple providers (Google, GitHub, etc.).

### API Key (`/apikey`)
API key validation for service-to-service authentication.

### Passwordless (`/passwordless`)
Email/SMS OTP and magic link authentication flows.

### Passkey (`/passkey`)
WebAuthn/FIDO2 implementation for passwordless authentication using biometrics or security keys.

## Contract

All implementations must adhere to the contracts defined in `contract.go`:

```go
// See /01_credential/contract.go for interface definitions
```

## Usage Examples

See `/examples/01_credential` folder for implementation examples.

## API Reference

(To be documented)
