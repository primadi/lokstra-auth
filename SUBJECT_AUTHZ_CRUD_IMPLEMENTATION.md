# Subject & Authz CRUD Implementation Summary

## Overview
Created complete CRUD structure for Subject (Roles & Permissions) and Authz (Policies) following the same pattern as `/core`.

## Directory Structure Created

```
rbac/
├── application/
│   ├── role_service.go          # Role CRUD endpoints
│   ├── permission_service.go    # Permission CRUD endpoints
│   └── register.go              # Dependency injection
├── domain/
│   ├── role.go                  # Role domain models & DTOs
│   └── permission.go            # Permission domain models & DTOs
└── infrastructure/
    └── repository/
        ├── contract.go          # Repository interfaces
        └── postgres/
            ├── role_store.go        # PostgreSQL role store
            └── permission_store.go  # PostgreSQL permission store

authz/
├── application/
│   ├── policy_service.go        # Policy CRUD endpoints (TODO)
│   └── register.go              # Dependency injection (TODO)
├── domain/
│   └── policy.go                # Policy domain models & DTOs
└── infrastructure/
    └── repository/
        ├── contract.go          # Repository interfaces (TODO)
        └── postgres/
            └── policy_store.go  # PostgreSQL policy store (TODO)
```

## Database Tables (Migration: 004_subject_rbac.sql)

### Subject Tables

1. **roles** - Role definitions (tenant+app scoped)
   - id, tenant_id, app_id, name, description, status, metadata
   - Unique constraint: (tenant_id, app_id, name)

2. **user_roles** - User-to-role assignments
   - user_id, tenant_id, app_id, role_id, granted_at, revoked_at
   - Primary key: (user_id, tenant_id, app_id, role_id)

3. **permissions** - Permission definitions (tenant+app scoped)
   - id, tenant_id, app_id, name, description, resource, action, status, metadata
   - Unique constraint: (tenant_id, app_id, name)

4. **role_permissions** - Role-to-permission assignments
   - role_id, tenant_id, app_id, permission_id, granted_at, revoked_at
   - Primary key: (role_id, tenant_id, app_id, permission_id)

5. **user_permissions** - Direct user-to-permission assignments
   - user_id, tenant_id, app_id, permission_id, granted_at, revoked_at
   - Primary key: (user_id, tenant_id, app_id, permission_id)

### Authz Tables (TODO: 005_authz_policies.sql)

6. **policies** - Authorization policies (tenant+app scoped)
   - id, tenant_id, app_id, name, description, effect, subjects, resources, actions, conditions, status, metadata

## API Endpoints

### Subject - Role Management

```
POST   /api/rbac/roles                    - Create role
GET    /api/rbac/roles/:role_id           - Get role
PUT    /api/rbac/roles/:role_id           - Update role
DELETE /api/rbac/roles/:role_id           - Delete role
GET    /api/rbac/roles                    - List roles

POST   /api/rbac/users/:user_id/roles     - Assign role to user
DELETE /api/rbac/users/:user_id/roles/:role_id - Revoke role from user
GET    /api/rbac/users/:user_id/roles     - List user's roles
```

### Subject - Permission Management

```
POST   /api/rbac/permissions                           - Create permission
GET    /api/rbac/permissions/:permission_id            - Get permission
PUT    /api/rbac/permissions/:permission_id            - Update permission
DELETE /api/rbac/permissions/:permission_id            - Delete permission
GET    /api/rbac/permissions                           - List permissions

POST   /api/rbac/roles/:role_id/permissions            - Assign permission to role
DELETE /api/rbac/roles/:role_id/permissions/:permission_id - Revoke permission from role
GET    /api/rbac/roles/:role_id/permissions            - List role's permissions

POST   /api/rbac/users/:user_id/permissions            - Assign permission to user
DELETE /api/rbac/users/:user_id/permissions/:permission_id - Revoke permission from user
GET    /api/rbac/users/:user_id/permissions            - List user's permissions (with roles)
```

### Authz - Policy Management (TODO)

```
POST   /api/authz/policies                    - Create policy
GET    /api/authz/policies/:policy_id         - Get policy
PUT    /api/authz/policies/:policy_id         - Update policy
DELETE /api/authz/policies/:policy_id         - Delete policy
GET    /api/authz/policies                    - List policies

GET    /api/authz/policies/by-rbac/:subject_id - Find policies by subject
GET    /api/authz/policies/by-resource/:resource  - Find policies by resource

POST   /api/authz/evaluate                    - Evaluate authorization request
```

## Features Implemented

### Role Service
- ✅ Create role (with duplicate name check)
- ✅ Get role by ID
- ✅ Update role (partial updates)
- ✅ Delete role
- ✅ List roles with filters (status, pagination)
- ✅ Assign role to user
- ✅ Revoke role from user
- ✅ List user's roles
- ✅ Check if user has role

### Permission Service
- ✅ Create permission (with duplicate name check)
- ✅ Get permission by ID
- ✅ Update permission (partial updates)
- ✅ Delete permission
- ✅ List permissions with filters (resource, action, status, pagination)
- ✅ Assign permission to role
- ✅ Revoke permission from role
- ✅ List role's permissions
- ✅ Assign permission directly to user
- ✅ Revoke permission from user
- ✅ List user's permissions (including from roles) - with JOIN query
- ✅ Check if user has permission (direct or via role)

### PostgreSQL Stores
- ✅ Full CRUD operations
- ✅ Tenant+App isolation
- ✅ JSONB metadata support
- ✅ Soft delete with revoked_at timestamps
- ✅ Efficient JOIN queries for user permissions via roles
- ✅ Exists queries for permission checks
- ✅ Pagination support

## Key Features

### Multi-Tenancy & App Isolation
- All entities scoped to tenant_id + app_id
- Foreign key constraints for data integrity
- Indexes for query performance

### Soft Delete Pattern
- user_roles: revoked_at timestamp
- role_permissions: revoked_at timestamp
- user_permissions: revoked_at timestamp
- Enables audit trail and re-granting

### Permission Inheritance
- Users get permissions from:
  1. Direct user_permissions
  2. Permissions from assigned roles (via user_roles → role_permissions)
- Query optimization with UNION in ListUserPermissionsWithRoles

### Validation
- Duplicate name checks at service layer
- Foreign key validation at database layer
- Required field validation via domain Validate()

## TODO - Authz Policy Service

Still need to create:

1. **authz/infrastructure/repository/contract.go**
   - PolicyStore interface

2. **authz/infrastructure/repository/postgres/policy_store.go**
   - PostgreSQL implementation
   - Pattern matching for subjects, resources
   - Policy evaluation logic

3. **authz/application/policy_service.go**
   - CRUD operations
   - FindBySubject, FindByResource
   - Evaluate authorization requests

4. **authz/application/register.go**
   - Dependency injection

5. **deployment/migrations/005_authz_policies.sql**
   - CREATE TABLE policies
   - Indexes for performance

## Integration with Token Flow

### How It Works:

1. **Login** → Generate tokens with claims (tenant_id, app_id, user_id)

2. **Token Verification** → Extract claims

3. **Build IdentityContext**:
   ```go
   // RoleProvider fetches from user_roles + roles tables
   roles := userRoleStore.ListUserRoles(ctx, tenantID, appID, userID)
   
   // PermissionProvider fetches from user_permissions + role_permissions
   permissions := userPermissionStore.ListUserPermissionsWithRoles(ctx, tenantID, appID, userID)
   
   identity := &IdentityContext{
       Subject: subject,
       Roles: roles,
       Permissions: permissions,
   }
   ```

4. **Authorization Check**:
   ```go
   // Middleware check
   if !identity.HasRole("admin") {
       return 403
   }
   
   // Or permission check
   if !identity.HasPermission("documents:delete") {
       return 403
   }
   
   // Or policy evaluation
   decision := policyEvaluator.Evaluate(ctx, &AuthorizationRequest{
       Subject: identity,
       Resource: "document:123",
       Action: "delete",
   })
   ```

## Example Usage

### 1. Create Role & Permissions

```http
POST /api/rbac/roles
{
  "tenant_id": "acme-corp",
  "app_id": "main-app",
  "name": "editor",
  "description": "Content editor role"
}

POST /api/rbac/permissions
{
  "tenant_id": "acme-corp",
  "app_id": "main-app",
  "name": "documents:write",
  "description": "Write documents",
  "resource": "documents",
  "action": "write"
}
```

### 2. Assign Permission to Role

```http
POST /api/rbac/roles/{role_id}/permissions
{
  "tenant_id": "acme-corp",
  "app_id": "main-app",
  "role_id": "role-123",
  "permission_id": "perm-456"
}
```

### 3. Assign Role to User

```http
POST /api/rbac/users/john.doe/roles
{
  "tenant_id": "acme-corp",
  "app_id": "main-app",
  "user_id": "john.doe",
  "role_id": "role-123"
}
```

### 4. Check User Permissions

```http
GET /api/rbac/users/john.doe/permissions?tenant_id=acme-corp&app_id=main-app

Response:
[
  {
    "id": "perm-456",
    "name": "documents:write",
    "resource": "documents",
    "action": "write",
    ...
  }
]
```

## Database Schema Highlights

### Efficient Permission Query
```sql
-- Gets all user permissions (direct + from roles)
SELECT DISTINCT p.*
FROM permissions p
WHERE (p.tenant_id, p.app_id, p.id) IN (
    -- Direct permissions
    SELECT tenant_id, app_id, permission_id
    FROM user_permissions
    WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3
      AND revoked_at IS NULL
    
    UNION
    
    -- Permissions from roles
    SELECT rp.tenant_id, rp.app_id, rp.permission_id
    FROM role_permissions rp
    INNER JOIN user_roles ur ON rp.role_id = ur.role_id
    WHERE ur.tenant_id = $1 AND ur.app_id = $2 AND ur.user_id = $3
      AND ur.revoked_at IS NULL
      AND rp.revoked_at IS NULL
)
```

## Next Steps

1. ✅ Run migration: `004_subject_rbac.sql`
2. ⏳ Complete Authz Policy implementation
3. ⏳ Create HTTP test files for Subject endpoints
4. ⏳ Integrate RoleProvider & PermissionProvider with IdentityContextBuilder
5. ⏳ Create seed data for default roles/permissions
6. ⏳ Add bulk operations (assign multiple roles/permissions)
7. ⏳ Add permission dependency management
8. ⏳ Add role hierarchy support

## Files Created

### Subject (Complete)
- `rbac/domain/role.go` - 140 lines
- `rbac/domain/permission.go` - 170 lines
- `rbac/infrastructure/repository/contract.go` - 120 lines
- `rbac/infrastructure/repository/postgres/role_store.go` - 450 lines
- `rbac/infrastructure/repository/postgres/permission_store.go` - 650 lines
- `rbac/application/role_service.go` - 160 lines
- `rbac/application/permission_service.go` - 250 lines
- `rbac/application/register.go` - 60 lines
- `deployment/migrations/004_subject_rbac.sql` - 100 lines

### Authz (Partial)
- `authz/domain/policy.go` - 160 lines

**Total: ~2,260 lines of code**
