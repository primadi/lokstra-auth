package main

import (
	"context"
	"fmt"
	"log"
	"time"

	token "github.com/primadi/lokstra-auth/02_token"
	"github.com/primadi/lokstra-auth/02_token/simple"
)

func main() {
	fmt.Println("=== Simple Opaque Token Manager Example ===")
	fmt.Println()

	ctx := context.Background()

	// Create simple token manager with custom config
	config := &simple.Config{
		TokenLength:      32,
		TokenDuration:    1 * time.Hour,
		EnableRevocation: true,
	}

	manager := simple.NewManager(config)

	// Example 1: Generate Token
	fmt.Println("1️⃣  Generating Opaque Token...")
	claims := token.Claims{
		"sub":        "user456",
		"email":      "alice@example.com",
		"name":       "Alice Smith",
		"department": "Engineering",
		"role":       "developer",
	}

	token1, err := manager.Generate(ctx, claims)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Opaque Token Generated:\n")
	fmt.Printf("   Value: %s\n", token1.Value)
	fmt.Printf("   Type: %s\n", token1.Type)
	fmt.Printf("   Issued At: %s\n", token1.IssuedAt.Format(time.RFC3339))
	fmt.Printf("   Expires At: %s\n", token1.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("   Duration: %v\n", time.Until(token1.ExpiresAt).Round(time.Minute))
	fmt.Printf("   Token Type: %s\n", token1.Metadata["token_type"])
	fmt.Println()

	// Example 2: Verify Token
	fmt.Println("2️⃣  Verifying Opaque Token...")
	result, err := manager.Verify(ctx, token1.Value)
	if err != nil {
		log.Fatal(err)
	}

	if result.Valid {
		fmt.Printf("✅ Token Verification Successful!\n")
		fmt.Printf("   Subject: %s\n", result.Claims["sub"])
		fmt.Printf("   Email: %s\n", result.Claims["email"])
		fmt.Printf("   Name: %s\n", result.Claims["name"])
		fmt.Printf("   Department: %s\n", result.Claims["department"])
		fmt.Printf("   Role: %s\n", result.Claims["role"])
		fmt.Println()
	} else {
		fmt.Printf("❌ Verification Failed: %v\n\n", result.Error)
	}

	// Example 3: Generate Multiple Tokens for Same User
	fmt.Println("3️⃣  Generating Multiple Tokens...")
	token2, _ := manager.Generate(ctx, token.Claims{
		"sub":    "user456",
		"device": "mobile",
	})

	token3, _ := manager.Generate(ctx, token.Claims{
		"sub":    "user456",
		"device": "web",
	})

	fmt.Printf("✅ Generated 3 tokens for user456:\n")
	fmt.Printf("   Token 1 (desktop): %s...\n", token1.Value[:20])
	fmt.Printf("   Token 2 (mobile): %s...\n", token2.Value[:20])
	fmt.Printf("   Token 3 (web): %s...\n", token3.Value[:20])
	fmt.Println()

	// Example 4: Verify All Tokens
	fmt.Println("4️⃣  Verifying All Tokens...")
	tokens := []*token.Token{token1, token2, token3}
	for i, tk := range tokens {
		result, _ := manager.Verify(ctx, tk.Value)
		if result.Valid {
			device := "desktop"
			if d, ok := result.Claims["device"]; ok {
				device = d.(string)
			}
			fmt.Printf("   ✅ Token %d (%s) - Valid\n", i+1, device)
		}
	}
	fmt.Println()

	// Example 5: Revoke Token
	fmt.Println("5️⃣  Revoking Token 2 (mobile)...")
	err = manager.Revoke(ctx, token2.Value)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Token Revoked Successfully\n")
	fmt.Println()

	// Example 6: Verify Revoked Token
	fmt.Println("6️⃣  Attempting to Verify Revoked Token...")
	revokedResult, err := manager.Verify(ctx, token2.Value)
	if err != nil {
		log.Fatal(err)
	}

	if !revokedResult.Valid {
		fmt.Printf("❌ Revoked Token Rejected (expected): %v\n", revokedResult.Error)
		fmt.Println()
	}

	// Example 7: Other Tokens Still Valid
	fmt.Println("7️⃣  Verifying Other Tokens...")
	result1, _ := manager.Verify(ctx, token1.Value)
	result3, _ := manager.Verify(ctx, token3.Value)

	if result1.Valid && result3.Valid {
		fmt.Printf("✅ Token 1 and Token 3 still valid\n")
		fmt.Printf("   Token 1: %s\n", result1.Claims["sub"])
		fmt.Printf("   Token 3: %s\n", result3.Claims["sub"])
		fmt.Println()
	}

	// Example 8: Invalid Token
	fmt.Println("8️⃣  Testing Invalid Token...")
	invalidResult, err := manager.Verify(ctx, "invalid-opaque-token-xyz123")
	if err != nil {
		log.Fatal(err)
	}

	if !invalidResult.Valid {
		fmt.Printf("❌ Invalid Token Rejected (expected): %v\n", invalidResult.Error)
		fmt.Println()
	}

	// Example 9: Expired Token Simulation
	fmt.Println("9️⃣  Testing Expired Token...")
	shortConfig := &simple.Config{
		TokenLength:   32,
		TokenDuration: 1 * time.Second,
	}
	shortManager := simple.NewManager(shortConfig)

	shortToken, _ := shortManager.Generate(ctx, claims)
	fmt.Printf("   Token generated with 1 second expiry\n")
	fmt.Printf("   Waiting for expiration...\n")
	time.Sleep(2 * time.Second)

	expiredResult, err := shortManager.Verify(ctx, shortToken.Value)
	if err != nil {
		log.Fatal(err)
	}

	if !expiredResult.Valid {
		fmt.Printf("❌ Expired Token Rejected (expected): %v\n", expiredResult.Error)
		fmt.Println()
	}

	// Example 10: Different Token Lengths
	fmt.Println("🔟 Testing Different Token Lengths...")

	configs := []int{16, 32, 64}
	for _, length := range configs {
		cfg := &simple.Config{
			TokenLength:   length,
			TokenDuration: 1 * time.Hour,
		}
		mgr := simple.NewManager(cfg)
		tk, _ := mgr.Generate(ctx, token.Claims{"sub": "test"})

		// Decode base64 to get actual byte length
		fmt.Printf("   Token Length %d bytes -> %d chars (base64)\n", length, len(tk.Value))
	}
	fmt.Println()

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("✅ Simple Token Features Demonstrated:")
	fmt.Println("   - Opaque token generation")
	fmt.Println("   - Token verification")
	fmt.Println("   - Multiple tokens per user")
	fmt.Println("   - Token revocation")
	fmt.Println("   - Expiry validation")
	fmt.Println("   - Invalid token detection")
	fmt.Println("   - Configurable token length")
	fmt.Println()
	fmt.Println("🔒 Security Features:")
	fmt.Println("   - Cryptographically secure random generation")
	fmt.Println("   - Opaque tokens (no information leakage)")
	fmt.Println("   - Revocation list support")
	fmt.Println("   - Automatic cleanup of expired tokens")
	fmt.Println()
	fmt.Println("⚙️  Configuration:")
	fmt.Printf("   - Token Length: %d bytes\n", config.TokenLength)
	fmt.Printf("   - Token Duration: %v\n", config.TokenDuration)
	fmt.Printf("   - Revocation Enabled: %v\n", config.EnableRevocation)
	fmt.Println()
	fmt.Println("📊 Use Cases:")
	fmt.Println("   - Session tokens")
	fmt.Println("   - API tokens")
	fmt.Println("   - Temporary access tokens")
	fmt.Println("   - Device-specific tokens")
}
