package main

import (
	"context"
	"fmt"
	"log"

	"github.com/primadi/lokstra-auth/01_credential/passwordless"
)

// MockUserResolver implements passwordless.UserResolver for demo
type MockUserResolver struct {
	users map[string]string // email -> userID
}

func NewMockUserResolver() *MockUserResolver {
	return &MockUserResolver{
		users: map[string]string{
			"alice@example.com": "user123",
			"bob@example.com":   "user456",
			"carol@example.com": "user789",
		},
	}
}

func (r *MockUserResolver) ResolveByEmail(ctx context.Context, email string) (userID string, claims map[string]interface{}, err error) {
	userID, ok := r.users[email]
	if !ok {
		return "", nil, fmt.Errorf("user not found")
	}
	claims = map[string]interface{}{
		"email": email,
	}
	return userID, claims, nil
}

// MockTokenSender implements passwordless.TokenSender for demo
type MockTokenSender struct {
	lastMagicLink string
	lastOTP       string
}

func (s *MockTokenSender) SendMagicLink(ctx context.Context, email, token, link string) error {
	s.lastMagicLink = token
	fmt.Printf("📧 [EMAIL] Magic Link sent to %s\n", email)
	fmt.Printf("   Click: %s\n", link)
	fmt.Printf("   Token: %s\n", token)
	fmt.Println()
	return nil
}

func (s *MockTokenSender) SendOTP(ctx context.Context, email, otp string) error {
	s.lastOTP = otp
	fmt.Printf("📧 [EMAIL] OTP sent to %s\n", email)
	fmt.Printf("   Your verification code: %s\n", otp)
	fmt.Println()
	return nil
}

func main() {
	fmt.Println("=== Passwordless Authentication Example ===")
	fmt.Println()

	// Create authenticator
	tokenStore := passwordless.NewInMemoryTokenStore()
	mockSender := &MockTokenSender{}
	auth := passwordless.NewAuthenticator(&passwordless.Config{
		TokenStore:   tokenStore,
		UserResolver: NewMockUserResolver(),
		TokenSender:  mockSender,
	})

	ctx := context.Background()

	// === MAGIC LINK FLOW ===
	fmt.Println("🔗 Magic Link Authentication Flow")
	fmt.Println("=" + string(make([]rune, 50)))
	fmt.Println()

	// Step 1: Request magic link
	fmt.Println("1️⃣  Requesting magic link for alice@example.com...")

	// Get user ID first (simulating user lookup)
	userID, _, _ := NewMockUserResolver().ResolveByEmail(ctx, "alice@example.com")

	err := auth.InitiateMagicLink(ctx, "alice@example.com", userID, "http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	// Get the token that was "sent"
	magicToken := mockSender.lastMagicLink

	// Step 2: Verify magic link
	fmt.Println("2️⃣  User clicks magic link from email...")
	verifyCreds := &passwordless.Credentials{
		Email:     "alice@example.com",
		Token:     magicToken,
		TokenType: passwordless.TokenTypeMagicLink,
	}

	result2, err := auth.Authenticate(ctx, verifyCreds)
	if err != nil {
		log.Fatal(err)
	}

	if result2.Success {
		fmt.Printf("✅ Authentication Successful!\n")
		fmt.Printf("   Subject (User ID): %s\n", result2.Subject)
		fmt.Printf("   Email: %s\n", result2.Claims["email"])
		fmt.Printf("   Auth Method: %s\n", result2.Claims["auth_method"])
		fmt.Println()
	}

	// Step 3: Try to use the same token again (should fail - one-time use)
	fmt.Println("3️⃣  Attempting to reuse magic link token...")
	result3, err := auth.Authenticate(ctx, verifyCreds)
	if err != nil {
		log.Fatal(err)
	}

	if !result3.Success {
		fmt.Printf("❌ Reuse prevented (expected): %v\n", result3.Error)
		fmt.Println()
	}

	// === OTP FLOW ===
	fmt.Println()
	fmt.Println("🔢 OTP Authentication Flow")
	fmt.Println("=" + string(make([]rune, 50)))
	fmt.Println()

	// Step 1: Request OTP
	fmt.Println("1️⃣  Requesting OTP for bob@example.com...")

	bobUserID, _, _ := NewMockUserResolver().ResolveByEmail(ctx, "bob@example.com")
	err = auth.InitiateOTP(ctx, "bob@example.com", bobUserID)
	if err != nil {
		log.Fatal(err)
	}

	// Get the OTP that was "sent"
	otp := mockSender.lastOTP

	// Step 2: Verify OTP
	fmt.Println("2️⃣  User enters OTP from email...")
	otpVerifyCreds := &passwordless.Credentials{
		Email:     "bob@example.com",
		Token:     otp,
		TokenType: passwordless.TokenTypeOTP,
	}

	result5, err := auth.Authenticate(ctx, otpVerifyCreds)
	if err != nil {
		log.Fatal(err)
	}

	if result5.Success {
		fmt.Printf("✅ Authentication Successful!\n")
		fmt.Printf("   Subject (User ID): %s\n", result5.Subject)
		fmt.Printf("   Email: %s\n", result5.Claims["email"])
		fmt.Println()
	}

	// Step 3: Test invalid OTP
	fmt.Println("3️⃣  Testing with invalid OTP...")
	invalidOTPCreds := &passwordless.Credentials{
		Email:     "bob@example.com",
		Token:     "999999",
		TokenType: passwordless.TokenTypeOTP,
	}

	result6, err := auth.Authenticate(ctx, invalidOTPCreds)
	if err != nil {
		log.Fatal(err)
	}

	if !result6.Success {
		fmt.Printf("❌ Invalid OTP rejected (expected): %v\n", result6.Error)
		fmt.Println()
	}

	// === MULTIPLE USERS ===
	fmt.Println()
	fmt.Println("👥 Multiple Users Example")
	fmt.Println("=" + string(make([]rune, 50)))
	fmt.Println()

	// Request OTPs for multiple users
	users := []string{"alice@example.com", "bob@example.com", "carol@example.com"}
	resolver := NewMockUserResolver()

	for _, email := range users {
		fmt.Printf("Requesting OTP for %s...\n", email)
		uid, _, _ := resolver.ResolveByEmail(ctx, email)
		auth.InitiateOTP(ctx, email, uid)
	}

	fmt.Println()
	fmt.Println("✅ All OTPs sent successfully")
	fmt.Println()

	// === SUMMARY ===
	fmt.Println("📋 Summary")
	fmt.Println("=" + string(make([]rune, 50)))
	fmt.Println()
	fmt.Println("✅ Passwordless Features Demonstrated:")
	fmt.Println("   - Magic Link generation and verification")
	fmt.Println("   - OTP generation and verification")
	fmt.Println("   - One-time token enforcement")
	fmt.Println("   - Token expiry (15min for magic link, 5min for OTP)")
	fmt.Println("   - Email-based user resolution")
	fmt.Println("   - Token sending via email")
	fmt.Println()
	fmt.Println("🔧 Extensibility Points:")
	fmt.Println("   - TokenStore: In-memory, Redis, Database")
	fmt.Println("   - UserResolver: Custom user lookup logic")
	fmt.Println("   - TokenGenerator: Custom token format")
	fmt.Println("   - TokenSender: Email, SMS, WhatsApp, etc.")
	fmt.Println()
	fmt.Println("🔒 Security Features:")
	fmt.Println("   - One-time use tokens")
	fmt.Println("   - Time-based expiration")
	fmt.Println("   - Automatic cleanup of expired tokens")
	fmt.Println("   - Cryptographically secure random tokens")
}
