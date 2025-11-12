package main

import (
	"context"
	"fmt"
	"log"

	subject "github.com/primadi/lokstra-auth/03_subject"
	authz "github.com/primadi/lokstra-auth/04_authz"
	"github.com/primadi/lokstra-auth/04_authz/rbac"
)

func main() {
	fmt.Println("=== RBAC Authorization Example ===")

	ctx := context.Background()

	// Create RBAC evaluator with role permissions
	rolePermissions := map[string][]string{
		"admin":  {"document:*"},
		"editor": {"document:read", "document:write"},
		"viewer": {"document:read"},
	}
	evaluator := rbac.NewEvaluator(rolePermissions)

	fmt.Println("Role Permissions:")
	fmt.Println("- admin: document:*")
	fmt.Println("- editor: document:read, document:write")
	fmt.Println("- viewer: document:read")
	fmt.Println()

	// Test Case 1: Admin user
	adminIdentity := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-1",
			Type: "user",
		},
		Roles:       []string{"admin"},
		Permissions: []string{},
	}

	request1 := &authz.AuthorizationRequest{
		Subject: adminIdentity,
		Action:  authz.Action("document:delete"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-123",
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
	editorIdentity := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-2",
			Type: "user",
		},
		Roles:       []string{"editor"},
		Permissions: []string{},
	}

	request2 := &authz.AuthorizationRequest{
		Subject: editorIdentity,
		Action:  authz.Action("document:write"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-456",
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
		Action:  authz.Action("document:delete"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-456",
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
	viewerIdentity := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-3",
			Type: "user",
		},
		Roles:       []string{"viewer"},
		Permissions: []string{},
	}

	request4 := &authz.AuthorizationRequest{
		Subject: viewerIdentity,
		Action:  authz.Action("document:read"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-789",
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
