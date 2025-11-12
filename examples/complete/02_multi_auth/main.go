package main

import (
	"context"
	"fmt"
	"log"

	"github.com/primadi/lokstra-auth/01_credential/basic"
	"github.com/primadi/lokstra-auth/01_credential/passwordless"
	"github.com/primadi/lokstra-auth/02_token/jwt"
	subject "github.com/primadi/lokstra-auth/03_subject"
	"github.com/primadi/lokstra-auth/03_subject/simple"
	authz "github.com/primadi/lokstra-auth/04_authz"
	"github.com/primadi/lokstra-auth/04_authz/rbac"
)

// MockEmailSender implements passwordless.TokenSender
type MockEmailSender struct{}

func (s *MockEmailSender) SendMagicLink(ctx context.Context, email, token, link string) error {
	fmt.Printf("📧 [MOCK] Magic link sent to: %s\n", email)
	fmt.Printf("    Token: %s\n", token)
	fmt.Printf("    Link: %s\n", link)
	return nil
}

func (s *MockEmailSender) SendOTP(ctx context.Context, email, otp string) error {
	fmt.Printf("📧 [MOCK] OTP sent to %s: %s\n", email, otp)
	return nil
}

// MockUserResolver implements passwordless.UserResolver
type MockUserResolver struct {
	users map[string]string // email -> userID
}

func (r *MockUserResolver) ResolveByEmail(ctx context.Context, email string) (string, map[string]any, error) {
	if userID, ok := r.users[email]; ok {
		claims := map[string]any{
			"email": email,
			"sub":   userID,
		}
		return userID, claims, nil
	}
	return "", nil, fmt.Errorf("user not found: %s", email)
}

func main() {
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  LOKSTRA AUTH - Multi-Credential Authentication Demo")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("This demo showcases multiple authentication methods:")
	fmt.Println("1. Basic Auth (Username/Password)")
	fmt.Println("2. Passwordless Auth (Magic Link)")
	fmt.Println("3. API Key Auth (coming soon)")
	fmt.Println()

	ctx := context.Background()

	// ========== Layer 1: Configure Multiple Authentication Methods ==========
	fmt.Println("🔐 Layer 1: Configuring Authentication Methods")
	fmt.Println("────────────────────────────────────────────────────────────────")

	// 1. Basic Authentication (Username/Password)
	userProvider := basic.NewInMemoryUserProvider()

	adminHash, _ := basic.HashPassword("Admin@123456")
	devHash, _ := basic.HashPassword("Developer@123")

	userProvider.AddUser(&basic.User{
		ID:           "user-001",
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: adminHash,
		Disabled:     false,
		Metadata: map[string]any{
			"department": "IT",
			"role":       "administrator",
		},
	})

	userProvider.AddUser(&basic.User{
		ID:           "user-002",
		Username:     "developer",
		Email:        "developer@example.com",
		PasswordHash: devHash,
		Disabled:     false,
		Metadata: map[string]any{
			"department": "Engineering",
			"role":       "developer",
		},
	})

	validator := basic.NewValidator(basic.DefaultValidatorConfig())
	basicAuth := basic.NewAuthenticator(userProvider, validator)
	fmt.Println("✅ Basic Authentication configured")

	// 2. Passwordless Authentication (Magic Link & OTP)
	passwordlessConfig := passwordless.DefaultConfig()
	passwordlessConfig.TokenStore = passwordless.NewInMemoryTokenStore()
	passwordlessConfig.TokenGenerator = passwordless.NewDefaultTokenGenerator()
	passwordlessConfig.UserResolver = &MockUserResolver{
		users: map[string]string{
			"developer@example.com": "user-002",
			"admin@example.com":     "user-001",
		},
	}
	passwordlessConfig.TokenSender = &MockEmailSender{}

	passwordlessAuth := passwordless.NewAuthenticator(passwordlessConfig)
	fmt.Println("✅ Passwordless Authentication configured")

	// ========== Layer 2: Token Management ==========
	fmt.Println("\n🎫 Layer 2: Configuring Token Management")
	fmt.Println("────────────────────────────────────────────────────────────────")

	jwtConfig := jwt.DefaultConfig("super-secret-key-change-in-production")
	tokenManager := jwt.NewManager(jwtConfig)
	fmt.Println("✅ JWT Token Manager configured")

	// ========== Layer 3: Subject Resolution ==========
	fmt.Println("\n👤 Layer 3: Configuring Subject Resolution")
	fmt.Println("────────────────────────────────────────────────────────────────")

	subjectResolver := simple.NewResolver()

	roleProvider := simple.NewStaticRoleProvider(map[string][]string{
		"user-001": {"admin", "developer"},
		"user-002": {"developer"},
	})

	permissionProvider := simple.NewStaticPermissionProvider(map[string][]string{
		"user-001": {"users:*", "posts:*", "settings:*"},
		"user-002": {"posts:read", "posts:create", "posts:update"},
	})

	groupProvider := simple.NewStaticGroupProvider(map[string][]string{
		"user-001": {"admins", "engineering"},
		"user-002": {"engineering"},
	})

	profileProvider := simple.NewStaticProfileProvider(map[string]map[string]any{
		"user-001": {
			"display_name": "Admin User",
			"email":        "admin@example.com",
		},
		"user-002": {
			"display_name": "Developer User",
			"email":        "developer@example.com",
		},
	})

	contextBuilder := simple.NewContextBuilder(
		roleProvider,
		permissionProvider,
		groupProvider,
		profileProvider,
	)
	fmt.Println("✅ Subject Resolution configured")

	// ========== Layer 4: Authorization (RBAC) ==========
	fmt.Println("\n🔒 Layer 4: Configuring Authorization")
	fmt.Println("────────────────────────────────────────────────────────────────")

	rolePermissions := map[string][]string{
		"admin": {
			"*", // Admin has all permissions
		},
		"developer": {
			"read:posts",
			"create:posts",
			"update:posts",
		},
	}
	rbacEvaluator := rbac.NewEvaluator(rolePermissions)
	fmt.Println("✅ RBAC Evaluator configured")

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()

	// ========== SCENARIO 1: Basic Authentication ==========
	fmt.Println("📝 SCENARIO 1: Basic Authentication (Username/Password)")
	fmt.Println("────────────────────────────────────────────────────────────────")

	basicCreds := &basic.BasicCredentials{
		Username: "admin",
		Password: "Admin@123456",
	}

	fmt.Printf("🔑 Authenticating with username: %s\n", basicCreds.Username)

	authResult1, err := basicAuth.Authenticate(ctx, basicCreds)
	if err != nil {
		log.Fatalf("❌ Authentication error: %v", err)
	}

	if !authResult1.Success {
		log.Fatalf("❌ Authentication failed: %v", authResult1.Error)
	}

	fmt.Printf("✅ Authenticated as: %s\n", authResult1.Subject)

	// Generate JWT
	token1, _ := tokenManager.Generate(ctx, authResult1.Claims)
	fmt.Printf("🎫 JWT Token generated\n")
	fmt.Printf("   Expires: %v\n", token1.ExpiresAt.Format("2006-01-02 15:04:05"))

	// Build identity
	verifyResult1, _ := tokenManager.Verify(ctx, token1.Value)
	sub1, _ := subjectResolver.Resolve(ctx, verifyResult1.Claims)
	identity1, _ := contextBuilder.Build(ctx, sub1)

	fmt.Printf("👤 Identity built:\n")
	fmt.Printf("   ID: %s\n", identity1.Subject.ID)
	fmt.Printf("   Roles: %v\n", identity1.Roles)
	fmt.Printf("   Permissions: %v\n", identity1.Permissions)

	// Test authorization
	fmt.Println("\n🔐 Testing RBAC Authorization:")
	testAuthorization(ctx, rbacEvaluator, identity1, []authzTest{
		{"users", authz.ActionRead},
		{"users", authz.ActionCreate},
		{"posts", authz.Action("delete")},
		{"settings", authz.Action("update")},
	})

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()

	// ========== SCENARIO 2: Passwordless Authentication (Magic Link) ==========
	fmt.Println("📝 SCENARIO 2: Passwordless Authentication (Magic Link)")
	fmt.Println("────────────────────────────────────────────────────────────────")

	email := "developer@example.com"
	fmt.Printf("📧 Initiating magic link for: %s\n\n", email)

	err = passwordlessAuth.InitiateMagicLink(ctx, email, "user-002", "https://example.com/auth/callback")
	if err != nil {
		log.Printf("❌ Failed to initiate magic link: %v\n", err)
	} else {
		fmt.Println("✅ Magic link initiated successfully")
		fmt.Println("📬 In production:")
		fmt.Println("   - User receives email with magic link")
		fmt.Println("   - User clicks link to authenticate")
		fmt.Println("   - Backend verifies token and completes login")
		fmt.Println()
		fmt.Println("ℹ️  This demo shows the initiation step only")
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()

	// ========== SCENARIO 3: Passwordless Authentication (OTP) ==========
	fmt.Println("📝 SCENARIO 3: Passwordless Authentication (OTP)")
	fmt.Println("────────────────────────────────────────────────────────────────")

	fmt.Printf("📱 Initiating OTP for: %s\n\n", email)

	err = passwordlessAuth.InitiateOTP(ctx, email, "user-002")
	if err != nil {
		log.Printf("❌ Failed to initiate OTP: %v\n", err)
	} else {
		fmt.Println("✅ OTP initiated successfully")
		fmt.Println("📬 In production:")
		fmt.Println("   - User receives SMS/email with 6-digit code")
		fmt.Println("   - User enters code in app")
		fmt.Println("   - Backend verifies OTP and completes login")
		fmt.Println()
		fmt.Println("ℹ️  This demo shows the initiation step only")
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()

	// ========== SCENARIO 4: Different User with Basic Auth ==========
	fmt.Println("📝 SCENARIO 4: Developer Login with Basic Auth")
	fmt.Println("────────────────────────────────────────────────────────────────")

	devCreds := &basic.BasicCredentials{
		Username: "developer",
		Password: "Developer@123",
	}

	fmt.Printf("🔑 Authenticating with username: %s\n", devCreds.Username)

	authResult4, err := basicAuth.Authenticate(ctx, devCreds)
	if err != nil {
		log.Fatalf("❌ Authentication error: %v", err)
	}
	if !authResult4.Success {
		log.Fatalf("❌ Authentication failed: %v", authResult4.Error)
	}

	token4, err := tokenManager.Generate(ctx, authResult4.Claims)
	if err != nil {
		log.Fatalf("❌ Token generation error: %v", err)
	}

	verifyResult4, err := tokenManager.Verify(ctx, token4.Value)
	if err != nil {
		log.Fatalf("❌ Token verification error: %v", err)
	}

	sub4, err := subjectResolver.Resolve(ctx, verifyResult4.Claims)
	if err != nil {
		log.Fatalf("❌ Subject resolution error: %v", err)
	}

	identity4, err := contextBuilder.Build(ctx, sub4)
	if err != nil {
		log.Fatalf("❌ Identity building error: %v", err)
	}

	fmt.Printf("✅ Authenticated as: %s\n", authResult4.Subject)
	fmt.Printf("👤 Identity: %s\n", identity4.Subject.ID)
	fmt.Printf("👥 Roles: %v\n", identity4.Roles)
	fmt.Printf("🔑 Permissions: %v\n", identity4.Permissions)

	fmt.Println("\n🔐 Testing RBAC Authorization:")
	testAuthorization(ctx, rbacEvaluator, identity4, []authzTest{
		{"posts", authz.ActionRead},
		{"posts", authz.ActionCreate},
		{"posts", authz.Action("delete")}, // Should be denied
		{"users", authz.ActionRead},       // Should be denied
	})

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("✅ Multi-Credential Demo Completed Successfully!")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Println("- Demonstrated 3 authentication methods")
	fmt.Println("- Basic Auth: Username/password flow completed")
	fmt.Println("- Passwordless: Magic link & OTP initiation shown")
	fmt.Println("- All authenticated users get proper authorization")
	fmt.Println()
}

type authzTest struct {
	resource string
	action   authz.Action
}

func testAuthorization(ctx context.Context, evaluator *rbac.Evaluator, identity *subject.IdentityContext, tests []authzTest) {
	for _, test := range tests {
		request := &authz.AuthorizationRequest{
			Subject: identity,
			Resource: &authz.Resource{
				Type: test.resource,
				ID:   "123", // Add ID for proper permission matching
			},
			Action: test.action,
		}

		result, err := evaluator.Evaluate(ctx, request)
		if err != nil {
			log.Printf("   ❌ Error: %v\n", err)
			continue
		}

		status := "✅ ALLOW"
		if !result.Allowed {
			status = "❌ DENY "
		}
		fmt.Printf("   %s: %s %s\n", status, test.action, test.resource)
		if result.Reason != "" {
			fmt.Printf("          Reason: %s\n", result.Reason)
		}
	}
}
