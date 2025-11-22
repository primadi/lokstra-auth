# API Key Security Implementation

## Overview

This document describes the security implementation for API key generation and storage in the `AppKeyService`.

## Security Features

### 1. Cryptographically Secure Secret Generation

**Problem:** Using simple ID generation (like `utils.GenerateID("secret")`) is not cryptographically secure.

**Solution:** Use `crypto/rand` for cryptographically secure random bytes:

```go
// Generate 32 bytes (256 bits) of cryptographically secure random data
secret, err := generateSecureSecret(32)
if err != nil {
    return nil, fmt.Errorf("failed to generate secure secret: %w", err)
}
```

**Implementation Details:**
- Uses `crypto/rand.Read()` for random byte generation
- Base64url encoding for URL-safe representation
- No padding characters (RFC 4648)
- 256-bit security level (32 bytes)

### 2. SHA3-256 Secret Hashing

**Problem:** Storing secrets in plain text allows compromise if database is breached.

**Solution:** Hash secrets using SHA3-256 before storage:

```go
// Hash the secret using SHA3-256
secretHash, err := hashSecret(secret)
if err != nil {
    return nil, fmt.Errorf("failed to hash secret: %w", err)
}

// Store ONLY the hash
apiKey.SecretHash = secretHash  // ✅ Hashed
```

**Why SHA3-256?**
- Latest NIST-approved cryptographic hash function
- Resistant to length-extension attacks (unlike SHA-2)
- More secure than SHA-256 for this use case
- Industry standard for secret hashing

**Alternative:** SHA-256 fallback available via `hashSecretSHA256()` for compatibility.

### 3. Secret Exposure Prevention

**Problem:** Returning full `APIKey` with `SecretHash` field exposes sensitive data.

**Solution:** Use sanitized DTOs that exclude secret hash:

```go
// ❌ NEVER return this in Get/List operations
type APIKey struct {
    SecretHash string  // Contains hashed secret
    // ...other fields
}

// ✅ Return this instead
type AppKeyInfo struct {
    KeyID      string  // Public identifier only
    // ...other safe fields
    // ❌ NO SecretHash field!
}
```

**Implementation:**
- `GenerateKey()` → Returns `AppKeyResponse` with plain secret (ONLY once!)
- `GetKey()` → Returns `AppKeyInfo` (no secret hash)
- `ListKeys()` → Returns `[]AppKeyInfo` (no secret hashes)

### 4. Constant-Time Comparison

**Problem:** String comparison can leak information via timing attacks.

**Solution:** Use `subtle.ConstantTimeCompare()` for verification:

```go
func verifySecret(secret, hash string) (bool, error) {
    computedHash, err := hashSecret(secret)
    if err != nil {
        return false, err
    }
    
    // Constant-time comparison prevents timing attacks
    return subtle.ConstantTimeCompare(
        []byte(computedHash),
        []byte(hash),
    ) == 1, nil
}
```

## Security Best Practices

### For API Key Generation

1. **Use Secure Random Generation:**
   ```go
   secret, _ := generateSecureSecret(32)  // ✅ 256-bit security
   ```

2. **Always Hash Before Storage:**
   ```go
   apiKey.SecretHash = hashSecret(secret)  // ✅ SHA3-256
   ```

3. **Return Secret ONLY Once:**
   ```go
   // ✅ Only in GenerateKey response
   return &AppKeyResponse{
       KeyString: keyString,  // Contains plain secret
       AppKey:    apiKey,
   }
   ```

### For API Key Retrieval

1. **Never Expose Secret Hash:**
   ```go
   // ❌ BAD - exposes hash
   return apiKey
   
   // ✅ GOOD - sanitized
   return domain.ToAppKeyInfo(apiKey)
   ```

2. **Use Sanitized DTOs:**
   ```go
   func GetKey() (*AppKeyInfo, error) {  // ✅ Sanitized type
       // ...
   }
   ```

### For Secret Verification

1. **Use Constant-Time Comparison:**
   ```go
   isValid, _ := verifySecret(providedSecret, storedHash)
   ```

2. **Never Log Secrets:**
   ```go
   log.Info("generated key", "keyID", keyID)  // ✅ Log ID only
   log.Error("invalid secret")                // ❌ Never log secret!
   ```

## Key String Format

API keys follow this format:

```
{app_id}_{key_id}.{secret}
```

Example:
```
app_abc123_appkey_xyz789.Kx7mP2qR8vN3wY9zA5bC1dE4fG6hJ0
```

**Components:**
- `app_id`: Application identifier (public)
- `key_id`: Key identifier (public)
- `secret`: Cryptographically secure random string (private)

**Usage:**
- Full key string shown ONLY when generated
- Subsequent operations use `key_id` only
- Secret verified via constant-time comparison with stored hash

## Migration Guide

If you have existing keys with plain text secrets:

1. **Generate new keys** using secure crypto
2. **Notify users** to rotate their keys
3. **Deprecate old keys** after grace period
4. **Delete old keys** once all clients have migrated

**DO NOT** attempt to hash existing plain text secrets - they are already compromised if stored in plain text.

## Security Checklist

- [x] Use `crypto/rand` for secret generation
- [x] Hash secrets with SHA3-256 before storage
- [x] Never return secret hash in API responses
- [x] Use constant-time comparison for verification
- [x] Never log or expose secrets
- [x] Return plain secret ONLY once (at generation)
- [x] Use sanitized DTOs for Get/List operations
- [x] Implement key rotation capability
- [x] Set appropriate key expiration times

## References

- [NIST SHA-3 Standard](https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.202.pdf)
- [Go crypto/rand](https://pkg.go.dev/crypto/rand)
- [Go subtle package](https://pkg.go.dev/crypto/subtle)
- [RFC 4648 - Base64 Encoding](https://tools.ietf.org/html/rfc4648)
