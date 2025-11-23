# HTTP Tests for Lokstra Auth Services

This folder contains HTTP test files for testing all core services of Lokstra Auth.

## Prerequisites

- Install REST Client extension for VS Code
- Ensure the server is running on `http://localhost:8080`
- Or modify `@baseUrl` variable in each file

## Test Files

### 00-workflow.http
Complete end-to-end workflow for testing the entire system. This is the recommended starting point.

**Features:**
- Automated variable extraction from responses
- Step-by-step tenant setup
- Verification checks
- All services in correct order

**Usage:**
1. Open `00-workflow.http`
2. Execute requests sequentially from top to bottom
3. Variables will be automatically populated

---

### 01-tenant-service.http
Tests for tenant management operations.

**Endpoints:**
- Create tenant
- Get tenant by ID
- Get tenant by code
- List all tenants
- Update tenant
- Delete tenant

---

### 02-app-service.http
Tests for application management within a tenant.

**Endpoints:**
- Create app
- Get app by ID
- Get app by code
- List all apps
- Update app
- Delete app

**Required Variables:**
- `@tenantId` - Must be set before testing

---

### 03-branch-service.http
Tests for branch management within an app.

**Endpoints:**
- Create branch (production & development)
- Get branch by ID
- Get branch by code
- List all branches
- Update branch
- Delete branch

**Required Variables:**
- `@tenantId`
- `@appId`

---

### 04-user-service.http
Tests for user management within a tenant.

**Endpoints:**
- Create user (regular & admin)
- Get user by ID
- Get user by username
- Get user by email
- List all users
- Update user
- Update password
- Enable/disable user
- Delete user

**Required Variables:**
- `@tenantId`

---

### 05-app-key-service.http
Tests for API key management.

**Endpoints:**
- Create app key
- Get app key by ID
- List all app keys
- Validate app key
- Update app key
- Revoke app key
- Rotate app key
- Delete app key

**Required Variables:**
- `@tenantId`
- `@appId`
- `@branchId`

---

### 06-credential-config-service.http
Tests for credential configuration management.

**Endpoints:**
- Create configs for all credential types:
  - Basic Auth
  - OAuth2
  - API Key
  - Passkey
  - Passwordless
- Get config by ID
- Get config by type
- List all configs
- Update config
- Enable/disable config
- Delete config

**Required Variables:**
- `@tenantId`
- `@appId`
- `@branchId`

---

## Quick Start Guide

### Option 1: Complete Workflow (Recommended)
```
1. Open 00-workflow.http
2. Click "Send Request" for each step sequentially
3. Variables will auto-populate
4. Verify at the end
```

### Option 2: Individual Service Testing
```
1. Start with 01-tenant-service.http
2. Copy the tenant ID from response
3. Set @tenantId in the next file
4. Continue with 02-app-service.http
5. Repeat for each service
```

## Common Variables

All files use these common variables:

```http
@baseUrl = http://localhost:8080
@contentType = application/json
@tenantId = paste-tenant-id-here
@appId = paste-app-id-here
@branchId = paste-branch-id-here
@userId = paste-user-id-here
@appKeyId = paste-app-key-id-here
@configId = paste-config-id-here
```

## Testing Different Deployments

To test different deployment modes, change the `@baseUrl`:

### Monolith (default)
```http
@baseUrl = http://localhost:8080
```

### Microservices
```http
# For core services (tenant, app, branch, user, app-key, credential-config)
@baseUrl = http://localhost:8081

# For credential services (when testing auth)
@baseUrl = http://localhost:8082
```

### Development
```http
@baseUrl = http://localhost:3000
```

## Tips

1. **VS Code REST Client Extension**: Install it for the best experience
2. **Sequential Testing**: Always test services in order (tenant → app → branch)
3. **Save IDs**: Copy important IDs from responses to use in subsequent requests
4. **Environment Variables**: Use REST Client's environment feature for different configs
5. **Response Extraction**: Use `@name` and `{{requestName.response.body.$.field}}` for auto-extraction

## Troubleshooting

### Server Not Running
```
Error: connect ECONNREFUSED
Solution: Start the server first
```

### Invalid Tenant ID
```
Error: 404 Not Found
Solution: Verify the tenant exists and ID is correct
```

### Unauthorized
```
Error: 401 Unauthorized
Solution: Check if authentication is required and provide valid credentials
```

## Example Workflow

1. **Create Tenant** → Get `tenant_id`
2. **Create App** (with `tenant_id`) → Get `app_id`
3. **Create Branch** (with `tenant_id` + `app_id`) → Get `branch_id`
4. **Create User** (with `tenant_id`)
5. **Create App Key** (with `tenant_id` + `app_id` + `branch_id`)
6. **Configure Credentials** (with `tenant_id` + `app_id` + `branch_id`)

Now you're ready to test authentication!
