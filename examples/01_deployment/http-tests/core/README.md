# HTTP Tests for Lokstra Auth Core Services

This folder contains HTTP test files for testing all core services (core) of Lokstra Auth.

## Prerequisites

- Install REST Client extension for VS Code
- Ensure the server is running on `http://localhost:9090`
- Or modify `@baseUrl` variable in each file

## Test Files Overview

### 00-workflow.http
**Complete end-to-end workflow** for testing the entire system. This is the recommended starting point.

**Features:**
- Automated variable extraction from responses
- Step-by-step tenant setup from scratch
- All core services in correct order
- Verification checks included

**Workflow Steps:**
1. Create Tenant
2. Create App
3. Create Branch (optional - deprecated)
4. Create User
5. Generate App Key
6. Configure Credentials
7. Verification checks

**Usage:**
1. Open `00-workflow.http`
2. Execute requests sequentially from top to bottom
3. Variables will be automatically populated from previous responses

---

### 01-tenant-service.http
Tests for **multi-tenant management** operations.

**Endpoints:**
- `POST /api/auth/core/tenants` - Create tenant
- `GET /api/auth/core/tenants/id/{id}` - Get tenant by ID
- `GET /api/auth/core/tenants` - List all tenants
- `PUT /api/auth/core/tenants/id/{id}` - Update tenant
- `POST /api/auth/core/tenants/id/{id}/activate` - Activate tenant
- `POST /api/auth/core/tenants/id/{id}/suspend` - Suspend tenant
- `DELETE /api/auth/core/tenants/id/{id}` - Delete tenant

**Key Features:**
- Multi-database tenancy (db_dsn + db_schema)
- Tenant status management (active/suspended/deleted)
- Settings & metadata support
- Config merging for partial updates

---

### 02-app-service.http
Tests for **application management** within a tenant.

**Endpoints:**
- `POST /api/auth/core/tenants/{tenant_id}/apps` - Create app
- `GET /api/auth/core/tenants/{tenant_id}/apps/id/{id}` - Get app
- `GET /api/auth/core/tenants/{tenant_id}/apps` - List apps
- `PUT /api/auth/core/tenants/{tenant_id}/apps/id/{id}` - Update app
- `DELETE /api/auth/core/tenants/{tenant_id}/apps/id/{id}` - Delete app

**App Types:**
- `web` - Web applications
- `mobile` - Mobile apps
- `api` - API services
- `spa` - Single Page Apps

**Required Variables:**
- `@tenantId` - Must exist before testing

---

### 03-branch-service.http
Tests for **branch management** within an app.

> **Note:** Branch concept is being deprecated. Most operations now work directly at app level.

**Endpoints:**
- `POST /api/auth/core/tenants/{tenant_id}/apps/{app_id}/branches` - Create branch
- `GET /api/auth/core/tenants/{tenant_id}/apps/{app_id}/branches/id/{id}` - Get branch
- `GET /api/auth/core/tenants/{tenant_id}/apps/{app_id}/branches` - List branches
- `PUT /api/auth/core/tenants/{tenant_id}/apps/{app_id}/branches/id/{id}` - Update branch
- `DELETE /api/auth/core/tenants/{tenant_id}/apps/{app_id}/branches/id/{id}` - Delete branch

**Common Branches:**
- `production` - Production environment
- `staging` - Staging environment
- `development` - Development environment

**Required Variables:**
- `@tenantId`
- `@appId`

---

### 04-user-service.http
Tests for **user management** within a tenant.

**Endpoints:**
- `POST /api/auth/core/tenants/{tenant_id}/users` - Create user
- `GET /api/auth/core/tenants/{tenant_id}/users/id/{id}` - Get user by ID
- `GET /api/auth/core/tenants/{tenant_id}/users/by-username/{username}` - Get by username
- `GET /api/auth/core/tenants/{tenant_id}/users/by-email/{email}` - Get by email
- `GET /api/auth/core/tenants/{tenant_id}/users` - List all users
- `PUT /api/auth/core/tenants/{tenant_id}/users/id/{id}` - Update user
- `PUT /api/auth/core/tenants/{tenant_id}/users/id/{id}/password` - Update password
- `POST /api/auth/core/tenants/{tenant_id}/users/id/{id}/enable` - Enable user
- `POST /api/auth/core/tenants/{tenant_id}/users/id/{id}/disable` - Disable user
- `DELETE /api/auth/core/tenants/{tenant_id}/users/id/{id}` - Delete user

**User Roles:**
- `user` - Regular user
- `admin` - Administrator
- `superadmin` - Super administrator

**User Status:**
- `active` - Can authenticate
- `disabled` - Cannot authenticate
- `deleted` - Soft deleted

**Required Variables:**
- `@tenantId`

---

### 05-app-key-service.http
Tests for **API key management** (service-to-service authentication).

**Endpoints:**
- `POST /api/auth/core/tenants/{tenant_id}/apps/{app_id}/keys` - Generate key
- `GET /api/auth/core/tenants/{tenant_id}/apps/{app_id}/keys/{key_id}` - Get key
- `GET /api/auth/core/tenants/{tenant_id}/apps/{app_id}/keys` - List keys
- `POST /api/auth/core/tenants/{tenant_id}/apps/{app_id}/keys/{key_id}/revoke` - Revoke key
- `POST /api/auth/core/tenants/{tenant_id}/apps/{app_id}/keys/{key_id}/rotate` - Rotate key
- `DELETE /api/auth/core/tenants/{tenant_id}/apps/{app_id}/keys/{key_id}` - Delete key

**Key Features:**
- Secure key generation with SHA3-256 hashing
- Key types: `service`, `user`, `integration`
- Environment-specific keys: `production`, `staging`, `development`
- Key rotation support
- Expiration management
- Scopes & metadata support

**Security:**
- Secret only shown once during generation
- Keys stored as hashed values
- Revoked keys cannot be rotated
- Multi-tenant isolation enforced

**Workflow Example:**
1. Generate key for service-to-service auth
2. Store secret securely (shown once!)
3. Use key for API authentication
4. Rotate periodically for security
5. Revoke when compromised

**Required Variables:**
- `@tenantId`
- `@appId`

---

### 06-credential-config-service.http
Tests for **credential provider configuration** management.

**Endpoints:**

**Tenant-Level (defaults for all apps):**
- `GET /api/auth/core/config/credentials/tenants/{tenant_id}` - Get tenant defaults
- `PUT /api/auth/core/config/credentials/tenants/{tenant_id}` - Update tenant defaults

**App-Level (override tenant defaults):**
- `GET /api/auth/core/config/credentials/tenants/{tenant_id}/apps/{app_id}` - Get app config
- `PUT /api/auth/core/config/credentials/tenants/{tenant_id}/apps/{app_id}` - Update app config

**Supported Credential Types:**

1. **Basic Auth** (Username/Password)
   - Password strength requirements
   - Login attempt limits & lockout
   - Session timeout

2. **OAuth2** (Social Login)
   - Multiple providers: Google, Azure, GitHub, Facebook
   - Custom OAuth2 providers
   - State management & callbacks

3. **API Key** (Service Auth)
   - Key generation settings
   - Hash algorithm (sha3-256/sha256)
   - Expiration policies
   - Rate limiting

4. **Passwordless** (Magic Link/OTP)
   - Email/SMS delivery
   - Code/link expiration
   - Attempt limits & cooldown

5. **Passkey** (WebAuthn/FIDO2)
   - Relying party configuration
   - User verification levels
   - Attestation preferences

**Configuration Hierarchy:**
```
Global Defaults
    ↓
Tenant Default Config (applies to all apps)
    ↓
App-Specific Config (overrides tenant defaults)
```

**Smart Merge Behavior:**
- Only updates credential types when config details provided
- Preserves existing configs not mentioned in request
- Enables partial updates without losing data

**Workflow Examples:**

1. **Initial Tenant Setup:**
   - Set tenant defaults (basic + apikey)
   - All new apps inherit these settings

2. **Add OAuth2 to Existing Tenant:**
   - Update tenant config with OAuth2 providers
   - Existing basic/apikey configs preserved

3. **Production App with Maximum Security:**
   - Override tenant defaults for specific app
   - Stricter password policy
   - Shorter expiration
   - Required passkey authentication

**Required Variables:**
- `@tenantId` (for tenant-level config)
- `@tenantId` + `@appId` (for app-level config)

---

## Quick Start Guide

### Option 1: Complete Workflow (Recommended)
```
1. Open 00-workflow.http
2. Click "Send Request" for each step sequentially
3. Variables will auto-populate from responses
4. Verify at the end
```

### Option 2: Individual Service Testing
```
1. Start with 01-tenant-service.http
2. Copy the tenant ID from response
3. Set @tenantId in subsequent files
4. Continue with 02-app-service.http
5. Repeat for each service
```

## Common Variables

All files use these common variables:

```http
@baseUrl = http://localhost:9090
@contentType = application/json
@tenantId = acme-corp
@appId = main-app
@branchId = hq-jakarta
@userId = user_xxx
@appKeyId = appkey_xxx
```

## Route Patterns

All services follow consistent route patterns:

```http
# ID-based access
GET /api/auth/core/tenants/id/{id}
GET /api/auth/core/tenants/{tenant_id}/users/id/{id}

# Query by specific field
GET /api/auth/core/tenants/{tenant_id}/users/by-email/{email}
GET /api/auth/core/tenants/{tenant_id}/users/by-username/{username}

# List/Collection
GET /api/auth/core/tenants
GET /api/auth/core/tenants/{tenant_id}/users

# Actions (POST for state changes)
POST /api/auth/core/tenants/id/{id}/activate
POST /api/auth/core/tenants/{tenant_id}/users/id/{id}/enable
POST /api/auth/core/tenants/{tenant_id}/apps/{app_id}/keys/{key_id}/revoke
```

## Testing Different Deployments

To test different deployment modes, change the `@baseUrl`:

### Monolith (default)
```http
@baseUrl = http://localhost:9090
```

### Microservices
```http
# Core services (tenant, app, user, app-key, credential-config)
@baseUrl = http://localhost:8081

# Credential services (authentication)
@baseUrl = http://localhost:8082
```

### Production
```http
@baseUrl = https://api.your-domain.com
```

## Tips

1. **VS Code REST Client**: Required extension for testing
2. **Sequential Testing**: Always test in order (tenant → app → user → keys → config)
3. **Save IDs**: Copy important IDs from responses for subsequent requests
4. **Auto-extraction**: Use `# @name` and `{{requestName.response.body.$.field}}`
5. **Multi-tenant Testing**: Create multiple tenants to test isolation
6. **Config Merging**: Test partial updates to verify smart merge behavior

## Troubleshooting

### Server Not Running
```
Error: connect ECONNREFUSED
Solution: Start server with `go run .` from examples/01_deployment
```

### Invalid Tenant ID
```
Error: tenant not found
Solution: Verify tenant exists and ID matches exactly
```

### Router Conflict
```
Error: pattern conflicts with existing route
Solution: Use /id/{id} pattern instead of /{id} for dynamic routes
```

### Config Lost After Update
```
Issue: Credential configs disappearing after partial update
Solution: Smart merge now preserves configs not mentioned in request
```

## Security Best Practices

1. **App Keys:**
   - Store secrets securely (only shown once!)
   - Rotate keys periodically
   - Use environment-specific keys
   - Revoke compromised keys immediately

2. **Credential Config:**
   - Use strong password policies in production
   - Enable multiple auth methods for flexibility
   - Configure appropriate session timeouts
   - Use passkey/WebAuthn for maximum security

3. **Multi-tenant Isolation:**
   - Always include tenant_id in requests
   - Database query filters enforce tenant boundaries
   - App keys scoped to specific tenant+app

## Example Complete Workflow

1. **Create Tenant** → `acme-corp`
2. **Set Tenant Credential Defaults** → Basic + OAuth2 + API Key
3. **Create App** → `main-app` (inherits tenant defaults)
4. **Override App Config** → Stricter settings for production
5. **Create Users** → `admin@acme.com`, `user@acme.com`
6. **Generate App Keys** → For service-to-service auth
7. **Test Authentication** → Use configured credential providers

Now ready for authentication testing with credential services!
