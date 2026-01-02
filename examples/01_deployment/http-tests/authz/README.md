# Authorization (Authz) Service HTTP Tests

This folder contains HTTP test files for Authorization Service API endpoints.

## Endpoints Tested

### RBAC (Role-Based Access Control)
- `POST /api/auth/authz/roles` - Create/manage roles
- `GET /api/auth/authz/roles` - List roles
- `POST /api/auth/authz/roles/{id}/users` - Assign roles to users
- `GET /api/auth/authz/users/{id}/roles` - Get user's roles
- `GET /api/auth/authz/users/{id}/permissions` - Get user's effective permissions
- `POST /api/auth/authz/check` - Check authorization

### ABAC (Attribute-Based Access Control)
- `POST /api/auth/authz/abac/rules` - Create/manage ABAC rules
- `GET /api/auth/authz/abac/rules` - List ABAC rules
- `POST /api/auth/authz/abac/policies` - Create policies (rule collections)
- `POST /api/auth/authz/check` - Check with attributes

### ACL (Access Control Lists)
- `POST /api/auth/authz/acl/grant` - Grant permissions on resources
- `POST /api/auth/authz/acl/revoke` - Revoke permissions
- `GET /api/auth/authz/acl` - Get resource ACL
- `GET /api/auth/authz/acl/permissions` - Get user's permissions on resource
- `POST /api/auth/authz/check` - Check ACL-based authorization

### Unified Check Endpoint
- `POST /api/auth/authz/check` - Universal authorization check
- `POST /api/auth/authz/check/batch` - Batch authorization checks

## Test Files

### 01-rbac-service.http
**Role-Based Access Control** tests:
- Role creation (admin, editor, viewer)
- Permission assignment to roles
- Role assignment to users
- Role hierarchy
- Authorization checks
- Bulk operations

### 02-abac-service.http
**Attribute-Based Access Control** tests:
- ABAC rule creation (department, time, IP, status-based)
- Condition operators (equals, contains, greater_than, etc.)
- Policy management (rule collections)
- Attribute evaluation
- Context-aware authorization
- Complex attribute scenarios

### 03-acl-service.http
**Access Control Lists** tests:
- Direct permission grants (user, role, public)
- Permission revocation
- Wildcard permissions
- ACL inheritance (folder → document)
- ACL copying
- Audit and history
- Bulk operations

## Authorization Models Overview

### RBAC - Best for:
- **Organizational structures** with defined roles
- **Simple permission management**
- **Predictable access patterns**

Example:
```
Admin → [document:*, user:*, role:*]
Editor → [document:read, document:write]
Viewer → [document:read]
```

### ABAC - Best for:
- **Dynamic authorization** based on attributes
- **Context-aware decisions** (time, location, device)
- **Complex business rules**

Example:
```
Allow IF:
  subject.department == resource.department
  AND context.time.hour BETWEEN 9 AND 17
  AND context.ip_address STARTS_WITH "192.168."
```

### ACL - Best for:
- **Resource-specific permissions**
- **Sharing scenarios** (documents, files)
- **Fine-grained control** per resource

Example:
```
Document "doc-123":
  - usr-001: [read, write]
  - usr-002: [read]
  - role-editor: [read, write, update]
  - public: []
```

## Prerequisites

Before running these tests:

1. **Have a running server** at `http://localhost:9090`
2. **Authenticate** and get `access_token`
3. **Create tenant** and **app** using core tests
4. **Create users** for permission testing

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

Open each test file and update:

```http
@accessToken = eyJhbGci...   # From login response
@tenantId = acme-corp
@appId = main-app
@userId = usr-123
```

### 3. Run Tests

Click "Send Request" above each test case.

## Common Request Patterns

### Authorization Check Request

```http
POST /api/auth/authz/check
Content-Type: application/json
Authorization: Bearer {{accessToken}}
X-Tenant-ID: {{tenantId}}
X-App-ID: {{appId}}

{
  "subject": {
    "type": "user",
    "id": "usr-123",
    "roles": ["editor"],           // RBAC
    "attributes": {                 // ABAC
      "department": "engineering",
      "level": "senior"
    }
  },
  "action": "document:write",
  "resource": {
    "type": "document",
    "id": "doc-456",
    "attributes": {                 // ABAC
      "department": "engineering",
      "status": "draft"
    }
  },
  "context": {                      // ABAC
    "ip_address": "192.168.1.100",
    "time": {
      "hour": 14
    }
  }
}
```

### Success Response

```json
{
  "allowed": true,
  "reason": "User has 'editor' role with 'document:write' permission",
  "matched_rules": ["rbac-editor-rule"],
  "evaluated_at": "2025-11-26T14:30:00Z"
}
```

### Denial Response

```json
{
  "allowed": false,
  "reason": "User does not have required permission",
  "required_permissions": ["document:write"],
  "user_permissions": ["document:read"],
  "evaluated_at": "2025-11-26T14:30:00Z"
}
```

## Test Scenarios

### RBAC Scenarios

1. **Admin with wildcard permission** → ✅ Allow all actions
2. **Editor with specific permissions** → ✅ Allow read/write, ❌ Deny delete
3. **Viewer with read-only** → ✅ Allow read, ❌ Deny write/delete
4. **User with multiple roles** → ✅ Union of all role permissions
5. **No roles assigned** → ❌ Deny all (unless ACL grants access)

### ABAC Scenarios

1. **Same department** → ✅ Allow access
2. **Different department** → ❌ Deny access
3. **Manager override** → ✅ Allow cross-department access
4. **Working hours (9 AM - 5 PM)** → ✅ Allow, ❌ Deny outside hours
5. **Office IP whitelist** → ✅ Allow from 192.168.x.x, ❌ Deny others
6. **Published documents only** → ✅ Allow published, ❌ Deny draft

### ACL Scenarios

1. **Direct user grant** → ✅ User explicitly granted permission
2. **Role-based grant** → ✅ User has role with ACL entry
3. **Public resource** → ✅ Anyone can access
4. **Private resource** → ❌ Only explicitly granted users
5. **Owner wildcard** → ✅ Owner has all permissions
6. **Inherited permissions** → ✅ Child inherits parent ACL
7. **Expired grant** → ❌ Time-based grant expired

## Combining Authorization Models

You can use multiple models together:

```http
POST /api/auth/authz/check
{
  "subject": {
    "id": "usr-001",
    "roles": ["editor"],              // RBAC
    "attributes": {                    // ABAC
      "department": "engineering"
    }
  },
  "action": "document:publish",
  "resource": {
    "id": "doc-123",
    "attributes": {
      "department": "engineering"
    }
  }
}
```

Evaluation order:
1. **ACL** - Check if user explicitly has permission on resource
2. **RBAC** - Check if user's roles grant permission
3. **ABAC** - Evaluate attribute-based rules
4. **Result** - Allow if ANY method allows (OR logic)

## Batch Operations

### Batch Check Example

```http
POST /api/auth/authz/check/batch
{
  "subject": {
    "id": "usr-001",
    "roles": ["editor"]
  },
  "checks": [
    {
      "action": "document:read",
      "resource": {"type": "document", "id": "doc-1"}
    },
    {
      "action": "document:write",
      "resource": {"type": "document", "id": "doc-1"}
    },
    {
      "action": "document:delete",
      "resource": {"type": "document", "id": "doc-1"}
    }
  ]
}
```

**Response:**
```json
{
  "results": [
    {"allowed": true, "reason": "..."},
    {"allowed": true, "reason": "..."},
    {"allowed": false, "reason": "..."}
  ]
}
```

## Best Practices

### ✅ DO

- **Use RBAC** for role-based organizational structure
- **Use ABAC** for dynamic, attribute-based decisions
- **Use ACL** for resource-specific sharing
- **Combine models** for comprehensive authorization
- **Cache decisions** for frequently checked permissions
- **Log denials** for security auditing
- **Test edge cases** (empty roles, missing attributes)

### ❌ DON'T

- Don't rely solely on client-side checks
- Don't grant wildcard permissions carelessly
- Don't forget to revoke permissions when users leave
- Don't expose internal permission logic in error messages
- Don't skip authorization checks for "internal" APIs

## Performance Tips

1. **Batch checks** instead of individual requests
2. **Cache role-permission mappings**
3. **Use indexes** on frequently queried resources
4. **Limit ACL entries** per resource (max 100-1000)
5. **Prune expired grants** regularly

## Security Considerations

### Defense in Depth
```
Frontend Check (UX only)
    ↓
API Gateway Check (Performance)
    ↓
Service Check (Authoritative) ← THIS IS CRITICAL
    ↓
Database Row-Level Security (Last resort)
```

### Principle of Least Privilege
- Grant minimum required permissions
- Use time-limited grants for temporary access
- Review and revoke unused permissions regularly

### Audit Logging
All authorization decisions should be logged:
- Who requested access
- What action they tried
- On which resource
- Decision (allow/deny)
- Timestamp and context

## Troubleshooting

### "Permission denied" but user has role
- Check role actually has the permission
- Verify role is active (not expired)
- Check if ACL denies access explicitly
- Look for ABAC rules that deny

### "Allowed" but shouldn't be
- Check for wildcard permissions
- Verify public access not enabled
- Check inherited permissions from parent resources
- Review ABAC rules for unintended matches

### Slow authorization checks
- Too many ACL entries on resource
- Complex ABAC rules with nested attributes
- No caching of role permissions
- Database query optimization needed

## Related Documentation

- [Authorization Module](../../../04_authz/) - Implementation details
- [RBAC Evaluator](../../../04_authz/rbac/evaluator.go)
- [ABAC Evaluator](../../../04_authz/abac/evaluator.go)
- [ACL Manager](../../../04_authz/acl/manager.go)

## Next Steps

After testing authorization:
1. Integrate with middleware → `middleware-tests.http`
2. Test complete workflows → `core/00-workflow.http`
3. Load testing for performance tuning
4. Set up monitoring and alerting for auth failures
