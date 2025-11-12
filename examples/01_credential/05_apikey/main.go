package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra-auth/01_credential/apikey"
)

func main() {
	fmt.Println("=== API Key Authentication Example ===")
	fmt.Println()

	// Create authenticator with in-memory key store
	keyStore := apikey.NewInMemoryKeyStore()
	auth := apikey.NewAuthenticator(&apikey.Config{
		KeyStore: keyStore,
	})

	ctx := context.Background()

	// Example 1: Generate API key with expiration
	fmt.Println("1. Generating API key for user 'alice'...")
	expiresIn := 30 * 24 * time.Hour // 30 days
	keyString1, apiKey1, err := auth.GenerateKey(
		ctx,
		"alice123",
		"Production API Key",
		[]string{"read:users", "write:posts", "delete:posts"},
		&expiresIn,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ API Key Generated:\n")
	fmt.Printf("   Key: %s\n", keyString1)
	fmt.Printf("   Key ID: %s\n", apiKey1.ID)
	fmt.Printf("   Prefix: %s\n", apiKey1.Prefix)
	fmt.Printf("   Owner: %s\n", apiKey1.UserID)
	fmt.Printf("   Name: %s\n", apiKey1.Name)
	fmt.Printf("   Scopes: %v\n", apiKey1.Scopes)
	fmt.Printf("   Expires: %s\n\n", apiKey1.ExpiresAt.Format(time.RFC3339))

	// Example 2: Generate API key without expiration
	fmt.Println("2. Generating permanent API key for user 'bob'...")
	keyString2, apiKey2, err := auth.GenerateKey(
		ctx,
		"bob456",
		"Development API Key",
		[]string{"read:users"},
		nil, // No expiration
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Permanent API Key Generated:\n")
	fmt.Printf("   Key: %s\n", keyString2)
	fmt.Printf("   Key ID: %s\n", apiKey2.ID)
	fmt.Printf("   Scopes: %v\n", apiKey2.Scopes)
	fmt.Printf("   Expires: Never\n\n")

	// Add metadata to API key
	apiKey2.Metadata = map[string]interface{}{
		"app_name":    "Mobile App",
		"environment": "development",
		"version":     "1.0.0",
	}
	keyStore.Store(ctx, apiKey2)

	// Example 3: Authenticate with API key
	fmt.Println("3. Authenticating with Alice's API key...")
	creds1 := &apikey.Credentials{
		APIKey: keyString1,
	}

	result1, err := auth.Authenticate(ctx, creds1)
	if err != nil {
		log.Fatal(err)
	}

	if result1.Success {
		fmt.Printf("✅ Authentication Successful!\n")
		fmt.Printf("   Subject (User ID): %s\n", result1.Subject)
		fmt.Printf("   Key ID: %s\n", result1.Claims["key_id"])
		fmt.Printf("   Key Name: %s\n", result1.Claims["key_name"])
		fmt.Printf("   Scopes: %v\n", result1.Claims["scopes"])
		fmt.Printf("   Auth Type: %s\n\n", result1.Claims["auth_type"])
	} else {
		fmt.Printf("❌ Authentication Failed: %v\n\n", result1.Error)
	}

	// Example 4: Authenticate with Bob's API key (with metadata)
	fmt.Println("4. Authenticating with Bob's API key...")
	creds2 := &apikey.Credentials{
		APIKey: keyString2,
	}

	result2, err := auth.Authenticate(ctx, creds2)
	if err != nil {
		log.Fatal(err)
	}

	if result2.Success {
		fmt.Printf("✅ Authentication Successful!\n")
		fmt.Printf("   Subject (User ID): %s\n", result2.Subject)
		fmt.Printf("   Metadata:\n")
		fmt.Printf("     - App: %s\n", result2.Claims["app_name"])
		fmt.Printf("     - Environment: %s\n", result2.Claims["environment"])
		fmt.Printf("     - Version: %s\n\n", result2.Claims["version"])
	}

	// Example 5: Invalid API key
	fmt.Println("5. Testing with invalid API key...")
	invalidCreds := &apikey.Credentials{
		APIKey: "invalid-key-12345",
	}

	result3, err := auth.Authenticate(ctx, invalidCreds)
	if err != nil {
		log.Fatal(err)
	}

	if !result3.Success {
		fmt.Printf("❌ Authentication Failed (expected): %v\n\n", result3.Error)
	}

	// Example 6: Revoke API key
	fmt.Println("6. Revoking Alice's API key...")
	err = auth.RevokeKey(ctx, apiKey1.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ API Key revoked: %s\n\n", apiKey1.ID)

	// Example 7: Try to use revoked key
	fmt.Println("7. Testing with revoked API key...")
	result4, err := auth.Authenticate(ctx, creds1)
	if err != nil {
		log.Fatal(err)
	}

	if !result4.Success {
		fmt.Printf("❌ Authentication Failed (expected): %v\n\n", result4.Error)
	}

	// Example 8: Generate key with short expiration for testing
	fmt.Println("8. Testing expired API key...")
	shortExpiry := 1 * time.Second
	keyString3, _, err := auth.GenerateKey(
		ctx,
		"charlie789",
		"Short-lived Key",
		[]string{"read:users"},
		&shortExpiry,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Short-lived key generated (expires in 1 second)\n")
	fmt.Printf("   Waiting for expiration...\n")
	time.Sleep(2 * time.Second)

	creds3 := &apikey.Credentials{
		APIKey: keyString3,
	}

	result5, err := auth.Authenticate(ctx, creds3)
	if err != nil {
		log.Fatal(err)
	}

	if !result5.Success {
		fmt.Printf("❌ Authentication Failed (expected): %v\n\n", result5.Error)
	}

	// Example 9: List keys by prefix
	fmt.Println("9. Listing keys by prefix...")
	keys, err := keyStore.GetByPrefix(ctx, apiKey2.Prefix)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Found %d key(s) with prefix '%s':\n", len(keys), apiKey2.Prefix)
	for _, key := range keys {
		fmt.Printf("   - ID: %s, Name: %s, User: %s\n", key.ID, key.Name, key.UserID)
	}
	fmt.Println()

	// Example 10: Check last used timestamp
	fmt.Println("10. Checking last used timestamp...")
	hasher := apikey.NewKeyHasher()
	apiKey2Updated, _ := keyStore.GetByHash(ctx, hasher.Hash(keyString2))
	if apiKey2Updated.LastUsed != nil {
		fmt.Printf("✅ Bob's key last used: %s\n", apiKey2Updated.LastUsed.Format(time.RFC3339))
		fmt.Printf("   Time since last use: %s\n\n", time.Since(*apiKey2Updated.LastUsed))
	}

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Println("✅ API Key Features Demonstrated:")
	fmt.Println("   - Key generation with and without expiration")
	fmt.Println("   - Scope-based permissions")
	fmt.Println("   - Metadata support")
	fmt.Println("   - Key revocation")
	fmt.Println("   - Expiry validation")
	fmt.Println("   - Last used tracking")
	fmt.Println("   - Prefix-based key listing")
	fmt.Println("\n✅ Security Features:")
	fmt.Println("   - SHA3-256 key hashing")
	fmt.Println("   - Constant-time comparison")
	fmt.Println("   - One-time key display")
	fmt.Println("   - Automatic expiry checking")
	fmt.Println("   - Soft delete with revocation timestamp")
}
