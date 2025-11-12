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
	fmt.Println("=== JWT Token Manager Example ===")
	fmt.Println()

	ctx := context.Background()

	// Create JWT manager with default config
	config := jwt.DefaultConfig("my-secret-key-change-in-production")
	config.AccessTokenDuration = 15 * time.Minute
	config.RefreshTokenDuration = 7 * 24 * time.Hour
	config.EnableRevocation = true

	manager := jwt.NewManager(config)

	// Example 1: Generate Access Token
	fmt.Println("1️⃣  Generating JWT Access Token...")
	claims := token.Claims{
		"sub":   "user123",
		"email": "john@example.com",
		"name":  "John Doe",
		"role":  "admin",
	}

	accessToken, err := manager.Generate(ctx, claims)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Access Token Generated:\n")
	fmt.Printf("   Value: %s...\n", accessToken.Value[:50])
	fmt.Printf("   Type: %s\n", accessToken.Type)
	fmt.Printf("   Issued At: %s\n", accessToken.IssuedAt.Format(time.RFC3339))
	fmt.Printf("   Expires At: %s\n", accessToken.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("   Duration: %v\n", time.Until(accessToken.ExpiresAt).Round(time.Second))
	fmt.Printf("   Algorithm: %s\n", accessToken.Metadata["algorithm"])
	fmt.Println()

	// Example 2: Verify Access Token
	fmt.Println("2️⃣  Verifying JWT Access Token...")
	result, err := manager.Verify(ctx, accessToken.Value)
	if err != nil {
		log.Fatal(err)
	}

	if result.Valid {
		fmt.Printf("✅ Token Verification Successful!\n")
		fmt.Printf("   Subject: %s\n", result.Claims["sub"])
		fmt.Printf("   Email: %s\n", result.Claims["email"])
		fmt.Printf("   Name: %s\n", result.Claims["name"])
		fmt.Printf("   Role: %s\n", result.Claims["role"])
		fmt.Printf("   Issuer: %s\n", result.Claims["iss"])

		// Extract expiry using helper
		if exp, ok := result.Claims.GetInt64("exp"); ok {
			expTime := time.Unix(exp, 0)
			fmt.Printf("   Expires: %s\n", expTime.Format(time.RFC3339))
		}
		fmt.Println()
	} else {
		fmt.Printf("❌ Verification Failed: %v\n\n", result.Error)
	}

	// Example 3: Generate Refresh Token
	fmt.Println("3️⃣  Generating JWT Refresh Token...")
	refreshToken, err := manager.GenerateRefreshToken(ctx, claims)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Refresh Token Generated:\n")
	fmt.Printf("   Value: %s...\n", refreshToken.Value[:50])
	fmt.Printf("   Expires At: %s\n", refreshToken.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("   Duration: %v\n", time.Until(refreshToken.ExpiresAt).Round(time.Hour))
	fmt.Printf("   Type: %s\n", refreshToken.Metadata["type"])
	fmt.Println()

	// Example 4: Verify Refresh Token
	fmt.Println("4️⃣  Verifying Refresh Token...")
	refreshResult, err := manager.Verify(ctx, refreshToken.Value)
	if err != nil {
		log.Fatal(err)
	}

	if refreshResult.Valid {
		fmt.Printf("✅ Refresh Token Valid!\n")
		fmt.Printf("   Subject: %s\n", refreshResult.Claims["sub"])
		fmt.Printf("   Token Type: %s\n", refreshResult.Claims["type"])
		fmt.Println()
	}

	// Example 5: Refresh Access Token
	fmt.Println("5️⃣  Using Refresh Token to Get New Access Token...")
	newAccessToken, err := manager.Refresh(ctx, refreshToken.Value)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ New Access Token Generated from Refresh Token!\n")
	fmt.Printf("   Value: %s...\n", newAccessToken.Value[:50])
	fmt.Printf("   Expires: %s\n", newAccessToken.ExpiresAt.Format(time.RFC3339))
	fmt.Println()

	// Example 6: Token Revocation
	fmt.Println("6️⃣  Revoking Access Token...")
	err = manager.Revoke(ctx, accessToken.Value)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Token Revoked Successfully\n")
	fmt.Println()

	// Example 7: Verify Revoked Token
	fmt.Println("7️⃣  Attempting to Verify Revoked Token...")
	revokedResult, err := manager.Verify(ctx, accessToken.Value)
	if err != nil {
		log.Fatal(err)
	}

	if !revokedResult.Valid {
		fmt.Printf("❌ Verification Failed (expected): %v\n", revokedResult.Error)
		fmt.Println()
	}

	// Example 8: Invalid Token
	fmt.Println("8️⃣  Testing Invalid Token...")
	invalidResult, err := manager.Verify(ctx, "invalid.jwt.token")
	if err != nil {
		log.Fatal(err)
	}

	if !invalidResult.Valid {
		fmt.Printf("❌ Invalid Token Rejected (expected): %v\n", invalidResult.Error)
		fmt.Println()
	}

	// Example 9: Expired Token Simulation
	fmt.Println("9️⃣  Testing Expired Token...")
	shortConfig := jwt.DefaultConfig("my-secret-key")
	shortConfig.AccessTokenDuration = 1 * time.Second
	shortManager := jwt.NewManager(shortConfig)

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

	// Example 10: Claims Helper Methods
	fmt.Println("🔟 Using Claims Helper Methods...")
	testClaims := token.Claims{
		"string_value": "hello",
		"int_value":    int64(42),
		"bool_value":   true,
		"array_value":  []string{"admin", "user"},
	}

	if strVal, ok := testClaims.GetString("string_value"); ok {
		fmt.Printf("   String: %s\n", strVal)
	}

	if intVal, ok := testClaims.GetInt64("int_value"); ok {
		fmt.Printf("   Int64: %d\n", intVal)
	}

	if boolVal, ok := testClaims.GetBool("bool_value"); ok {
		fmt.Printf("   Bool: %v\n", boolVal)
	}

	if arrVal, ok := testClaims.GetStringSlice("array_value"); ok {
		fmt.Printf("   Array: %v\n", arrVal)
	}
	fmt.Println()

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("✅ JWT Token Features Demonstrated:")
	fmt.Println("   - Access token generation and verification")
	fmt.Println("   - Refresh token generation")
	fmt.Println("   - Token refresh mechanism")
	fmt.Println("   - Token revocation")
	fmt.Println("   - Expiry validation")
	fmt.Println("   - Invalid token detection")
	fmt.Println("   - Claims extraction helpers")
	fmt.Println()
	fmt.Println("🔒 Security Features:")
	fmt.Println("   - HMAC-SHA256 signature")
	fmt.Println("   - Issuer and audience validation")
	fmt.Println("   - Expiration enforcement")
	fmt.Println("   - Revocation list support")
	fmt.Println()
	fmt.Println("⚙️  Configuration:")
	fmt.Printf("   - Access Token Duration: %v\n", config.AccessTokenDuration)
	fmt.Printf("   - Refresh Token Duration: %v\n", config.RefreshTokenDuration)
	fmt.Printf("   - Signing Algorithm: %s\n", config.SigningMethod.Alg())
	fmt.Printf("   - Issuer: %s\n", config.Issuer)
	fmt.Printf("   - Revocation Enabled: %v\n", config.EnableRevocation)
}
