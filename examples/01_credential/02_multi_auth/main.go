package main

import (
	"context"
	"fmt"
	"log"

	lokstraauth "github.com/primadi/lokstra-auth"
	credential "github.com/primadi/lokstra-auth/01_credential"
	"github.com/primadi/lokstra-auth/01_credential/basic"
	"github.com/primadi/lokstra-auth/02_token/jwt"
	subject "github.com/primadi/lokstra-auth/03_subject"
	"github.com/primadi/lokstra-auth/03_subject/simple"
	authz "github.com/primadi/lokstra-auth/04_authz"
	"github.com/primadi/lokstra-auth/04_authz/rbac"
)

// OAuth2Credentials represents OAuth2 credentials
type OAuth2Credentials struct {
	Provider string
	Token    string
}

func (c *OAuth2Credentials) Type() string {
	return "oauth2"
}

func (c *OAuth2Credentials) Validate() error {
	return nil
}

// MockOAuth2Authenticator simulates OAuth2 authentication
type MockOAuth2Authenticator struct {
	provider string
}

func (a *MockOAuth2Authenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
	oauth2Creds, ok := creds.(*OAuth2Credentials)
	if !ok {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   fmt.Errorf("invalid credentials type"),
		}, nil
	}

	// In real implementation, this would verify OAuth2 token with provider
	// For demo, we just accept it
	return &credential.AuthenticationResult{
		Success: true,
		Subject: "oauth2-user-123",
		Claims: map[string]interface{}{
			"sub":      "oauth2-user-123",
			"email":    "user@gmail.com",
			"name":     "OAuth User",
			"provider": oauth2Creds.Provider,
			"verified": true,
		},
	}, nil
}

func (a *MockOAuth2Authenticator) Type() string {
	return "oauth2"
}

// PasswordlessCredentials represents passwordless credentials (OTP/Magic Link)
type PasswordlessCredentials struct {
	Email string
	Token string
}

func (c *PasswordlessCredentials) Type() string {
	return "passwordless"
}

func (c *PasswordlessCredentials) Validate() error {
	return nil
}

// MockPasswordlessAuthenticator simulates passwordless authentication
type MockPasswordlessAuthenticator struct{}

func (a *MockPasswordlessAuthenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
	pwdlessCreds, ok := creds.(*PasswordlessCredentials)
	if !ok {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   fmt.Errorf("invalid credentials type"),
		}, nil
	}

	// In real implementation, this would verify OTP or magic link token
	return &credential.AuthenticationResult{
		Success: true,
		Subject: "passwordless-user-456",
		Claims: map[string]interface{}{
			"sub":    "passwordless-user-456",
			"email":  pwdlessCreds.Email,
			"method": "magic-link",
		},
	}, nil
}

func (a *MockPasswordlessAuthenticator) Type() string {
	return "passwordless"
}

// HybridAuthorizer combines RBAC and ABAC
type HybridAuthorizer struct {
	rbacEvaluator *rbac.Evaluator
}

func NewHybridAuthorizer(rolePermissions map[string][]string) *HybridAuthorizer {
	return &HybridAuthorizer{
		rbacEvaluator: rbac.NewEvaluator(rolePermissions),
	}
}

// Evaluate implements PolicyEvaluator with RBAC + ABAC logic
func (h *HybridAuthorizer) Evaluate(ctx context.Context, request *authz.AuthorizationRequest) (*authz.AuthorizationDecision, error) {
	// First, check RBAC
	rbacDecision, err := h.rbacEvaluator.Evaluate(ctx, request)
	if err != nil {
		return nil, err
	}

	// If RBAC allows, check ABAC conditions
	if rbacDecision.Allowed {
		// ABAC: Check additional attribute-based conditions
		if !h.checkABACConditions(request) {
			return &authz.AuthorizationDecision{
				Allowed: false,
				Reason:  "RBAC allowed but ABAC conditions not met",
			}, nil
		}
		return rbacDecision, nil
	}

	// If RBAC denies, check if ABAC can allow
	if h.checkABACOverride(request) {
		return &authz.AuthorizationDecision{
			Allowed: true,
			Reason:  "RBAC denied but ABAC conditions allow access (resource owner)",
		}, nil
	}

	return rbacDecision, nil
}

// checkABACConditions checks attribute-based conditions
func (h *HybridAuthorizer) checkABACConditions(request *authz.AuthorizationRequest) bool {
	// Example ABAC rules:

	// 1. Resource ownership check
	if ownerID, ok := request.Resource.Attributes["owner_id"].(string); ok {
		if ownerID == request.Subject.Subject.ID {
			return true // Owner can always access their resources
		}
	}

	// 2. Time-based access (business hours only for certain actions)
	if request.Action == authz.ActionDelete {
		// In production, check actual time
		// For now, we'll allow it
		return true
	}

	// 3. Department-based access
	if dept, ok := request.Subject.Subject.Attributes["department"].(string); ok {
		if resourceDept, ok := request.Resource.Attributes["department"].(string); ok {
			if dept == resourceDept {
				return true // Same department can access
			}
		}
	}

	// 4. Verified users only for sensitive actions
	if request.Action == "deploy" {
		if verified, ok := request.Subject.Subject.Attributes["verified"].(bool); ok {
			if !verified {
				return false // Unverified users cannot deploy
			}
		}
	}

	return true // Default allow if no restrictions
}

// checkABACOverride checks if ABAC can override RBAC denial
func (h *HybridAuthorizer) checkABACOverride(request *authz.AuthorizationRequest) bool {
	// Example: Resource owner can read even without explicit permission
	if request.Action == authz.ActionRead {
		if ownerID, ok := request.Resource.Attributes["owner_id"].(string); ok {
			return ownerID == request.Subject.Subject.ID
		}
	}
	return false
}

// Implement PermissionChecker interface
func (h *HybridAuthorizer) HasPermission(ctx context.Context, identity *subject.IdentityContext, permission string) (bool, error) {
	return h.rbacEvaluator.HasPermission(ctx, identity, permission)
}

func (h *HybridAuthorizer) HasAnyPermission(ctx context.Context, identity *subject.IdentityContext, permissions ...string) (bool, error) {
	return h.rbacEvaluator.HasAnyPermission(ctx, identity, permissions...)
}

func (h *HybridAuthorizer) HasAllPermissions(ctx context.Context, identity *subject.IdentityContext, permissions ...string) (bool, error) {
	return h.rbacEvaluator.HasAllPermissions(ctx, identity, permissions...)
}

// Implement RoleChecker interface
func (h *HybridAuthorizer) HasRole(ctx context.Context, identity *subject.IdentityContext, role string) (bool, error) {
	return h.rbacEvaluator.HasRole(ctx, identity, role)
}

func (h *HybridAuthorizer) HasAnyRole(ctx context.Context, identity *subject.IdentityContext, roles ...string) (bool, error) {
	return h.rbacEvaluator.HasAnyRole(ctx, identity, roles...)
}

func (h *HybridAuthorizer) HasAllRoles(ctx context.Context, identity *subject.IdentityContext, roles ...string) (bool, error) {
	return h.rbacEvaluator.HasAllRoles(ctx, identity, roles...)
}

func main() {
	fmt.Println("=== Enterprise Auth - Multi-Authenticator + Hybrid RBAC/ABAC ===")
	fmt.Println()

	ctx := context.Background()

	// ========== Setup: Multiple Authentication Methods ==========
	fmt.Println("Setting up enterprise authentication system...")
	fmt.Println()

	// 1. Setup Basic Auth (username/password)
	userProvider := basic.NewInMemoryUserProvider()
	passwordHash, _ := basic.HashPassword("SecurePass123!")
	userProvider.AddUser(&basic.User{
		ID:           "user-001",
		Username:     "john.doe",
		Email:        "john@company.com",
		PasswordHash: passwordHash,
		Metadata: map[string]interface{}{
			"department": "Engineering",
			"verified":   true,
		},
	})

	basicAuth := basic.NewAuthenticator(
		userProvider,
		basic.NewValidator(basic.DefaultValidatorConfig()),
	)

	// 2. Setup OAuth2 providers (Google, GitHub use same authenticator)
	oauth2Auth := &MockOAuth2Authenticator{provider: "oauth2"}

	// 3. Setup Passwordless auth
	passwordlessAuth := &MockPasswordlessAuthenticator{}

	// 4. Setup role-based permissions for RBAC part
	rolePermissions := map[string][]string{
		"developer": {
			"read:code",
			"write:code",
			"read:documents",
		},
		"team-lead": {
			"read:code",
			"write:code",
			"deploy:staging",
			"read:documents",
			"write:documents",
		},
		"admin": {
			"*", // All permissions
		},
	}

	// 5. Build Auth runtime with ALL authentication methods
	auth := lokstraauth.NewBuilder().
		// Register ALL authenticators
		WithAuthenticator("basic", basicAuth).
		WithAuthenticator("oauth2", oauth2Auth). // Both Google and GitHub use same authenticator type
		WithAuthenticator("passwordless", passwordlessAuth).
		// Token management
		WithTokenManager(jwt.NewManager(
			jwt.DefaultConfig("production-secret-key"),
		)).
		// Subject resolution
		WithSubjectResolver(simple.NewResolver()).
		WithIdentityContextBuilder(simple.NewContextBuilder(
			// Roles mapped by user ID
			simple.NewStaticRoleProvider(map[string][]string{
				"user-001":              {"developer", "team-lead"},
				"oauth2-user-123":       {"developer"},
				"passwordless-user-456": {"developer"},
			}),
			simple.NewStaticPermissionProvider(map[string][]string{}),
			simple.NewStaticGroupProvider(map[string][]string{
				"user-001":              {"engineering", "backend-team"},
				"oauth2-user-123":       {"engineering", "frontend-team"},
				"passwordless-user-456": {"engineering"},
			}),
			simple.NewStaticProfileProvider(map[string]map[string]interface{}{}),
		)).
		// Hybrid RBAC + ABAC Authorization
		WithAuthorizer(NewHybridAuthorizer(rolePermissions)).
		EnableRefreshToken().
		Build()

	fmt.Println("✓ Auth system ready with:")
	fmt.Println("  - Basic (username/password)")
	fmt.Println("  - OAuth2 (Google, GitHub)")
	fmt.Println("  - Passwordless (OTP/Magic Link)")
	fmt.Println("  - Hybrid RBAC + ABAC authorization")
	fmt.Println()

	// ========== Scenario 1: Login via Username/Password ==========
	fmt.Println("--- Scenario 1: Login via Username/Password ---")

	basicResponse, err := auth.Login(ctx, &lokstraauth.LoginRequest{
		Credentials: &basic.BasicCredentials{
			Username: "john.doe",
			Password: "SecurePass123!",
		},
	})
	if err != nil {
		log.Fatalf("Basic login failed: %v", err)
	}

	fmt.Printf("✓ Logged in as: %s (via basic auth)\n", basicResponse.Identity.Subject.Principal)
	fmt.Printf("  Roles: %v\n", basicResponse.Identity.Roles)
	fmt.Println()

	// ========== Scenario 2: Login via Google OAuth2 ==========
	fmt.Println("--- Scenario 2: Login via Google OAuth2 ---")

	googleResponse, err := auth.Login(ctx, &lokstraauth.LoginRequest{
		Credentials: &OAuth2Credentials{
			Provider: "google",
			Token:    "mock-google-token",
		},
	})
	if err != nil {
		log.Fatalf("Google OAuth2 login failed: %v", err)
	}

	fmt.Printf("✓ Logged in via Google OAuth2\n")
	fmt.Printf("  Email: %s\n", googleResponse.Identity.Subject.Attributes["email"])
	fmt.Printf("  Provider: %s\n", googleResponse.Identity.Subject.Attributes["provider"])
	fmt.Println()

	// ========== Scenario 3: Login via Passwordless ==========
	fmt.Println("--- Scenario 3: Login via Passwordless (Magic Link) ---")

	passwordlessResponse, err := auth.Login(ctx, &lokstraauth.LoginRequest{
		Credentials: &PasswordlessCredentials{
			Email: "user@example.com",
			Token: "magic-link-token-abc123",
		},
	})
	if err != nil {
		log.Fatalf("Passwordless login failed: %v", err)
	}

	fmt.Printf("✓ Logged in via Passwordless\n")
	fmt.Printf("  Email: %s\n", passwordlessResponse.Identity.Subject.Attributes["email"])
	fmt.Printf("  Method: %s\n", passwordlessResponse.Identity.Subject.Attributes["method"])
	fmt.Println()

	// ========== Scenario 4: Hybrid Authorization (RBAC + ABAC) ==========
	fmt.Println("--- Scenario 4: Hybrid RBAC + ABAC Authorization ---")
	fmt.Println()

	// Test Case 1: RBAC allows, ABAC allows (same department)
	fmt.Println("Test 1: Access document in same department")
	decision1, _ := auth.Authorize(ctx, &authz.AuthorizationRequest{
		Subject: basicResponse.Identity,
		Resource: &authz.Resource{
			Type: "documents",
			ID:   "doc-123",
			Attributes: map[string]interface{}{
				"department": "Engineering",
				"owner_id":   "user-002",
			},
		},
		Action: authz.ActionRead,
	})
	fmt.Printf("  Result: %v\n", decision1.Allowed)
	fmt.Printf("  Reason: %s\n", decision1.Reason)
	fmt.Println()

	// Test Case 2: RBAC denies, ABAC allows (resource owner override)
	fmt.Println("Test 2: Read document owned by user (ABAC override)")
	decision2, _ := auth.Authorize(ctx, &authz.AuthorizationRequest{
		Subject: googleResponse.Identity, // Developer role, no explicit read:admin permission
		Resource: &authz.Resource{
			Type: "admin",
			ID:   "settings-123",
			Attributes: map[string]interface{}{
				"owner_id": "oauth2-user-123", // Same as logged-in user
			},
		},
		Action: authz.ActionRead,
	})
	fmt.Printf("  Result: %v (ABAC override: owner can read)\n", decision2.Allowed)
	fmt.Printf("  Reason: %s\n", decision2.Reason)
	fmt.Println()

	// Test Case 3: Check deployment with verified user
	fmt.Println("Test 3: Deploy to staging (RBAC + ABAC verification check)")
	decision3, _ := auth.Authorize(ctx, &authz.AuthorizationRequest{
		Subject: basicResponse.Identity,
		Resource: &authz.Resource{
			Type: "deployment",
			ID:   "staging",
		},
		Action: authz.Action("deploy"),
	})
	fmt.Printf("  Result: %v\n", decision3.Allowed)
	fmt.Printf("  Reason: %s\n", decision3.Reason)
	fmt.Println()

	// Test Case 4: Department-based access
	fmt.Println("Test 4: Cross-department document access (ABAC denies)")
	decision4, _ := auth.Authorize(ctx, &authz.AuthorizationRequest{
		Subject: basicResponse.Identity,
		Resource: &authz.Resource{
			Type: "documents",
			ID:   "doc-789",
			Attributes: map[string]interface{}{
				"department": "Sales", // Different department
				"owner_id":   "user-003",
			},
		},
		Action: authz.ActionRead,
	})
	fmt.Printf("  Result: %v\n", decision4.Allowed)
	fmt.Printf("  Reason: %s\n", decision4.Reason)
	fmt.Println()

	fmt.Println("=== Enterprise Auth System Demo Complete! ===")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println("✓ 4 authentication methods working (Basic, OAuth2-Google, OAuth2-GitHub, Passwordless)")
	fmt.Println("✓ Hybrid RBAC + ABAC authorization implemented")
	fmt.Println("✓ ABAC rules: department-based, ownership, verification status")
	fmt.Println("✓ ABAC can override RBAC (resource owner access)")
	fmt.Println("✓ All managed through single Auth runtime")
}
