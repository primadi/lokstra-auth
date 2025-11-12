package main

import (
	"context"
	"fmt"
	"log"

	subject "github.com/primadi/lokstra-auth/03_subject"
	"github.com/primadi/lokstra-auth/03_subject/enriched"
	"github.com/primadi/lokstra-auth/03_subject/simple"
)

func main() {
	fmt.Println("=== Enriched Identity Context Example ===")
	fmt.Println()

	ctx := context.Background()

	// Setup base components
	resolver := simple.NewResolver()

	roles := map[string][]string{
		"user123": {"admin", "developer"},
		"user456": {"user"},
	}
	permissions := map[string][]string{
		"user123": {"users:*", "projects:*", "settings:*"},
		"user456": {"projects:read"},
	}
	groups := map[string][]string{
		"user123": {"engineering", "admin-group"},
		"user456": {"engineering"},
	}
	profiles := map[string]map[string]any{
		"user123": {
			"full_name":  "John Doe",
			"department": "Engineering",
			"level":      "Senior",
		},
		"user456": {
			"full_name":  "Jane Smith",
			"department": "Engineering",
			"level":      "Junior",
		},
	}

	roleProvider := simple.NewStaticRoleProvider(roles)
	permProvider := simple.NewStaticPermissionProvider(permissions)
	groupProvider := simple.NewStaticGroupProvider(groups)
	profileProvider := simple.NewStaticProfileProvider(profiles)

	baseBuilder := simple.NewContextBuilder(
		roleProvider,
		permProvider,
		groupProvider,
		profileProvider,
	)

	// Example 1: Attribute Enrichment
	fmt.Println("1️⃣  Attribute Enrichment...")

	claims := map[string]any{
		"sub":      "user123",
		"username": "john_doe",
		"email":    "john@example.com",
		"timezone": "America/Los_Angeles",
		"language": "en-US",
	}

	sub, _ := resolver.Resolve(ctx, claims)

	attrEnricher := enriched.NewAttributeEnricher()
	attrEnricher.AttributeMapping = map[string]string{
		"email":    "user_email",
		"timezone": "user_timezone",
	}

	enrichedBuilder := enriched.NewContextBuilder(baseBuilder, attrEnricher)
	identity, err := enrichedBuilder.Build(ctx, sub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Identity with Attributes:")
	fmt.Printf("   Subject: %s\n", identity.Subject.ID)
	fmt.Printf("   Metadata:\n")
	for key, value := range identity.Metadata {
		fmt.Printf("     %s: %v\n", key, value)
	}
	fmt.Println()

	// Example 2: Role-Based Enrichment
	fmt.Println("2️⃣  Role-Based Enrichment...")

	roleAttributes := map[string]map[string]any{
		"admin": {
			"can_impersonate": true,
			"dashboard_view":  "admin",
			"max_api_calls":   10000,
		},
		"developer": {
			"has_api_access": true,
			"max_api_calls":  1000,
		},
		"user": {
			"dashboard_view": "standard",
			"max_api_calls":  100,
		},
	}

	roleEnricher := enriched.NewRoleBasedEnricher(roleAttributes)

	roleEnrichedBuilder := enriched.NewContextBuilder(baseBuilder, roleEnricher)
	roleIdentity, err := roleEnrichedBuilder.Build(ctx, sub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Identity with Role-Based Attributes:")
	fmt.Printf("   Roles: %v\n", roleIdentity.Roles)
	fmt.Printf("   Can Impersonate: %v\n", roleIdentity.Metadata["can_impersonate"])
	fmt.Printf("   Dashboard View: %v\n", roleIdentity.Metadata["dashboard_view"])
	fmt.Printf("   Max API Calls: %v\n", roleIdentity.Metadata["max_api_calls"])
	fmt.Printf("   Has API Access: %v\n", roleIdentity.Metadata["has_api_access"])
	fmt.Println()

	// Example 3: Session Enrichment
	fmt.Println("3️⃣  Session Enrichment...")

	// Add session info to context
	sessionCtx := context.WithValue(ctx, "session_id", "sess_abc123")
	sessionCtx = context.WithValue(sessionCtx, "ip_address", "192.168.1.100")
	sessionCtx = context.WithValue(sessionCtx, "user_agent", "Mozilla/5.0")

	sessionEnricher := enriched.NewSessionEnricher()
	sessionBuilder := enriched.NewContextBuilder(baseBuilder, sessionEnricher)

	sessionIdentity, err := sessionBuilder.Build(sessionCtx, sub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Identity with Session:")
	if sessionIdentity.Session != nil {
		fmt.Printf("   Session ID: %s\n", sessionIdentity.Session.ID)
		fmt.Printf("   IP Address: %s\n", sessionIdentity.Session.IPAddress)
		fmt.Printf("   User Agent: %s\n", sessionIdentity.Session.UserAgent)
	}
	fmt.Println()

	// Example 4: Chain Multiple Enrichers
	fmt.Println("4️⃣  Chaining Multiple Enrichers...")

	chainEnricher := enriched.NewChainEnricher(
		attrEnricher,
		roleEnricher,
		sessionEnricher,
	)

	chainBuilder := enriched.NewContextBuilder(baseBuilder, chainEnricher)
	chainIdentity, err := chainBuilder.Build(sessionCtx, sub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Fully Enriched Identity:")
	fmt.Printf("   Subject: %s\n", chainIdentity.Subject.ID)
	fmt.Printf("   Roles: %v\n", chainIdentity.Roles)
	fmt.Printf("   Groups: %v\n", chainIdentity.Groups)
	fmt.Printf("   Metadata items: %d\n", len(chainIdentity.Metadata))
	fmt.Printf("   Has Session: %v\n", chainIdentity.Session != nil)
	fmt.Println()

	// Example 5: Custom Enricher
	fmt.Println("5️⃣  Custom Enricher Logic...")

	customEnricher := enriched.NewCustomEnricher(func(ctx context.Context, identity *subject.IdentityContext) error {
		// Add custom business logic
		if identity.HasRole("admin") {
			identity.Metadata["is_super_user"] = true
			identity.Metadata["ui_theme"] = "admin-dark"
		}

		if identity.HasPermission("users:*") {
			identity.Metadata["can_manage_users"] = true
		}

		// Add computed fields
		identity.Metadata["permission_count"] = len(identity.Permissions)
		identity.Metadata["role_count"] = len(identity.Roles)

		return nil
	})

	customBuilder := enriched.NewContextBuilder(baseBuilder, customEnricher)
	customIdentity, err := customBuilder.Build(ctx, sub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Custom Enriched Identity:")
	fmt.Printf("   Is Super User: %v\n", customIdentity.Metadata["is_super_user"])
	fmt.Printf("   UI Theme: %v\n", customIdentity.Metadata["ui_theme"])
	fmt.Printf("   Can Manage Users: %v\n", customIdentity.Metadata["can_manage_users"])
	fmt.Printf("   Permission Count: %v\n", customIdentity.Metadata["permission_count"])
	fmt.Printf("   Role Count: %v\n", customIdentity.Metadata["role_count"])
	fmt.Println()

	// Example 6: Dynamic Enricher Addition
	fmt.Println("6️⃣  Dynamic Enricher Management...")

	dynamicBuilder := enriched.NewContextBuilder(baseBuilder)

	fmt.Println("   Adding enrichers dynamically...")
	dynamicBuilder.AddEnricher(attrEnricher)
	dynamicBuilder.AddEnricher(roleEnricher)

	dynamicIdentity, err := dynamicBuilder.Build(ctx, sub)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Built with %d metadata items\n", len(dynamicIdentity.Metadata))
	fmt.Println()

	// Example 7: Different Subjects
	fmt.Println("7️⃣  Enriching Different Subject Types...")

	userClaims := map[string]any{
		"sub":      "user456",
		"username": "jane_smith",
		"email":    "jane@example.com",
	}

	userSub, _ := resolver.Resolve(ctx, userClaims)
	userIdentity, _ := chainBuilder.Build(sessionCtx, userSub)

	fmt.Println("✅ Regular User Identity:")
	fmt.Printf("   Subject: %s\n", userIdentity.Subject.ID)
	fmt.Printf("   Roles: %v\n", userIdentity.Roles)
	fmt.Printf("   Dashboard View: %v\n", userIdentity.Metadata["dashboard_view"])
	fmt.Printf("   Max API Calls: %v\n", userIdentity.Metadata["max_api_calls"])
	fmt.Println()

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("✅ Enrichment Types:")
	fmt.Println("   - Attribute mapping")
	fmt.Println("   - Role-based attributes")
	fmt.Println("   - Session information")
	fmt.Println("   - Custom business logic")
	fmt.Println("   - Chained enrichers")
	fmt.Println()
	fmt.Println("🔧 Enricher Features:")
	fmt.Println("   - Pluggable architecture")
	fmt.Println("   - Chain multiple enrichers")
	fmt.Println("   - Dynamic enricher addition")
	fmt.Println("   - Custom enrichment logic")
	fmt.Println("   - Non-intrusive enrichment")
	fmt.Println()
	fmt.Println("💡 Use Cases:")
	fmt.Println("   - Add computed fields")
	fmt.Println("   - Inject session data")
	fmt.Println("   - Apply role-based configuration")
	fmt.Println("   - Customize user experience")
	fmt.Println("   - Centralize business logic")
}
