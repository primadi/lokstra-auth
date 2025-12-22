# Subject Module HTTP Tests

This folder contains HTTP test files for Subject Module API endpoints.

## Overview

The Subject module provides **database-backed RBAC (Role-Based Access Control)** for managing:
- Roles (CRUD)
- Permissions (CRUD)
- User-Role assignments
- Role-Permission assignments
- User-Permission assignments (direct)

## Endpoints

### Roles
- `POST /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles` - Create role
- `GET /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles/{role_id}` - Get role
- `PUT /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles/{role_id}` - Update role
- `DELETE /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles/{role_id}` - Delete role
- `GET /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles` - List roles

### User-Role Assignments
- `POST /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/roles` - Assign role
- `DELETE /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/roles/{role_id}` - Revoke role
- `GET /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/roles` - List user's roles

### Permissions
- `POST /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/permissions` - Create permission
- `GET /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/permissions/{permission_id}` - Get permission
- `PUT /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/permissions/{permission_id}` - Update permission
- `DELETE /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/permissions/{permission_id}` - Delete permission
- `GET /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/permissions` - List permissions

### Role-Permission Assignments
- `POST /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles/{role_id}/permissions` - Assign permission to role
- `DELETE /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles/{role_id}/permissions/{permission_id}` - Revoke permission from role
- `GET /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/roles/{role_id}/permissions` - List role's permissions

### User-Permission Assignments
- `POST /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/permissions` - Assign direct permission
- `DELETE /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/permissions/{permission_id}` - Revoke direct permission
- `GET /api/auth/rbac/tenants/{tenant_id}/apps/{app_id}/users/{user_id}/permissions` - List user's permissions (direct + inherited)

## Test Files

### 01-role-service.http
**Role Management and User-Role Assignments**
- Create roles (admin, editor, viewer)
- Update role details and status
- Delete roles
- List and filter roles
- Assign roles to users
- Revoke roles from users
- List user's roles
- Multi-tenant scenarios
- Edge cases and error handling

### 02-permission-service.http
**Permission Management and Assignments**
- Create permissions for different resources and actions
- Update permission details
- Delete permissions
- List and filter permissions
- Assign permissions to roles
- Revoke permissions from roles
- List role's permissions
- Direct user-permission assignments
- Permission inheritance verification
- Edge cases and error handling

## Architecture

### Database Tables
```sql
-- Roles
roles (id, tenant_id, app_id, name, description, status, metadata)

-- User-Role assignments
user_roles (user_id, tenant_id, app_id, role_id, granted_at, revoked_at)

-- Permissions
permissions (id, tenant_id, app_id, name, description, resource, action, status, metadata)

-- Role-Permission assignments
role_permissions (role_id, tenant_id, app_id, permission_id, granted_at, revoked_at)

-- User-Permission assignments (direct)
user_permissions (user_id, tenant_id, app_id, permission_id, granted_at, revoked_at)
```

### Multi-Tenancy
All resources are scoped to `(tenant_id, app_id)`:
- Roles are unique per tenant+app
- Permissions are unique per tenant+app
- Users can have different roles in different tenants/apps
- Complete data isolation between tenants

### Permission Inheritance
Users get permissions from **TWO sources**:
1. **Direct**: Permissions assigned directly via `user_permissions` table
2. **Inherited**: Permissions from roles assigned via `user_roles` → `role_permissions`

Query example:
```sql
SELECT DISTINCT p.* FROM permissions p
WHERE (
  -- Direct permissions
  EXISTS (
    SELECT 1 FROM user_permissions up 
    WHERE up.user_id = $1 AND up.permission_id = p.id 
    AND up.revoked_at IS NULL
  )
  OR
  -- Inherited from roles
  EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN role_permissions rp ON rp.role_id = ur.role_id
    WHERE ur.user_id = $1 AND rp.permission_id = p.id
    AND ur.revoked_at IS NULL AND rp.revoked_at IS NULL
  )
)
```

## Prerequisites

1. **Server running** at `http://localhost:9090`
2. **Tenant and App created** (use `core/01-tenant-service.http`, `core/02-app-service.http`)
3. **Users created** (use `core/04-user-service.http`)
4. **Access token** obtained (use `credential/01-basic-auth-service.http`)

## Quick Start

### 1. Get Access Token
```http
POST http://localhost:9090/api/auth/cred/basic/login
X-Tenant-ID: acme-corp
X-App-ID: main-app
Content-Type: application/json

{
  "username": "admin",
  "password": "SecurePass123!"
}
```

### 2. Update Variables
Open `01-role-service.http` and `02-permission-service.http`:
```http
@accessToken = eyJhbGci...   # From login response
@tenantId = acme-corp
@appId = main-app
@userId = usr-123
```

### 3. Run Tests
Execute requests in order:
1. **01-role-service.http** - Create roles and assign to users
2. **02-permission-service.http** - Create permissions and link to roles

## Common Workflows

### Workflow 1: Complete RBAC Setup

**Step 1: Create Role Hierarchy**
```
Admin → Full access
  ├── Manager → Elevated access
  │     ├── Editor → Read + Write
  │     └── Viewer → Read only
```

**Step 2: Define Permissions**
```
Documents:
  - documents_read
  - documents_write
  - documents_delete
  - documents_publish

Users:
  - users_manage
  - users_view
```

**Step 3: Assign Permissions to Roles**
```
Admin → [documents_*, users_*]
Editor → [documents_read, documents_write]
Viewer → [documents_read]
```

**Step 4: Assign Roles to Users**
```
usr-001 → Admin
usr-002 → Editor
usr-003 → Viewer
```

**Step 5: Grant Direct Permissions**
```
usr-002 → documents_publish (special permission)
```

**Result:**
- `usr-001` has ALL permissions (from admin role)
- `usr-002` has read, write, publish (from editor role + direct permission)
- `usr-003` has read only (from viewer role)

### Workflow 2: Dynamic Permission Management

1. Create role without permissions
2. Assign role to users
3. Add permissions to role dynamically
4. All users with that role immediately get new permissions
5. Remove permission from role
6. All users lose that permission (unless granted directly)

### Workflow 3: Temporary Access

1. Assign direct permission to user (bypassing roles)
2. User has access
3. Revoke direct permission
4. User loses access (unless they have it from a role)

## Response Examples

### Create Role Response
```json
{
  "id": "role_abc123",
  "tenant_id": "acme-corp",
  "app_id": "main-app",
  "name": "admin",
  "description": "Administrator role",
  "status": "active",
  "metadata": {
    "level": "system",
    "priority": 100
  },
  "created_at": "2025-11-27T10:00:00Z",
  "updated_at": "2025-11-27T10:00:00Z"
}
```

### List User's Permissions Response
```json
[
  {
    "id": "perm_read_documents",
    "name": "read_documents",
    "description": "Permission to read documents",
    "resource": "documents",
    "action": "read",
    "source": "role:admin"  // or "direct"
  },
  {
    "id": "perm_write_documents",
    "name": "write_documents",
    "description": "Permission to write documents",
    "resource": "documents",
    "action": "write",
    "source": "role:admin"
  }
]
```

## Best Practices

### ✅ DO
- Use roles for organizational structure
- Grant direct permissions sparingly (exceptions only)
- Name permissions clearly: `{resource}_{action}` format
- Use metadata for additional context
- Soft delete with `revoked_at` for audit trail
- Test permission inheritance queries
- Paginate list operations

### ❌ DON'T
- Don't grant overly broad permissions
- Don't bypass role-based access with too many direct permissions
- Don't forget to revoke permissions when users change roles
- Don't create duplicate role or permission names
- Don't hardcode tenant/app IDs

## Troubleshooting

### Issue: User doesn't have expected permissions
**Check:**
1. User has role assigned? `GET /users/{user_id}/roles`
2. Role has permission? `GET /roles/{role_id}/permissions`
3. Permission still active? Check `status` and `revoked_at`
4. Correct tenant/app scope?

### Issue: Duplicate name error
**Cause:** Role/Permission names must be unique within `(tenant_id, app_id)` scope

**Solution:** Use different name or different tenant/app

### Issue: Permission not showing in user's list
**Check:**
1. Permission assigned to role? Check `role_permissions` table
2. Role assigned to user? Check `user_roles` table
3. No `revoked_at` timestamps?

## Performance Optimization

### Indexes
```sql
CREATE INDEX idx_roles_tenant_app ON roles(tenant_id, app_id);
CREATE INDEX idx_user_roles_user ON user_roles(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id) WHERE revoked_at IS NULL;
```

### Caching
- Cache role-permission mappings (invalidate on change)
- Cache user-role assignments (invalidate on assignment/revocation)
- TTL: 5-15 minutes for most use cases

## Related Documentation

- [Subject Domain Models](../../../rbac/domain/)
- [Subject Repository Implementation](../../../rbac/infrastructure/repository/postgres/)
- [Subject Application Services](../../../rbac/application/)
- [Database Migration](../../../deployment/migrations/004_subject_rbac.sql)

## Next Steps

After testing Subject module:
1. Test integration with Authz (Policy) module
2. Test middleware authorization checks
3. Create load tests for permission queries
4. Set up audit logging for role/permission changes
