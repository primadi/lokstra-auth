# HTTP Tests for Lokstra Auth

This folder contains HTTP test files demonstrating the complete authentication and authorization flow.

## Test Flow

1. **01-public.http** - Public endpoint (no auth required)
2. **02-login.http** - User login with basic auth
3. **03-protected.http** - Access protected endpoints with token
4. **04-authz-roles.http** - Role-based authorization demo
5. **05-authz-permissions.http** - Permission-based authorization demo
6. **06-refresh-token.http** - Token refresh flow
7. **07-complete-flow.http** - Complete end-to-end flow

## Test Users

### Alice (Admin)
- **Username:** alice
- **Password:** password123
- **Roles:** admin, editor
- **Permissions:** read, write, delete, document:*

### Bob (Editor)
- **Username:** bob
- **Password:** password123
- **Roles:** editor
- **Permissions:** read, write, document:read, document:write

### Charlie (Viewer)
- **Username:** charlie
- **Password:** password123
- **Roles:** viewer
- **Permissions:** read, document:read

## How to Use

1. Start the server: `go run .` from `examples/01_deployment`
2. Open any `.http` file in VS Code with REST Client extension
3. Click "Send Request" above each request

## Expected Results

- ✅ Alice can access all endpoints (admin + all permissions)
- ✅ Bob can access editor endpoints and read/write documents
- ✅ Charlie can only access viewer endpoints and read documents
- ❌ Unauthorized users get 401 errors
- ❌ Users without required roles/permissions get 403 errors

## API Structure

### Core Resources (Entity Management)
```
/api/auth/core/tenants
/api/auth/core/tenants/{tenantId}
/api/auth/core/tenants/{tenantId}/apps
/api/auth/core/tenants/{tenantId}/apps/{appId}
/api/auth/core/tenants/{tenantId}/apps/{appId}/branches
/api/auth/core/tenants/{tenantId}/apps/{appId}/keys
/api/auth/core/tenants/{tenantId}/users
/api/auth/core/tenants/{tenantId}/users/{userId}
/api/auth/core/tenants/{tenantId}/users/{userId}/identities
/api/auth/core/tenants/{tenantId}/credential-providers
```

### RBAC (Role-Based Access Control)
```
/api/auth/rbac/tenants/{tenantId}/apps/{appId}/roles
/api/auth/rbac/tenants/{tenantId}/apps/{appId}/roles/{roleId}
/api/auth/rbac/tenants/{tenantId}/apps/{appId}/roles/{roleId}/permissions
/api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions
/api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions/{permissionId}
/api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions/{permissionId}/compositions
/api/auth/rbac/tenants/{tenantId}/apps/{appId}/users/{userId}/roles
/api/auth/rbac/tenants/{tenantId}/apps/{appId}/users/{userId}/permissions
```

### Authorization (Policy-Based Access Control)
```
/api/auth/authz/tenants/{tenantId}/apps/{appId}/policies
/api/auth/authz/tenants/{tenantId}/apps/{appId}/policies/{policyId}
/api/auth/authz/check
/api/auth/authz/batch-check
```

### Authentication (Credentials)
```
/api/auth/cred/basic/login
/api/auth/cred/basic/register
/api/auth/cred/basic/change-password
/api/auth/cred/oauth2/authorize
/api/auth/cred/oauth2/callback
/api/auth/cred/oauth2/token
/api/auth/cred/passkey/register
/api/auth/cred/passkey/authenticate
/api/auth/cred/apikey/generate
/api/auth/cred/apikey/validate
```

### Token Management
```
/api/auth/token/validate
/api/auth/token/refresh
/api/auth/token/revoke
/api/auth/token/introspect
```

### Configuration (Credential Providers)
```
/api/auth/core/tenants/{tenantId}/config/credentials
/api/auth/core/tenants/{tenantId}/config/credentials/{providerId}
```

### Audit Logs
```
/api/auth/core/audit/logs
/api/auth/core/audit/logs/{logId}
/api/auth/core/audit/logs/cleanup
```

### Bootstrap & Admin
```
/api/auth/bootstrap
/api/auth/bootstrap/status
```

## Missing Endpoints (Recommendations)

### 1. **User-App Access Management**
```
POST   /api/auth/core/tenants/{tenantId}/apps/{appId}/users
DELETE /api/auth/core/tenants/{tenantId}/apps/{appId}/users/{userId}
GET    /api/auth/core/tenants/{tenantId}/apps/{appId}/users
```

### 2. **Tenant Ownership**
```
POST   /api/auth/core/tenants/{tenantId}/ownership/transfer
GET    /api/auth/core/tenants/{tenantId}/ownership/history
GET    /api/auth/core/tenants/{tenantId}/owner
```

### 3. **Permission Compositions (Compound Permissions)**
```
POST   /api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions/{permissionId}/compositions
DELETE /api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions/{permissionId}/compositions/{childPermissionId}
GET    /api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions/{permissionId}/compositions
GET    /api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions/{permissionId}/effective
```

### 4. **Batch Operations for RBAC**
```
POST   /api/auth/rbac/tenants/{tenantId}/apps/{appId}/users/{userId}/roles/batch
DELETE /api/auth/rbac/tenants/{tenantId}/apps/{appId}/users/{userId}/roles/batch
POST   /api/auth/rbac/tenants/{tenantId}/apps/{appId}/users/{userId}/permissions/batch
```

### 5. **User Search & Filtering**
```
GET    /api/auth/core/tenants/{tenantId}/users/search?q={query}
GET    /api/auth/core/tenants/{tenantId}/users?role={roleId}
GET    /api/auth/core/tenants/{tenantId}/users?status={status}
```

### 6. **Audit & Activity Logs**
```
GET    /api/auth/core/tenants/{tenantId}/users/{userId}/activity
GET    /api/auth/core/tenants/{tenantId}/audit/logs
GET    /api/auth/rbac/tenants/{tenantId}/apps/{appId}/roles/{roleId}/audit
```

### 7. **User Session Management**
```
GET    /api/auth/token/sessions
GET    /api/auth/token/sessions/{sessionId}
DELETE /api/auth/token/sessions/{sessionId}
DELETE /api/auth/token/sessions/all
```

### 8. **Account Management**
```
POST   /api/auth/core/tenants/{tenantId}/users/{userId}/lock
POST   /api/auth/core/tenants/{tenantId}/users/{userId}/unlock
POST   /api/auth/core/tenants/{tenantId}/users/{userId}/reset-failed-attempts
GET    /api/auth/core/tenants/{tenantId}/users/{userId}/lockout-status
```

## URL Design Improvements

### **Issues to Fix:**

1. **Double slash in app-key URL:**
   ```
   ❌ /api/auth/core//tenants/{tenantId}/apps/{appId}/keys
   ✅ /api/auth/core/tenants/{tenantId}/apps/{appId}/keys
   ```

2. **Inconsistent credential config path:**
   ```
   Current: /api/auth/core/config/credentials/tenants/{tenantId}
   Better:  /api/auth/core/tenants/{tenantId}/config/credentials
   
   OR keep as global config:
   /api/auth/config/credentials/tenants/{tenantId}
   ```

3. **User identity should be under user:**
   ```
   Current: /api/auth/core/config/credentials/tenants/{tenantId}  (unclear)
   Better:  /api/auth/core/tenants/{tenantId}/users/{userId}/identities
   ```

## Recommended Final Structure

```yaml
# Core - Entity Management
Core:
  Tenants:        /api/auth/core/tenants
  Apps:           /api/auth/core/tenants/{tenantId}/apps
  Branches:       /api/auth/core/tenants/{tenantId}/apps/{appId}/branches
  Users:          /api/auth/core/tenants/{tenantId}/users
  User Identities: /api/auth/core/tenants/{tenantId}/users/{userId}/identities
  User-App Access: /api/auth/core/tenants/{tenantId}/apps/{appId}/users
  App Keys:       /api/auth/core/tenants/{tenantId}/apps/{appId}/keys
  Credential Providers: /api/auth/core/tenants/{tenantId}/credential-providers
  Ownership:      /api/auth/core/tenants/{tenantId}/ownership

# RBAC - Role & Permission Management
RBAC:
  Roles:          /api/auth/rbac/tenants/{tenantId}/apps/{appId}/roles
  Permissions:    /api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions
  Permission Compositions: /api/auth/rbac/tenants/{tenantId}/apps/{appId}/permissions/{permissionId}/compositions
  Role Permissions: /api/auth/rbac/tenants/{tenantId}/apps/{appId}/roles/{roleId}/permissions
  User Roles:     /api/auth/rbac/tenants/{tenantId}/apps/{appId}/users/{userId}/roles
  User Permissions: /api/auth/rbac/tenants/{tenantId}/apps/{appId}/users/{userId}/permissions

# Authorization - Policy Engine
Authz:
  Policies:       /api/auth/authz/tenants/{tenantId}/apps/{appId}/policies
  Check:          /api/auth/authz/check
  Batch Check:    /api/auth/authz/batch-check

# Authentication - Credentials
Credentials:
  Basic Auth:     /api/auth/cred/basic/{action}
  OAuth2:         /api/auth/cred/oauth2/{action}
  Passkey:        /api/auth/cred/passkey/{action}
  API Key:        /api/auth/cred/apikey/{action}

# Token Management
Tokens:
  Validate:       /api/auth/token/validate
  Refresh:        /api/auth/token/refresh
  Revoke:         /api/auth/token/revoke
  Sessions:       /api/auth/token/sessions

# System
System:
  Bootstrap:      /api/auth/bootstrap
  Health:         /api/auth/health
  Metrics:        /api/auth/metrics
```

## Priority Implementation Order

1. ✅ **Already Done** - Core CRUD operations
2. 🔴 **High Priority**:
   - Permission compositions endpoints
   - Tenant ownership management
   - User-app access management
   - Account lockout management
3. 🟡 **Medium Priority**:
   - Batch operations for RBAC
   - User search & filtering
   - Token session management
4. 🟢 **Low Priority**:
   - Audit logs
   - Activity tracking
   - Advanced analytics
