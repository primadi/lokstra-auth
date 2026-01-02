# Layer 04: Authorization (Authz)

Layer 04 provides comprehensive authorization mechanisms for controlling access to resources. This layer builds on Layer 03 (Subject) to make access control decisions based on various models including RBAC, ABAC, ACL, and Policy-based authorization.

## Overview

Authorization determines what an authenticated and identified subject can do with resources. This layer supports multiple authorization models that can be used independently or combined:

- **RBAC** (Role-Based Access Control): Permissions based on roles
- **ABAC** (Attribute-Based Access Control): Permissions based on attributes
- **ACL** (Access Control Lists): Direct resource-level permissions
- **Policy-Based**: Flexible policy evaluation with multiple combining algorithms

## Core Concepts

### Authorization Request

```go
type AuthorizationRequest struct {
    Subject  *subject.IdentityContext  // Who is requesting access
    Action   Action                    // What action they want to perform
    Resource *Resource                 // What resource they want to access
    Context  map[string]any           // Additional context (environment, time, etc.)
}
```

### Authorization Decision

```go
type AuthorizationDecision struct {
    Allowed  bool           // Whether access is allowed
    Reason   string         // Human-readable reason
    Metadata map[string]any // Additional decision metadata
}
```

### Resource

```go
type Resource struct {
    Type       string         // Resource type (e.g., "document", "user")
    ID         string         // Resource identifier
    Attributes map[string]any // Resource attributes
}
```

## Components

### 1. RBAC (Role-Based Access Control)

RBAC grants permissions based on roles assigned to subjects. It supports:
- Role-to-permission mapping
- Wildcard permissions (`document:*`)
- Simple and complex permission formats

**Location**: `authz/rbac/evaluator.go`

**Usage**:

```go
rolePermissions := map[string][]string{
    "admin":  {"document:*"},
    "editor": {"document:read", "document:write"},
    "viewer": {"document:read"},
}
evaluator := rbac.NewEvaluator(rolePermissions)

decision, err := evaluator.Evaluate(ctx, request)
```

**Features**:
- `Evaluate()` - Evaluate authorization request
- `HasPermission()` - Check if identity has a specific permission
- `HasRole()` - Check if identity has a specific role
- `AddRolePermission()` - Add permissions to a role
- `GetRolePermissions()` - Get all permissions for a role

### 2. ABAC (Attribute-Based Access Control)

ABAC makes decisions based on attributes of the subject, resource, environment, and action. It supports:
- Rule-based evaluation
- Multiple condition types (subject, resource, environment, action)
- Various operators (eq, ne, in, not_in, gt, lt, contains)
- Priority-based rule ordering
- Allow and deny effects

**Location**: `authz/abac/evaluator.go`

**Rule Structure**:

```go
type Rule struct {
    ID          string
    Description string
    Effect      string      // "allow" or "deny"
    Conditions  []Condition
    Priority    int         // Higher priority = evaluated first
}

type Condition struct {
    Type     string // "subject", "resource", "environment", "action"
    Key      string
    Operator string // "eq", "ne", "in", "not_in", "gt", "lt", "contains"
    Value    any
}
```

**Usage**:

```go
evaluator := abac.NewEvaluator(attributeProvider, false)

rule := &abac.Rule{
    ID:       "allow-dept-access",
    Effect:   "allow",
    Priority: 10,
    Conditions: []abac.Condition{
        {
            Type:     "subject",
            Key:      "department",
            Operator: "eq",
            Value:    "engineering",
        },
    },
}
evaluator.AddRule(rule)

decision, err := evaluator.Evaluate(ctx, request)
```

**Operators**:
- `eq` - Equal
- `ne` - Not equal
- `in` - Value in list
- `not_in` - Value not in list
- `gt` - Greater than
- `lt` - Less than
- `contains` - String contains

### 3. ACL (Access Control Lists)

ACL provides fine-grained, resource-level access control. It supports:
- User and role-based permissions
- Per-resource permission management
- Wildcard permissions
- ACL inheritance and copying

**Location**: `authz/acl/manager.go`

**Usage**:

```go
manager := acl.NewManager()

// Grant permissions
manager.Grant(ctx, "document", "doc-123", "user-1", "user", "read", "write")
manager.Grant(ctx, "document", "doc-123", "editor", "role", "write")

// Check access
decision, err := manager.Evaluate(ctx, request)

// Get permissions
perms, err := manager.GetPermissions(ctx, "document", "doc-123", "user-1", identity)

// Revoke permissions
manager.Revoke(ctx, "document", "doc-123", "user-1", "user", "write")

// Copy ACL
manager.CopyACL(ctx, "document", "doc-123", "document", "doc-456")
```

**Features**:
- `Grant()` - Grant permissions to a subject for a resource
- `Revoke()` - Revoke specific permissions
- `RevokeAll()` - Revoke all permissions for a subject
- `Check()` - Check if subject has permission
- `GetPermissions()` - Get all permissions for a subject
- `GetSubjects()` - Get all subjects with access to a resource
- `GetACL()` - Get full ACL for a resource
- `SetACL()` - Set ACL for a resource
- `DeleteACL()` - Delete ACL for a resource
- `CopyACL()` - Copy ACL from one resource to another
- `Evaluate()` - Evaluate authorization request

### 4. Policy-Based Authorization

Policy-based authorization evaluates policies stored in a PolicyStore. It supports:
- Subject, resource, and action-based policies
- Wildcard matching for subjects and resources
- Multiple policy combining algorithms
- Conditional policies

**Location**: `authz/policy/`

**Components**:
- `PolicyStore` (`policy/store.go`) - Store and retrieve policies
- `PolicyEvaluator` (`policy/evaluator.go`) - Evaluate policies

**Policy Structure**:

```go
type Policy struct {
    ID          string
    Description string
    Effect      string         // "allow" or "deny"
    Subjects    []string       // Subject IDs, "*", or "role:rolename"
    Resources   []string       // "type:id", "type:*", or "*"
    Actions     []string       // Actions or "*"
    Conditions  map[string]any // Optional conditions
}
```

**Usage**:

```go
// Create store and evaluator
store := policy.NewInMemoryStore()
evaluator := policy.NewEvaluator(store, "deny-overrides")

// Create policy
policy := &authz.Policy{
    ID:          "allow-admin",
    Effect:      "allow",
    Subjects:    []string{"role:admin"},
    Resources:   []string{"*"},
    Actions:     []string{"*"},
}

// Store policy
store.Create(ctx, policy)

// Evaluate
decision, err := evaluator.Evaluate(ctx, request)
```

**Combining Algorithms**:
- `deny-overrides` - If any policy denies, result is deny (default)
- `allow-overrides` - If any policy allows, result is allow
- `first-applicable` - First matching policy wins

## Interface: Authorizer

All authorization components implement the `Authorizer` interface:

```go
type Authorizer interface {
    Evaluate(ctx context.Context, request *AuthorizationRequest) (*AuthorizationDecision, error)
}
```

This allows them to be used interchangeably or combined.

## Examples

See the `examples/authz/` directory for complete working examples:

- `01_rbac_example.go` - Role-based access control
- `02_abac_example.go` - Attribute-based access control
- `03_acl_example.go` - Access control lists

Each example can be run with:

```bash
go run examples/authz/01_rbac_example.go
go run examples/authz/02_abac_example.go
go run examples/authz/03_acl_example.go
```

## Integration with Lokstra

Authorization can be integrated into Lokstra request handlers:

```go
// In your handler
func (h *Handler) Handle(c *lokstra.Context) error {
    // Get identity from context (set by authentication middleware)
    identity := c.Get("identity").(*subject.IdentityContext)
    
    // Create authorization request
    request := &authz.AuthorizationRequest{
        Subject: identity,
        Action:  authz.Action("read"),
        Resource: &authz.Resource{
            Type: "document",
            ID:   c.Param("id"),
        },
    }
    
    // Evaluate
    decision, err := h.authorizer.Evaluate(c.Context(), request)
    if err != nil {
        return err
    }
    
    if !decision.Allowed {
        return c.Status(403).JSON(map[string]string{
            "error": "access denied",
            "reason": decision.Reason,
        })
    }
    
    // Proceed with request...
    return nil
}
```

## Combining Authorization Models

You can combine multiple authorization models:

```go
type CombinedAuthorizer struct {
    rbac  *rbac.Evaluator
    acl   *acl.Manager
    policy *policy.Evaluator
}

func (a *CombinedAuthorizer) Evaluate(ctx context.Context, request *authz.AuthorizationRequest) (*authz.AuthorizationDecision, error) {
    // Check ACL first (most specific)
    if decision, _ := a.acl.Evaluate(ctx, request); decision.Allowed {
        return decision, nil
    }
    
    // Check RBAC
    if decision, _ := a.rbac.Evaluate(ctx, request); decision.Allowed {
        return decision, nil
    }
    
    // Check policies (most flexible)
    return a.policy.Evaluate(ctx, request)
}
```

## Best Practices

1. **Choose the Right Model**:
   - Use RBAC for role-based systems
   - Use ABAC for complex attribute-based rules
   - Use ACL for fine-grained resource control
   - Use Policy for flexible, dynamic authorization

2. **Combine Models**: Different models can work together for defense-in-depth

3. **Default Deny**: Always default to denying access unless explicitly allowed

4. **Audit Decisions**: Log authorization decisions for security auditing

5. **Performance**: Cache authorization decisions when possible

6. **Test Thoroughly**: Test both positive and negative cases

## Thread Safety

All authorization components are thread-safe and can be used concurrently:
- RBAC evaluator (read-only after initialization)
- ABAC evaluator (synchronized rule management)
- ACL manager (synchronized with sync.RWMutex)
- Policy store (synchronized with sync.RWMutex)

## Error Handling

Authorization evaluators return errors only for unexpected conditions (e.g., storage failures). Access denial is indicated through the `Allowed` field in `AuthorizationDecision`, not through errors.

```go
decision, err := evaluator.Evaluate(ctx, request)
if err != nil {
    // Unexpected error (storage, network, etc.)
    return err
}

if !decision.Allowed {
    // Access denied - this is normal operation
    log.Printf("Access denied: %s", decision.Reason)
}
```

## Architecture

```
authz/
├── contract.go          # Core interfaces and types
├── rbac/
│   └── evaluator.go     # Role-based access control
├── abac/
│   └── evaluator.go     # Attribute-based access control
├── acl/
│   └── manager.go       # Access control lists
└── policy/
    ├── store.go         # Policy storage
    └── evaluator.go     # Policy evaluation
```

## Dependencies

- Layer 03 (Subject): Provides `IdentityContext` for authorization requests
- Standard library: `context`, `sync`, `fmt`, `strings`

## Next Steps

After implementing authorization:
- Add audit logging for authorization decisions
- Implement caching for frequently-checked permissions
- Add monitoring and metrics
- Consider implementing policy administration UI
- Add support for dynamic policy loading

## Additional Resources

- [NIST RBAC Model](https://csrc.nist.gov/projects/role-based-access-control)
- [XACML ABAC Standard](https://www.oasis-open.org/committees/tc_home.php?wg_abbrev=xacml)
- [AWS IAM Policy Evaluation Logic](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_evaluation-logic.html)
