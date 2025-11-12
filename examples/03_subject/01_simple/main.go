package main

import (
	"context"
	"fmt"
	"log"

	"github.com/primadi/lokstra-auth/03_subject/simple"
)

func main() {
	fmt.Println("=== Simple Subject Resolution Example ===")
	fmt.Println()

	ctx := context.Background()

	// Example 1: Basic Subject Resolution
	fmt.Println("1️⃣  Basic Subject Resolution...")

	resolver := simple.NewResolver()

	// Claims from JWT token
	claims := map[string]any{
		"sub":      "user123",
		"username": "john_doe",
		"email":    "john@example.com",
		"name":     "John Doe",
		"type":     "user",
	}

	sub, err := resolver.Resolve(ctx, claims)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Subject Resolved:")
	fmt.Printf("   ID: %s\n", sub.ID)
	fmt.Printf("   Type: %s\n", sub.Type)
	fmt.Printf("   Principal: %s\n", sub.Principal)
	fmt.Printf("   Attributes: %d items\n", len(sub.Attributes))
	fmt.Println()

	// Example 2: Subject with Custom Type
	fmt.Println("2️⃣  Service Account Subject...")

	serviceClaims := map[string]any{
		"sub":  "service-api-backend",
		"type": "service",
		"name": "Backend API Service",
		"env":  "production",
	}

	serviceSub, err := resolver.Resolve(ctx, serviceClaims)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Service Subject Resolved:")
	fmt.Printf("   ID: %s\n", serviceSub.ID)
	fmt.Printf("   Type: %s\n", serviceSub.Type)
	fmt.Printf("   Principal: %s\n", serviceSub.Principal)
	fmt.Printf("   Environment: %s\n", serviceSub.Attributes["env"])
	fmt.Println()

	// Example 3: Building Identity Context
	fmt.Println("3️⃣  Building Identity Context...")

	// Setup providers
	roles := map[string][]string{
		"user123": {"admin", "developer"},
	}
	permissions := map[string][]string{
		"user123": {"users:read", "users:write", "projects:read", "projects:write"},
	}
	groups := map[string][]string{
		"user123": {"engineering", "admin-group"},
	}
	profiles := map[string]map[string]any{
		"user123": {
			"full_name":  "John Doe",
			"department": "Engineering",
			"location":   "San Francisco",
			"avatar":     "https://example.com/avatar.jpg",
		},
	}

	roleProvider := simple.NewStaticRoleProvider(roles)
	permProvider := simple.NewStaticPermissionProvider(permissions)
	groupProvider := simple.NewStaticGroupProvider(groups)
	profileProvider := simple.NewStaticProfileProvider(profiles)

	builder := simple.NewContextBuilder(
		roleProvider,
		permProvider,
		groupProvider,
		profileProvider,
	)

	identity, err := builder.Build(ctx, sub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Identity Context Built:")
	fmt.Printf("   Subject ID: %s\n", identity.Subject.ID)
	fmt.Printf("   Roles: %v\n", identity.Roles)
	fmt.Printf("   Permissions: %d items\n", len(identity.Permissions))
	fmt.Printf("   Groups: %v\n", identity.Groups)
	fmt.Printf("   Profile: %d fields\n", len(identity.Profile))
	fmt.Println()

	// Example 4: Role and Permission Checks
	fmt.Println("4️⃣  Authorization Checks...")

	fmt.Printf("   Has 'admin' role: %v\n", identity.HasRole("admin"))
	fmt.Printf("   Has 'user' role: %v\n", identity.HasRole("user"))
	fmt.Printf("   Has 'users:write' permission: %v\n", identity.HasPermission("users:write"))
	fmt.Printf("   Has 'users:delete' permission: %v\n", identity.HasPermission("users:delete"))
	fmt.Println()

	// Example 5: Advanced Role Checks
	fmt.Println("5️⃣  Advanced Authorization Checks...")

	hasAnyAdminOrDev := identity.HasAnyRole("admin", "developer")
	fmt.Printf("   Has any role (admin OR developer): %v\n", hasAnyAdminOrDev)

	hasAllRoles := identity.HasAllRoles("admin", "developer")
	fmt.Printf("   Has all roles (admin AND developer): %v\n", hasAllRoles)

	hasAllManagerRoles := identity.HasAllRoles("admin", "manager")
	fmt.Printf("   Has all roles (admin AND manager): %v\n", hasAllManagerRoles)
	fmt.Println()

	// Example 6: Subject without Providers
	fmt.Println("6️⃣  Minimal Identity Context...")

	minimalBuilder := simple.NewContextBuilder(nil, nil, nil, nil)
	minimalIdentity, err := minimalBuilder.Build(ctx, serviceSub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Minimal Identity:")
	fmt.Printf("   Subject: %s\n", minimalIdentity.Subject.ID)
	fmt.Printf("   Roles: %v (empty)\n", minimalIdentity.Roles)
	fmt.Printf("   Permissions: %v (empty)\n", minimalIdentity.Permissions)
	fmt.Println()

	// Example 7: Custom Resolver Configuration
	fmt.Println("7️⃣  Custom Resolver Configuration...")

	customResolver := simple.NewResolver()
	customResolver.SubjectIDClaim = "user_id"
	customResolver.PrincipalClaim = "email"
	customResolver.DefaultSubjectType = "customer"

	customClaims := map[string]any{
		"user_id": "cust456",
		"email":   "alice@customer.com",
		"plan":    "premium",
	}

	customSub, err := customResolver.Resolve(ctx, customClaims)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Custom Subject Resolved:")
	fmt.Printf("   ID: %s\n", customSub.ID)
	fmt.Printf("   Type: %s\n", customSub.Type)
	fmt.Printf("   Principal: %s\n", customSub.Principal)
	fmt.Printf("   Plan: %s\n", customSub.Attributes["plan"])
	fmt.Println()

	// Example 8: Profile Access
	fmt.Println("8️⃣  Accessing Profile Data...")

	if identity.Profile != nil {
		fmt.Println("✅ Profile Information:")
		for key, value := range identity.Profile {
			fmt.Printf("   %s: %v\n", key, value)
		}
	}
	fmt.Println()

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("✅ Subject Resolution Features:")
	fmt.Println("   - Resolve subjects from claims")
	fmt.Println("   - Support multiple subject types")
	fmt.Println("   - Configurable claim mapping")
	fmt.Println("   - Extract all claims as attributes")
	fmt.Println()
	fmt.Println("✅ Identity Context Features:")
	fmt.Println("   - Role management")
	fmt.Println("   - Permission management")
	fmt.Println("   - Group membership")
	fmt.Println("   - Profile information")
	fmt.Println("   - Authorization helpers")
	fmt.Println()
	fmt.Println("💡 Use Cases:")
	fmt.Println("   - User authentication")
	fmt.Println("   - Service-to-service auth")
	fmt.Println("   - Role-based access control")
	fmt.Println("   - Permission-based access control")
	fmt.Println("   - Profile management")
}
