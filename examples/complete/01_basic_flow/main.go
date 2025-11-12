package main

import (
	"context"
	"fmt"
	"log"

	"github.com/primadi/lokstra-auth/01_credential/basic"
	"github.com/primadi/lokstra-auth/02_token/jwt"
	"github.com/primadi/lokstra-auth/03_subject/simple"
	authz "github.com/primadi/lokstra-auth/04_authz"
	"github.com/primadi/lokstra-auth/04_authz/rbac"
)

func main() {
	fmt.Println("=== Lokstra Auth - Complete Flow Example ===")
	fmt.Println()

	ctx := context.Background()

	// ========== Layer 1: Credential Input / Authentication ==========
	fmt.Println("Layer 1: Setting up authentication...")

	// Create in-memory user provider
	userProvider := basic.NewInMemoryUserProvider()

	// Hash password for test user
	passwordHash, err := basic.HashPassword("SecurePass123!")
	if err != nil {
		log.Fatal(err)
	}

	// Add test user
	userProvider.AddUser(&basic.User{
		ID:           "user-001",
		Username:     "john.doe",
		Email:        "john.doe@example.com",
		PasswordHash: passwordHash,
		Disabled:     false,
		Metadata: map[string]any{
			"department": "Engineering",
		},
	})

	// Create validator and authenticator
	validator := basic.NewValidator(basic.DefaultValidatorConfig())
	authenticator := basic.NewAuthenticator(userProvider, validator)

	// Authenticate user
	credentials := &basic.BasicCredentials{
		Username: "john.doe",
		Password: "SecurePass123!",
	}

	authResult, err := authenticator.Authenticate(ctx, credentials)
	if err != nil {
		log.Fatalf("Authentication error: %v", err)
	}

	if !authResult.Success {
		log.Fatalf("Authentication failed: %v", authResult.Error)
	}

	fmt.Printf("✓ User authenticated: %s\n", authResult.Subject)
	fmt.Printf("  Claims: %v\n\n", authResult.Claims)

	// ========== Layer 2: Token Generation ==========
	fmt.Println("Layer 2: Generating JWT token...")

	// Create JWT manager
	jwtConfig := jwt.DefaultConfig("my-secret-key-change-in-production")
	tokenManager := jwt.NewManager(jwtConfig)

	// Generate access token
	accessToken, err := tokenManager.Generate(ctx, authResult.Claims)
	if err != nil {
		log.Fatalf("Token generation error: %v", err)
	}

	fmt.Printf("✓ Access token generated\n")
	fmt.Printf("  Token: %s...\n", accessToken.Value[:50])
	fmt.Printf("  Expires: %v\n\n", accessToken.ExpiresAt)

	// Verify token
	verifyResult, err := tokenManager.Verify(ctx, accessToken.Value)
	if err != nil {
		log.Fatalf("Token verification error: %v", err)
	}

	if !verifyResult.Valid {
		log.Fatalf("Token verification failed: %v", verifyResult.Error)
	}

	fmt.Printf("✓ Token verified successfully\n")
	fmt.Printf("  Claims extracted: %v\n\n", verifyResult.Claims)

	// ========== Layer 3: Subject Resolution ==========
	fmt.Println("Layer 3: Resolving subject and building identity context...")

	// Create subject resolver
	subjectResolver := simple.NewResolver()

	// Resolve subject from claims
	subject, err := subjectResolver.Resolve(ctx, verifyResult.Claims)
	if err != nil {
		log.Fatalf("Subject resolution error: %v", err)
	}

	fmt.Printf("✓ Subject resolved\n")
	fmt.Printf("  ID: %s\n", subject.ID)
	fmt.Printf("  Type: %s\n", subject.Type)
	fmt.Printf("  Principal: %s\n\n", subject.Principal)

	// Create identity context builder with providers
	roleProvider := simple.NewStaticRoleProvider(map[string][]string{
		"user-001": {"developer", "team-lead"},
	})

	permissionProvider := simple.NewStaticPermissionProvider(map[string][]string{
		"user-001": {"read:code", "write:code", "deploy:staging"},
	})

	groupProvider := simple.NewStaticGroupProvider(map[string][]string{
		"user-001": {"engineering", "backend-team"},
	})

	profileProvider := simple.NewStaticProfileProvider(map[string]map[string]any{
		"user-001": {
			"display_name": "John Doe",
			"avatar_url":   "https://example.com/avatar.jpg",
		},
	})

	contextBuilder := simple.NewContextBuilder(
		roleProvider,
		permissionProvider,
		groupProvider,
		profileProvider,
	)

	// Build identity context
	identity, err := contextBuilder.Build(ctx, subject)
	if err != nil {
		log.Fatalf("Identity context building error: %v", err)
	}

	fmt.Printf("✓ Identity context built\n")
	fmt.Printf("  Roles: %v\n", identity.Roles)
	fmt.Printf("  Permissions: %v\n", identity.Permissions)
	fmt.Printf("  Groups: %v\n", identity.Groups)
	fmt.Printf("  Profile: %v\n\n", identity.Profile)

	// ========== Layer 4: Authorization ==========
	fmt.Println("Layer 4: Checking authorization...")

	// Create RBAC evaluator with role-permission mappings
	rolePermissions := map[string][]string{
		"developer": {
			"read:code",
			"write:code",
		},
		"team-lead": {
			"read:code",
			"write:code",
			"deploy:staging",
			"review:pull-requests",
		},
		"admin": {
			"*", // Admin has all permissions
		},
	}

	rbacEvaluator := rbac.NewEvaluator(rolePermissions)

	// Test authorization for different resources
	testCases := []struct {
		resourceType string
		resourceID   string
		action       authz.Action
	}{
		{"code", "repository-123", authz.ActionRead},
		{"code", "repository-123", authz.ActionWrite},
		{"deployment", "production", authz.Action("deploy")},
	}

	for _, tc := range testCases {
		request := &authz.AuthorizationRequest{
			Subject: identity,
			Resource: &authz.Resource{
				Type: tc.resourceType,
				ID:   tc.resourceID,
			},
			Action: tc.action,
		}

		decision, err := rbacEvaluator.Evaluate(ctx, request)
		if err != nil {
			log.Fatalf("Authorization error: %v", err)
		}

		status := "✗ DENIED"
		if decision.Allowed {
			status = "✓ ALLOWED"
		}

		fmt.Printf("%s: %s %s:%s\n", status, tc.action, tc.resourceType, tc.resourceID)
		fmt.Printf("  Reason: %s\n", decision.Reason)
	}

	fmt.Println("\n=== Complete Flow Finished Successfully! ===")
}
