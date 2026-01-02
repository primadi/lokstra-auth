package main

import (
	"context"
	"fmt"
	"log"
	"time"

	identity "github.com/primadi/lokstra-auth/identity"
	"github.com/primadi/lokstra-auth/identity/cached"
	"github.com/primadi/lokstra-auth/identity/simple"
)

func main() {
	fmt.Println("=== Cached & Stored Identity Example ===")
	fmt.Println()

	ctx := context.Background()

	// Setup components
	resolver := simple.NewResolver()

	roles := map[string][]string{
		"user123": {"admin", "developer"},
	}
	permissions := map[string][]string{
		"user123": {"users:*", "projects:*"},
	}

	roleProvider := simple.NewStaticRoleProvider(roles)
	permProvider := simple.NewStaticPermissionProvider(permissions)

	baseBuilder := simple.NewContextBuilder(roleProvider, permProvider, nil, nil)

	// Example 1: Cached Subject Resolution
	fmt.Println("1️⃣  Cached Subject Resolution...")

	cache := cached.NewInMemoryCache()
	cachedResolver := cached.NewResolver(resolver, cache, 5*time.Minute)

	claims := map[string]any{
		"sub":      "user123",
		"username": "john_doe",
		"email":    "john@example.com",
	}

	// First resolution - cache miss
	start := time.Now()
	sub1, err := cachedResolver.Resolve(ctx, claims)
	if err != nil {
		log.Fatal(err)
	}
	duration1 := time.Since(start)

	fmt.Printf("✅ First Resolution (cache miss):\n")
	fmt.Printf("   Subject: %s\n", sub1.ID)
	fmt.Printf("   Duration: %v\n", duration1)
	fmt.Println()

	// Second resolution - cache hit
	start = time.Now()
	sub2, err := cachedResolver.Resolve(ctx, claims)
	if err != nil {
		log.Fatal(err)
	}
	duration2 := time.Since(start)

	fmt.Printf("✅ Second Resolution (cache hit):\n")
	fmt.Printf("   Subject: %s\n", sub2.ID)
	fmt.Printf("   Duration: %v\n", duration2)
	fmt.Printf("   Speedup: %.2fx faster\n", float64(duration1)/float64(duration2))
	fmt.Println()

	// Example 2: Cached Identity Context
	fmt.Println("2️⃣  Cached Identity Context Building...")

	cachedBuilder := cached.NewContextBuilder(baseBuilder, cache, 5*time.Minute)

	// First build - cache miss
	start = time.Now()
	identity1, err := cachedBuilder.Build(ctx, sub1)
	if err != nil {
		log.Fatal(err)
	}
	buildDuration1 := time.Since(start)

	fmt.Printf("✅ First Build (cache miss):\n")
	fmt.Printf("   Subject: %s\n", identity1.Subject.ID)
	fmt.Printf("   Roles: %v\n", identity1.Roles)
	fmt.Printf("   Duration: %v\n", buildDuration1)
	fmt.Println()

	// Second build - cache hit
	start = time.Now()
	identity2, err := cachedBuilder.Build(ctx, sub1)
	if err != nil {
		log.Fatal(err)
	}
	buildDuration2 := time.Since(start)

	fmt.Printf("✅ Second Build (cache hit):\n")
	fmt.Printf("   Subject: %s\n", identity2.Subject.ID)
	fmt.Printf("   Roles: %v\n", identity2.Roles)
	fmt.Printf("   Duration: %v\n", buildDuration2)
	fmt.Printf("   Speedup: %.2fx faster\n", float64(buildDuration1)/float64(buildDuration2))
	fmt.Println()

	// Example 3: Cache Invalidation
	fmt.Println("3️⃣  Cache Invalidation...")

	fmt.Println("   Invalidating cache for user123...")
	err = cachedBuilder.Invalidate(ctx, "acme-corp", "acme-corp", "user123")
	if err != nil {
		log.Fatal(err)
	}

	// After invalidation - cache miss again
	start = time.Now()
	_, _ = cachedBuilder.Build(ctx, sub1)
	buildDuration3 := time.Since(start)

	fmt.Printf("✅ After Invalidation (cache miss):\n")
	fmt.Printf("   Duration: %v\n", buildDuration3)
	fmt.Println()

	// Example 4: Identity Store
	fmt.Println("4️⃣  Storing Identity in Session Store...")

	store := identity.NewInMemoryIdentityStore()

	// Store identity with session
	sessionID := "sess_abc123"
	identity1.Session = &identity.SessionInfo{
		ID:        sessionID,
		CreatedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0",
	}

	err = store.Store(ctx, sessionID, identity1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Identity Stored:\n")
	fmt.Printf("   Session ID: %s\n", sessionID)
	fmt.Printf("   Subject: %s\n", identity1.Subject.ID)
	fmt.Println()

	// Example 5: Retrieve from Store
	fmt.Println("5️⃣  Retrieving Identity from Store...")

	retrieved, err := store.Get(ctx, sessionID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Identity Retrieved:\n")
	fmt.Printf("   Subject: %s\n", retrieved.Subject.ID)
	fmt.Printf("   Roles: %v\n", retrieved.Roles)
	fmt.Printf("   Session IP: %s\n", retrieved.Session.IPAddress)
	fmt.Println()

	// Example 6: Multiple Sessions per User
	fmt.Println("6️⃣  Multiple Sessions for Same User...")

	sessions := []struct {
		ID        string
		IPAddress string
		UserAgent string
	}{
		{"sess_desktop_123", "192.168.1.100", "Chrome/Desktop"},
		{"sess_mobile_456", "192.168.1.101", "Safari/iPhone"},
		{"sess_tablet_789", "192.168.1.102", "Safari/iPad"},
	}

	for _, sess := range sessions {
		identity := &identity.IdentityContext{
			Subject: sub1,
			Roles:   identity1.Roles,
			Session: &identity.SessionInfo{
				ID:        sess.ID,
				CreatedAt: time.Now().Unix(),
				ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				IPAddress: sess.IPAddress,
				UserAgent: sess.UserAgent,
			},
		}
		store.Store(ctx, sess.ID, identity)
		fmt.Printf("   ✅ Stored session: %s (%s)\n", sess.ID, sess.UserAgent)
	}
	fmt.Println()

	// Example 7: List Sessions by Subject
	fmt.Println("7️⃣  Listing All Sessions for User...")

	userSessions, err := store.ListBySubject(ctx, "acme-corp", "user123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Found %d active sessions:\n", len(userSessions))
	for i, sess := range userSessions {
		fmt.Printf("   %d. %s - %s (%s)\n",
			i+1,
			sess.Session.ID,
			sess.Session.UserAgent,
			sess.Session.IPAddress,
		)
	}
	fmt.Println()

	// Example 8: List All Active Sessions
	fmt.Println("8️⃣  Listing All Active Sessions...")

	allSessions, err := store.ListSessions(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Total active sessions: %d\n", len(allSessions))
	fmt.Println()

	// Example 9: Update Identity
	fmt.Println("9️⃣  Updating Stored Identity...")

	retrieved.Roles = append(retrieved.Roles, "moderator")
	err = store.Update(ctx, sessionID, retrieved)
	if err != nil {
		log.Fatal(err)
	}

	updated, _ := store.Get(ctx, sessionID)
	fmt.Printf("✅ Identity Updated:\n")
	fmt.Printf("   New Roles: %v\n", updated.Roles)
	fmt.Println()

	// Example 10: Delete Session
	fmt.Println("🔟 Deleting Specific Session...")

	err = store.Delete(ctx, "sess_mobile_456")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Mobile session deleted")

	remainingSessions, _ := store.ListBySubject(ctx, "acme-corp", "user123")
	fmt.Printf("   Remaining sessions: %d\n", len(remainingSessions))
	fmt.Println()

	// Example 11: Logout All Sessions for User
	fmt.Println("1️⃣1️⃣  Logging Out All User Sessions...")

	err = store.DeleteBySubject(ctx, "acme-corp", "user123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ All sessions deleted for user123")

	finalSessions, _ := store.ListBySubject(ctx, "acme-corp", "user123")
	fmt.Printf("   Sessions remaining: %d\n", len(finalSessions))
	fmt.Println()

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("✅ Caching Features:")
	fmt.Println("   - Subject resolution caching")
	fmt.Println("   - Identity context caching")
	fmt.Println("   - TTL-based expiration")
	fmt.Println("   - Cache invalidation")
	fmt.Println("   - Automatic cleanup")
	fmt.Println()
	fmt.Println("✅ Storage Features:")
	fmt.Println("   - Session-based storage")
	fmt.Println("   - Multi-session per user")
	fmt.Println("   - Session listing")
	fmt.Println("   - Identity updates")
	fmt.Println("   - Bulk session deletion")
	fmt.Println()
	fmt.Println("💡 Use Cases:")
	fmt.Println("   - Performance optimization")
	fmt.Println("   - Session management")
	fmt.Println("   - Multi-device login tracking")
	fmt.Println("   - Force logout capabilities")
	fmt.Println("   - Identity lifecycle management")
}
