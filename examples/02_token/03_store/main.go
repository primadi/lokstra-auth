package main

import (
	"context"
	"fmt"
	"log"
	"time"

	token "github.com/primadi/lokstra-auth/02_token"
	"github.com/primadi/lokstra-auth/02_token/jwt"
)

func main() {
	fmt.Println("=== Token Store Management Example ===")
	fmt.Println()

	ctx := context.Background()

	// Create token store
	store := token.NewInMemoryTokenStore()

	// Create JWT manager
	jwtManager := jwt.NewManager(jwt.DefaultConfig("my-secret-key"))

	// Example 1: Store Tokens for Multiple Users
	fmt.Println("1️⃣  Storing Tokens for Multiple Users...")

	users := []struct {
		ID    string
		Name  string
		Email string
	}{
		{"user1", "Alice", "alice@example.com"},
		{"user2", "Bob", "bob@example.com"},
		{"user3", "Charlie", "charlie@example.com"},
	}

	tokenMap := make(map[string]*token.Token)

	for _, user := range users {
		claims := token.Claims{
			"sub":   user.ID,
			"name":  user.Name,
			"email": user.Email,
		}

		tk, err := jwtManager.Generate(ctx, claims)
		if err != nil {
			log.Fatal(err)
		}

		// Store token
		err = store.Store(ctx, user.ID, tk)
		if err != nil {
			log.Fatal(err)
		}

		tokenMap[user.ID] = tk
		fmt.Printf("   ✅ Stored token for %s (%s)\n", user.Name, user.ID)
	}
	fmt.Println()

	// Example 2: Store Multiple Tokens for Same User (Different Devices)
	fmt.Println("2️⃣  Storing Multiple Tokens for Same User...")

	devices := []string{"desktop", "mobile", "tablet"}

	for _, device := range devices {
		claims := token.Claims{
			"sub":    "user1",
			"device": device,
		}

		tk, _ := jwtManager.Generate(ctx, claims)

		// Add device info to metadata for identification
		tk.Metadata["token_id"] = fmt.Sprintf("user1-%s", device)
		tk.Metadata["device"] = device

		store.Store(ctx, "user1", tk)

		fmt.Printf("   ✅ Stored %s token for user1\n", device)
	}
	fmt.Println()

	// Example 3: Retrieve Token
	fmt.Println("3️⃣  Retrieving Token...")

	desktopToken, err := store.Get(ctx, "user1", "user1-desktop")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Retrieved Desktop Token:\n")
	fmt.Printf("   Token ID: %s\n", desktopToken.Metadata["token_id"])
	fmt.Printf("   Device: %s\n", desktopToken.Metadata["device"])
	fmt.Printf("   Expires: %s\n", desktopToken.ExpiresAt.Format(time.RFC3339))
	fmt.Println()

	// Example 4: List All Tokens for User
	fmt.Println("4️⃣  Listing All Tokens for user1...")

	userTokens, err := store.List(ctx, "user1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Found %d tokens for user1:\n", len(userTokens))
	for i, tk := range userTokens {
		device := "unknown"
		tokenID := "unknown"

		if d, ok := tk.Metadata["device"].(string); ok {
			device = d
		}
		if id, ok := tk.Metadata["token_id"].(string); ok {
			tokenID = id
		}

		fmt.Printf("   %d. %s (ID: %s)\n", i+1, device, tokenID)
	}
	fmt.Println()

	// Example 5: List Tokens for All Users
	fmt.Println("5️⃣  Listing Tokens for All Users...")

	for _, user := range users {
		tokens, _ := store.List(ctx, user.ID)
		fmt.Printf("   %s (%s): %d token(s)\n", user.Name, user.ID, len(tokens))
	}
	fmt.Println()

	// Example 6: Revoke a Specific Token
	fmt.Println("6️⃣  Revoking Mobile Token for user1...")

	err = store.Revoke(ctx, "user1-mobile")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Token Revoked Successfully\n")
	fmt.Println()

	// Example 7: Check Revocation Status
	fmt.Println("7️⃣  Checking Revocation Status...")

	tokensToCheck := []string{"user1-desktop", "user1-mobile", "user1-tablet"}
	for _, tokenID := range tokensToCheck {
		revoked, _ := store.IsRevoked(ctx, tokenID)
		status := "✅ Active"
		if revoked {
			status = "❌ Revoked"
		}
		fmt.Printf("   %s: %s\n", tokenID, status)
	}
	fmt.Println()

	// Example 8: Delete a Token
	fmt.Println("8️⃣  Deleting Tablet Token...")

	err = store.Delete(ctx, "user1", "user1-tablet")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Token Deleted Successfully\n")
	fmt.Println()

	// Example 9: List Remaining Tokens
	fmt.Println("9️⃣  Listing Remaining Tokens for user1...")

	remainingTokens, _ := store.List(ctx, "user1")
	fmt.Printf("✅ Remaining tokens: %d\n", len(remainingTokens))
	for _, tk := range remainingTokens {
		device := tk.Metadata["device"]
		tokenID := tk.Metadata["token_id"]

		status := "Active"
		if tokenID != nil {
			revoked, _ := store.IsRevoked(ctx, tokenID.(string))
			if revoked {
				status = "Revoked"
			}
		}

		fmt.Printf("   - %s (%s)\n", device, status)
	}
	fmt.Println()

	// Example 10: Cleanup Expired Tokens
	fmt.Println("🔟 Testing Token Cleanup...")

	// Create short-lived tokens
	shortConfig := jwt.DefaultConfig("secret")
	shortConfig.AccessTokenDuration = 2 * time.Second
	shortManager := jwt.NewManager(shortConfig)

	for i := 1; i <= 3; i++ {
		claims := token.Claims{"sub": "tempuser", "seq": i}
		tk, _ := shortManager.Generate(ctx, claims)
		tk.Metadata["token_id"] = fmt.Sprintf("temp-%d", i)
		store.Store(ctx, "tempuser", tk)
	}

	fmt.Printf("   Created 3 short-lived tokens (2 second expiry)\n")

	// List before expiry
	beforeCleanup, _ := store.List(ctx, "tempuser")
	fmt.Printf("   Tokens before expiry: %d\n", len(beforeCleanup))

	// Wait for expiration
	fmt.Printf("   Waiting for expiration...\n")
	time.Sleep(3 * time.Second)

	// Cleanup
	store.Cleanup(ctx)
	fmt.Printf("   ✅ Cleanup completed\n")

	// List after cleanup
	afterCleanup, _ := store.List(ctx, "tempuser")
	fmt.Printf("   Tokens after cleanup: %d\n", len(afterCleanup))
	fmt.Println()

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("✅ Token Store Features Demonstrated:")
	fmt.Println("   - Store tokens for multiple users")
	fmt.Println("   - Store multiple tokens per user")
	fmt.Println("   - Retrieve specific tokens")
	fmt.Println("   - List all tokens for a user")
	fmt.Println("   - Revoke tokens")
	fmt.Println("   - Check revocation status")
	fmt.Println("   - Delete tokens")
	fmt.Println("   - Cleanup expired tokens")
	fmt.Println()
	fmt.Println("📊 Current State:")

	for _, user := range users {
		tokens, _ := store.List(ctx, user.ID)
		fmt.Printf("   %s: %d active token(s)\n", user.Name, len(tokens))
	}
	fmt.Println()

	fmt.Println("💡 Use Cases:")
	fmt.Println("   - Session management")
	fmt.Println("   - Multi-device authentication")
	fmt.Println("   - Token lifecycle tracking")
	fmt.Println("   - Security auditing")
	fmt.Println("   - Forced logout (revocation)")
}
