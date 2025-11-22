package main

import (
	"context"
	"fmt"
	"log"

	subject "github.com/primadi/lokstra-auth/03_subject"
	authz "github.com/primadi/lokstra-auth/04_authz"
	"github.com/primadi/lokstra-auth/04_authz/acl"
)

func main() {
	fmt.Printf("=== ACL (Access Control List) Example ===\n")

	ctx := context.Background()
	tenantID := "demo-tenant"
	appID := "demo-app"

	// Create ACL manager
	manager := acl.NewManager()

	// Document 1: Public document
	// Grant read permission to everyone via "public" role
	manager.Grant(ctx, tenantID, appID, "public", &authz.Resource{Type: "document", ID: "doc-public"}, authz.ActionRead)

	// Document 2: Team document
	// Grant full access to team members
	manager.Grant(ctx, tenantID, appID, "user-1", &authz.Resource{Type: "document", ID: "doc-team"}, authz.ActionRead)
	manager.Grant(ctx, tenantID, appID, "user-1", &authz.Resource{Type: "document", ID: "doc-team"}, authz.ActionWrite)
	manager.Grant(ctx, tenantID, appID, "user-1", &authz.Resource{Type: "document", ID: "doc-team"}, authz.ActionDelete)
	manager.Grant(ctx, tenantID, appID, "user-2", &authz.Resource{Type: "document", ID: "doc-team"}, authz.ActionRead)
	manager.Grant(ctx, tenantID, appID, "user-2", &authz.Resource{Type: "document", ID: "doc-team"}, authz.ActionWrite)

	// Document 3: Owner-only document
	manager.Grant(ctx, tenantID, appID, "user-1", &authz.Resource{Type: "document", ID: "doc-private"}, authz.ActionRead)
	manager.Grant(ctx, tenantID, appID, "user-1", &authz.Resource{Type: "document", ID: "doc-private"}, authz.ActionWrite)
	manager.Grant(ctx, tenantID, appID, "user-1", &authz.Resource{Type: "document", ID: "doc-private"}, authz.ActionDelete)

	fmt.Println("ACL Configuration:")
	fmt.Println("- doc-public: public → read")
	fmt.Println("- doc-team: user-1 → read,write,delete | user-2 → read,write")
	fmt.Println("- doc-private: user-1 → read,write,delete")
	fmt.Println()

	// Test Case 1: User-1 accessing team document (write)
	user1Identity := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-1",
			Type: "user",
		},
		Roles: []string{},
	}

	request1 := &authz.AuthorizationRequest{
		Subject: user1Identity,
		Action:  authz.Action("write"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-team",
		},
	}

	decision1, err := manager.Evaluate(ctx, request1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 1 - User-1 writing team document:\n")
	fmt.Printf("  Subject: %s\n", user1Identity.Subject.ID)
	fmt.Printf("  Action: %s\n", request1.Action)
	fmt.Printf("  Resource: %s:%s\n", request1.Resource.Type, request1.Resource.ID)
	fmt.Printf("  Decision: %v\n", decision1.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision1.Reason)

	// Test Case 2: User-2 trying to delete team document (should be denied)
	user2Identity := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-2",
			Type: "user",
		},
		Roles: []string{},
	}

	request2 := &authz.AuthorizationRequest{
		Subject: user2Identity,
		Action:  authz.Action("delete"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-team",
		},
	}

	decision2, err := manager.Evaluate(ctx, request2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 2 - User-2 deleting team document (should fail):\n")
	fmt.Printf("  Subject: %s\n", user2Identity.Subject.ID)
	fmt.Printf("  Action: %s\n", request2.Action)
	fmt.Printf("  Resource: %s:%s\n", request2.Resource.Type, request2.Resource.ID)
	fmt.Printf("  Decision: %v\n", decision2.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision2.Reason)

	// Test Case 3: Editor role accessing team document
	editorIdentity := &subject.IdentityContext{
		Subject: &subject.Subject{
			ID:   "user-3",
			Type: "user",
		},
		Roles: []string{"editor"},
	}

	request3 := &authz.AuthorizationRequest{
		Subject: editorIdentity,
		Action:  authz.Action("write"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-team",
		},
	}

	decision3, err := manager.Evaluate(ctx, request3)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 3 - Editor writing team document:\n")
	fmt.Printf("  Subject: %s (roles: %v)\n", editorIdentity.Subject.ID, editorIdentity.Roles)
	fmt.Printf("  Action: %s\n", request3.Action)
	fmt.Printf("  Resource: %s:%s\n", request3.Resource.Type, request3.Resource.ID)
	fmt.Printf("  Decision: %v\n", decision3.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision3.Reason)

	// Test Case 4: User-1 accessing private document (wildcard permission)
	request4 := &authz.AuthorizationRequest{
		Subject: user1Identity,
		Action:  authz.Action("admin"),
		Resource: &authz.Resource{
			Type: "document",
			ID:   "doc-private",
		},
	}

	decision4, err := manager.Evaluate(ctx, request4)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test 4 - User-1 admin on private document (wildcard):\n")
	fmt.Printf("  Subject: %s\n", user1Identity.Subject.ID)
	fmt.Printf("  Action: %s\n", request4.Action)
	fmt.Printf("  Resource: %s:%s\n", request4.Resource.Type, request4.Resource.ID)
	fmt.Printf("  Decision: %v\n", decision4.Allowed)
	fmt.Printf("  Reason: %s\n\n", decision4.Reason)

	// Test ACL management functions
	fmt.Printf("=== ACL Management Functions ===\n")

	// Get all permissions for user-1 on team document
	perms, _ := manager.GetPermissions(ctx, tenantID, appID, "document", "doc-team", "user-1", user1Identity)
	fmt.Printf("User-1 permissions on doc-team: %v\n", perms)

	// Get all subjects with access to team document
	subjects, _ := manager.GetSubjects(ctx, tenantID, appID, "document", "doc-team")
	fmt.Printf("Subjects with access to doc-team: %v\n", subjects)

	// Revoke write permission from user-2
	manager.Revoke(ctx, tenantID, appID, "user-2", &authz.Resource{Type: "document", ID: "doc-team"}, authz.ActionWrite)
	perms2, _ := manager.GetPermissions(ctx, tenantID, appID, "document", "doc-team", "user-2", user2Identity)
	fmt.Printf("\nAfter revoking write from user-2: %v\n", perms2)

	// Copy ACL from one document to another
	manager.CopyACL(ctx, tenantID, appID, "document", "doc-team", "document", "doc-new")
	perms3, _ := manager.GetPermissions(ctx, tenantID, appID, "document", "doc-new", "user-1", user1Identity)
	fmt.Printf("Copied ACL to doc-new, user-1 permissions: %v\n", perms3)
}
