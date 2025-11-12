# Authorization Examples

This directory contains examples demonstrating the authorization capabilities of Layer 04.

## Examples

### 1. RBAC Example (`01_rbac/`)

Demonstrates Role-Based Access Control (RBAC):
- Defining role permissions
- Evaluating access requests based on roles
- Checking permissions and roles
- Wildcard permissions

**Run**:
```bash
go run examples/04_authz/01_rbac/main.go
```

**Key Concepts**:
- Admin role with wildcard permissions (`document:*`)
- Editor role with specific permissions
- Viewer role with read-only access
- Permission checking helpers

**Output Shows**:
- Admin can delete documents (wildcard)
- Editor can write but not delete
- Viewer can only read
- Role and permission helper functions

### 2. ABAC Example (`02_abac/`)

Demonstrates Attribute-Based Access Control (ABAC):
- Defining rules with conditions
- Attribute-based evaluation
- Priority-based rule ordering
- Allow and deny effects

**Run**:
```bash
go run examples/04_authz/02_abac/main.go
```

**Key Concepts**:
- Subject attributes (department, role)
- Resource attributes (classification, department)
- Environment attributes (time_of_day)
- Condition operators (eq, gt, etc.)
- Rule priorities

**Rules Demonstrated**:
- Allow access when user department matches document department
- Allow managers to access all documents
- Deny confidential documents outside business hours

**Output Shows**:
- Engineer accessing engineering documents
- Manager accessing any documents
- Time-based access restrictions
- Department-based access control

### 3. ACL Example (`03_acl/`)

Demonstrates Access Control Lists (ACL):
- Granting resource-specific permissions
- User and role-based ACLs
- Permission management (grant, revoke)
- ACL operations (copy, delete)

**Run**:
```bash
go run examples/04_authz/03_acl/main.go
```

**Key Concepts**:
- Resource-level permissions
- User-specific permissions
- Role-based permissions
- Wildcard permissions (`*`)
- ACL management operations

**Output Shows**:
- Granting permissions to users and roles
- Revoking permissions
- Querying permissions
- Copying ACLs between resources
- Wildcard permission evaluation

## Running Examples

Each example is a standalone Go program. You can run them individually:

```bash
# Run RBAC example
go run examples/04_authz/01_rbac/main.go

# Run ABAC example
go run examples/04_authz/02_abac/main.go

# Run ACL example
go run examples/04_authz/03_acl/main.go
```

Or run all examples:

```bash
# PowerShell
Get-ChildItem examples/04_authz/*/main.go | ForEach-Object { 
    Write-Host "Running $_"
    go run $_.FullName
    Write-Host ""
}

# Bash
for dir in examples/04_authz/*/; do
    echo "Running ${dir}main.go"
    go run "${dir}main.go"
    echo ""
done
```

## Example Structure

Each example follows this pattern:

1. **Setup**: Create the authorizer with configuration
2. **Configuration**: Define rules, roles, or ACLs
3. **Test Cases**: Multiple scenarios testing different conditions
4. **Output**: Clear display of:
   - Subject information
   - Action requested
   - Resource accessed
   - Decision (allowed/denied)
   - Reason for decision

## Common Patterns

### Authorization Request

```go
request := &authz.AuthorizationRequest{
    Subject: identity,           // IdentityContext from Layer 03
    Action:  authz.Action("read"),
    Resource: &authz.Resource{
        Type: "document",
        ID:   "doc-123",
        Attributes: map[string]any{
            "classification": "confidential",
        },
    },
    Context: map[string]any{
        "time_of_day": 14,
    },
}
```

### Evaluating Requests

```go
decision, err := evaluator.Evaluate(ctx, request)
if err != nil {
    log.Fatal(err)
}

if decision.Allowed {
    fmt.Println("Access granted:", decision.Reason)
} else {
    fmt.Println("Access denied:", decision.Reason)
}
```

## Testing Different Scenarios

Each example demonstrates:

### Positive Cases (Access Granted)
- User has required role/permission
- User attributes match resource requirements
- ACL explicitly grants access

### Negative Cases (Access Denied)
- Missing required role/permission
- Attribute mismatch
- Deny rules taking precedence
- No ACL entry for user

## Extending Examples

You can extend these examples to test:

1. **Complex Rules**: Multiple conditions combined
2. **Priority Ordering**: Higher priority rules overriding lower ones
3. **Wildcard Patterns**: Various wildcard combinations
4. **Combined Models**: Using multiple authorization models together
5. **Dynamic Policies**: Loading policies from configuration files

## Integration Example

Here's how to integrate authorization into a real application:

```go
package main

import (
    "context"
    "log"
    
    subject "github.com/primadi/lokstra-auth/03_subject"
    authz "github.com/primadi/lokstra-auth/04_authz"
    "github.com/primadi/lokstra-auth/04_authz/rbac"
)

type DocumentHandler struct {
    authorizer authz.Authorizer
}

func (h *DocumentHandler) DeleteDocument(ctx context.Context, identity *subject.IdentityContext, docID string) error {
    // Create authorization request
    request := &authz.AuthorizationRequest{
        Subject: identity,
        Action:  authz.Action("delete"),
        Resource: &authz.Resource{
            Type: "document",
            ID:   docID,
        },
    }
    
    // Check authorization
    decision, err := h.authorizer.Evaluate(ctx, request)
    if err != nil {
        return err
    }
    
    if !decision.Allowed {
        log.Printf("Authorization denied: %s", decision.Reason)
        return fmt.Errorf("access denied: %s", decision.Reason)
    }
    
    // Proceed with deletion
    return h.deleteDocumentImpl(docID)
}
```

## Best Practices Demonstrated

1. **Clear Output**: Each example provides human-readable output
2. **Multiple Test Cases**: Cover both success and failure scenarios
3. **Error Handling**: Proper error handling in all examples
4. **Context Usage**: All examples use context.Context properly
5. **Realistic Scenarios**: Examples mirror real-world use cases

## Understanding Output

Example output format:

```
=== Test Case Description ===

Test N - Scenario:
  Subject: user-id (role: admin)
  Action: write
  Resource: document:doc-123
  Decision: true
  Reason: role admin has permission document:*
```

Each test shows:
- **Subject**: Who is requesting access (with roles/attributes)
- **Action**: What action is requested
- **Resource**: What resource is being accessed
- **Decision**: Whether access is allowed (true/false)
- **Reason**: Why the decision was made

## Troubleshooting

If an example doesn't compile:

1. **Check imports**: Ensure all packages are imported
2. **Run go mod tidy**: Update dependencies
3. **Check Go version**: Requires Go 1.21+

If authorization doesn't work as expected:

1. **Check role/permission mappings**: Ensure roles have correct permissions
2. **Verify rule conditions**: All conditions must match for rule to apply
3. **Check priorities**: Higher priority rules are evaluated first
4. **Review decision reason**: The reason field explains why access was granted/denied

## Next Steps

After understanding these examples:

1. **Combine Models**: Create examples that use multiple authorization models
2. **Add Persistence**: Store policies in a database instead of memory
3. **Add Caching**: Cache authorization decisions for performance
4. **Add Auditing**: Log all authorization decisions
5. **Create Middleware**: Build Lokstra middleware for automatic authorization

## Additional Examples

You can create additional examples for:

- Policy-based authorization with PolicyStore
- Combining RBAC + ACL
- Time-based access control
- Geographic-based access control
- Resource ownership checks
- Dynamic attribute resolution
