package main

import (
	"context"
	"fmt"
	"log"

	lokstraauth "github.com/primadi/lokstra-auth"
	"github.com/primadi/lokstra-auth/01_credential/basic"
	"github.com/primadi/lokstra-auth/02_token/jwt"
	"github.com/primadi/lokstra-auth/03_subject/simple"
	authz "github.com/primadi/lokstra-auth/04_authz"
	"github.com/primadi/lokstra-auth/04_authz/rbac"
)

func main() {
	fmt.Println("=== Lokstra Auth - Runtime Example ===")
	fmt.Println()

	ctx := context.Background()

	// Step 1: Setup user provider (database, in-memory, etc.)
	fmt.Println("Step 1: Setting up user provider...")
	userProvider := basic.NewInMemoryUserProvider()

	passwordHash, err := basic.HashPassword("SecurePass123!")
	if err != nil {
		log.Fatal(err)
	}

	userProvider.AddUser(&basic.User{
		ID:           "user-001",
		Username:     "john.doe",
		Email:        "john.doe@example.com",
		PasswordHash: passwordHash,
		Disabled:     false,
		Metadata: map[string]interface{}{
			"department": "Engineering",
		},
	})
	fmt.Println("✓ User provider ready")
	fmt.Println()

	// Step 2: Build Auth runtime using Builder pattern
	fmt.Println("Step 2: Building Auth runtime...")

	auth := lokstraauth.NewBuilder().
		// Layer 1: Credential authenticators
		WithAuthenticator("basic", basic.NewAuthenticator(
			userProvider,
			basic.NewValidator(basic.DefaultValidatorConfig()),
		)).
		// Layer 2: Token manager
		WithTokenManager(jwt.NewManager(
			jwt.DefaultConfig("my-secret-key-change-in-production"),
		)).
		// Layer 3: Subject resolver and identity builder
		WithSubjectResolver(simple.NewResolver()).
		WithIdentityContextBuilder(simple.NewContextBuilder(
			simple.NewStaticRoleProvider(map[string][]string{
				"user-001": {"developer", "team-lead"},
			}),
			simple.NewStaticPermissionProvider(map[string][]string{
				"user-001": {"read:code", "write:code", "deploy:staging"},
			}),
			simple.NewStaticGroupProvider(map[string][]string{
				"user-001": {"engineering", "backend-team"},
			}),
			simple.NewStaticProfileProvider(map[string]map[string]interface{}{
				"user-001": {
					"display_name": "John Doe",
					"avatar_url":   "https://example.com/avatar.jpg",
				},
			}),
		)).
		// Layer 4: Authorizer
		WithAuthorizer(rbac.NewEvaluator(map[string][]string{
			"developer": {"read:code", "write:code"},
			"team-lead": {"read:code", "write:code", "deploy:staging", "review:pull-requests"},
			"admin":     {"*"},
		})).
		// Configuration
		EnableRefreshToken().
		SetDefaultAuthenticator("basic").
		Build()

	fmt.Println("✓ Auth runtime configured with all 4 layers")
	fmt.Println()

	// Step 3: Login (Layer 1 -> 2 -> 3)
	fmt.Println("Step 3: Performing login...")

	loginResponse, err := auth.Login(ctx, &lokstraauth.LoginRequest{
		Credentials: &basic.BasicCredentials{
			Username: "john.doe",
			Password: "SecurePass123!",
		},
		Metadata: map[string]interface{}{
			"ip_address": "192.168.1.100",
			"user_agent": "Mozilla/5.0",
		},
	})
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Println("✓ Login successful!")
	fmt.Printf("  Access Token: %s...\n", loginResponse.AccessToken.Value[:50])
	if loginResponse.RefreshToken != nil {
		fmt.Printf("  Refresh Token: %s...\n", loginResponse.RefreshToken.Value[:50])
	}
	fmt.Printf("  User: %s (%s)\n", loginResponse.Identity.Subject.Principal, loginResponse.Identity.Subject.ID)
	fmt.Printf("  Roles: %v\n", loginResponse.Identity.Roles)
	fmt.Printf("  Permissions: %v\n", loginResponse.Identity.Permissions)
	fmt.Println()

	// Step 4: Verify token (Layer 2 -> 3)
	fmt.Println("Step 4: Verifying access token...")

	verifyResponse, err := auth.Verify(ctx, &lokstraauth.VerifyRequest{
		Token:                loginResponse.AccessToken.Value,
		BuildIdentityContext: true,
	})
	if err != nil {
		log.Fatalf("Token verification failed: %v", err)
	}

	if !verifyResponse.Valid {
		log.Fatal("Token is invalid")
	}

	fmt.Println("✓ Token is valid")
	fmt.Printf("  Subject: %s\n", verifyResponse.Identity.Subject.ID)
	fmt.Println()

	// Step 5: Authorization checks (Layer 4)
	fmt.Println("Step 5: Checking authorization...")

	// Test 1: Check permission
	canWrite, err := auth.CheckPermission(ctx, loginResponse.Identity, "write:code")
	if err != nil {
		log.Fatalf("Permission check failed: %v", err)
	}
	fmt.Printf("  Can write code: %v\n", canWrite)

	// Test 2: Check role
	isTeamLead, err := auth.CheckRole(ctx, loginResponse.Identity, "team-lead")
	if err != nil {
		log.Fatalf("Role check failed: %v", err)
	}
	fmt.Printf("  Is team lead: %v\n", isTeamLead)

	// Test 3: Full authorization request
	decision, err := auth.Authorize(ctx, &authz.AuthorizationRequest{
		Subject: loginResponse.Identity,
		Resource: &authz.Resource{
			Type: "code",
			ID:   "repository-123",
		},
		Action: authz.ActionWrite,
	})
	if err != nil {
		log.Fatalf("Authorization failed: %v", err)
	}

	fmt.Printf("  Can write to repository-123: %v\n", decision.Allowed)
	if decision.Allowed {
		fmt.Printf("    Reason: %s\n", decision.Reason)
	}
	fmt.Println()

	fmt.Println("=== All operations completed successfully! ===")
}
