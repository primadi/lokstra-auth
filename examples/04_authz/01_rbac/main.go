package main

import (
	"context"
	"fmt"
	"log"

	authz "github.com/primadi/lokstra-auth/authz"
	"github.com/primadi/lokstra-auth/authz/rbac"
	identity "github.com/primadi/lokstra-auth/identity"
)

func main() {
	fmt.Println("=== RBAC Authorization Example ===")

	ctx := context.Background()

	// Create RBAC evaluator with role permissions (multi-tenant aware)
	// Key format: "tenantID:appID:role" -> []permissions
	rolePermissions := map[string][]string{
		"demo-tenant:demo-app:admin":  {"document:*"},
		"demo-tenant:demo-app:editor": {"document:read", "document:write"},
		"demo-tenant:demo-app:viewer": {"document:read"},
	}
	evaluator := rbac.NewEvaluator(rolePermissions)

	fmt.Println("Role Permissions (tenant: demo-tenant, app: demo-app):")
	fmt.Println("- admin: document:*")
	fmt.Println("- editor: document:read, document:write")
	fmt.Println("- viewer: document:read")
	fmt.Println()

	// Test Case 1: Admin user
	adminIdentity := &identity.IdentityContext{
		Subject: &identity.Subject{
			ID:       "user-1",
			TenantID: "demo-tenant",
			Type:     "user",
		},
		TenantID:    "demo-tenant",
		AppID:       "demo-app",
		Roles:       []string{"admin"},
		Permissions: []string{},
	}

	request1 := &authz.AuthorizationRequest{
		Subject: adminIdentity,
		Action:  authz.ActionDelete,
		Resource: &authz.Resource{
			Type:     "document",
			ID:       "doc-123",
			TenantID: "demo-tenant",
			AppID:    "demo-app",
		},
	}

	decision1, err := evaluator.Evaluate(ctx, request1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 1 - Admin deleting document:\n")
	fmt.Printf("  Subject: %s (roles: %v)\n", adminIdentity.Subject.ID, adminIdentity.Roles)
	fmt.Printf("  Action: %s\n", request1.Action)
	fmt.Printf("  Resource: %s:%s\n", request1.Resource.Type, request1.Resource.ID)
	fmt.Printf("  Decision: %v\n", decision1.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision1.Reason)

	// Test Case 2: Editor user
	editorIdentity := &identity.IdentityContext{
		Subject: &identity.Subject{
			ID:       "user-2",
			TenantID: "demo-tenant",
			Type:     "user",
		},
		TenantID:    "demo-tenant",
		AppID:       "demo-app",
		Roles:       []string{"editor"},
		Permissions: []string{},
	}

	request2 := &authz.AuthorizationRequest{
		Subject: editorIdentity,
		Action:  authz.ActionWrite,
		Resource: &authz.Resource{
			Type:     "document",
			ID:       "doc-456",
			TenantID: "demo-tenant",
			AppID:    "demo-app",
		},
	}

	decision2, err := evaluator.Evaluate(ctx, request2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 2 - Editor writing document:\n")
	fmt.Printf("  Subject: %s (roles: %v)\n", editorIdentity.Subject.ID, editorIdentity.Roles)
	fmt.Printf("  Action: %s\n", request2.Action)
	fmt.Printf("  Resource: %s:%s\n", request2.Resource.Type, request2.Resource.ID)
	fmt.Printf("  Decision: %v\n", decision2.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision2.Reason)

	// Test Case 3: Editor trying to delete (should be denied)
	request3 := &authz.AuthorizationRequest{
		Subject: editorIdentity,
		Action:  authz.ActionDelete,
		Resource: &authz.Resource{
			Type:     "document",
			ID:       "doc-456",
			TenantID: "demo-tenant",
			AppID:    "demo-app",
		},
	}

	decision3, err := evaluator.Evaluate(ctx, request3)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 3 - Editor deleting document (should fail):\n")
	fmt.Printf("  Subject: %s (roles: %v)\n", editorIdentity.Subject.ID, editorIdentity.Roles)
	fmt.Printf("  Action: %s\n", request3.Action)
	fmt.Printf("  Resource: %s:%s\n", request3.Resource.Type, request3.Resource.ID)
	fmt.Printf("  Decision: %v\n", decision3.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision3.Reason)

	// Test Case 4: Viewer reading document
	viewerIdentity := &identity.IdentityContext{
		Subject: &identity.Subject{
			ID:       "user-3",
			TenantID: "demo-tenant",
			Type:     "user",
		},
		TenantID:    "demo-tenant",
		AppID:       "demo-app",
		Roles:       []string{"viewer"},
		Permissions: []string{},
	}

	request4 := &authz.AuthorizationRequest{
		Subject: viewerIdentity,
		Action:  authz.ActionRead,
		Resource: &authz.Resource{
			Type:     "document",
			ID:       "doc-789",
			TenantID: "demo-tenant",
			AppID:    "demo-app",
		},
	}

	decision4, err := evaluator.Evaluate(ctx, request4)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 4 - Viewer reading document:\n")
	fmt.Printf("  Subject: %s (roles: %v)\n", viewerIdentity.Subject.ID, viewerIdentity.Roles)
	fmt.Printf("  Action: %s\n", request4.Action)
	fmt.Printf("  Resource: %s:%s\n", request4.Resource.Type, request4.Resource.ID)
	fmt.Printf("  Decision: %v\n", decision4.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision4.Reason)

	// Test helper functions
	fmt.Printf("=== Testing Helper Functions ===\n")

	hasRole, _ := evaluator.HasRole(ctx, adminIdentity, "admin")
	fmt.Printf("Admin has 'admin' role: %v\n", hasRole)

	hasPermission, _ := evaluator.HasPermission(ctx, editorIdentity, "document:write")
	fmt.Printf("Editor has 'document:write' permission: %v\n", hasPermission)

	hasPermission2, _ := evaluator.HasPermission(ctx, viewerIdentity, "document:delete")
	fmt.Printf("Viewer has 'document:delete' permission: %v\n", hasPermission2)
}
