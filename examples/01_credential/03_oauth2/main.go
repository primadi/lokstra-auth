package main

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra-auth/01_credential/oauth2"
)

func main() {
	fmt.Println("=== OAuth2 Authentication Example ===")
	fmt.Println()

	// Create OAuth2 authenticator with default providers
	auth := oauth2.NewAuthenticator(nil)

	ctx := context.Background()

	// NOTE: This example demonstrates the authentication flow
	// In production, you would:
	// 1. Redirect user to provider's authorization URL
	// 2. User grants permission
	// 3. Provider returns access token
	// 4. Use the access token to authenticate

	fmt.Println("📋 OAuth2 Flow Steps:")
	fmt.Println()
	fmt.Println("1️⃣  AUTHORIZATION REQUEST")
	fmt.Println("   Redirect user to provider's authorization URL:")
	fmt.Println()
	fmt.Println("   Google:")
	fmt.Println("   https://accounts.google.com/o/oauth2/v2/auth?")
	fmt.Println("     client_id=YOUR_CLIENT_ID")
	fmt.Println("     &redirect_uri=http://localhost:8080/auth/google/callback")
	fmt.Println("     &response_type=code")
	fmt.Println("     &scope=openid%20email%20profile")
	fmt.Println()
	fmt.Println("   GitHub:")
	fmt.Println("   https://github.com/login/oauth/authorize?")
	fmt.Println("     client_id=YOUR_CLIENT_ID")
	fmt.Println("     &redirect_uri=http://localhost:8080/auth/github/callback")
	fmt.Println("     &scope=user:email")
	fmt.Println()
	fmt.Println("   Facebook:")
	fmt.Println("   https://www.facebook.com/v12.0/dialog/oauth?")
	fmt.Println("     client_id=YOUR_APP_ID")
	fmt.Println("     &redirect_uri=http://localhost:8080/auth/facebook/callback")
	fmt.Println("     &scope=email,public_profile")
	fmt.Println()

	fmt.Println("2️⃣  USER GRANTS PERMISSION")
	fmt.Println("   User logs in and authorizes your application")
	fmt.Println()

	fmt.Println("3️⃣  AUTHORIZATION CODE CALLBACK")
	fmt.Println("   Provider redirects back with authorization code:")
	fmt.Println("   http://localhost:8080/auth/google/callback?code=AUTHORIZATION_CODE")
	fmt.Println()

	fmt.Println("4️⃣  TOKEN EXCHANGE & AUTHENTICATION")
	fmt.Println("   Use the access token to authenticate:")
	fmt.Println()

	// Example: Authenticate with Google OAuth2
	// In real app, you would get this token from OAuth2 flow
	exampleToken := "ya29.a0AfH6SMBxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	fmt.Println("   Example Google Authentication:")
	creds := &oauth2.Credentials{
		Provider:    oauth2.ProviderGoogle,
		AccessToken: exampleToken,
	}

	// NOTE: This will fail because we're using an example token
	// In production, use the actual access token from OAuth2 flow
	result, err := auth.Authenticate(ctx, creds)
	if err != nil {
		fmt.Printf("   ⚠️  Example authentication (expected to fail with example token)\n")
		fmt.Printf("   Error: %v\n", err)
		fmt.Println()
	}

	if result != nil && result.Success {
		fmt.Println("   ✅ Authentication Successful!")
		fmt.Printf("   Subject (User ID): %s\n", result.Subject)
		fmt.Printf("   Email: %s\n", result.Claims["email"])
		fmt.Printf("   Name: %s\n", result.Claims["name"])
		fmt.Printf("   Picture: %s\n", result.Claims["picture"])
		fmt.Printf("   Email Verified: %v\n", result.Claims["email_verified"])
		fmt.Println()
	}

	// Example code structure for a real implementation
	fmt.Println("📝 Production Implementation Example:")
	fmt.Println()
	fmt.Println("```go")
	fmt.Println("// After user completes OAuth2 flow, you receive access_token")
	fmt.Println("func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {")
	fmt.Println("    // Get access token from OAuth2 provider")
	fmt.Println("    accessToken := getAccessTokenFromProvider() // Your OAuth2 implementation")
	fmt.Println("    ")
	fmt.Println("    // Authenticate with access token")
	fmt.Println("    creds := &oauth2.Credentials{")
	fmt.Println("        Provider:    oauth2.ProviderGoogle,")
	fmt.Println("        AccessToken: accessToken,")
	fmt.Println("    }")
	fmt.Println("    ")
	fmt.Println("    result, err := auth.Authenticate(r.Context(), creds)")
	fmt.Println("    if err != nil || !result.Success {")
	fmt.Println("        http.Error(w, \"Authentication failed\", http.StatusUnauthorized)")
	fmt.Println("        return")
	fmt.Println("    }")
	fmt.Println("    ")
	fmt.Println("    // User authenticated successfully")
	fmt.Println("    userID := result.Subject")
	fmt.Println("    email := result.Claims[\"email\"].(string)")
	fmt.Println("    ")
	fmt.Println("    // Create session, generate JWT, etc.")
	fmt.Println("    // ...")
	fmt.Println("}")
	fmt.Println("```")
	fmt.Println()

	// Show provider-specific claims
	fmt.Println("📦 Provider-Specific Claims:")
	fmt.Println()
	fmt.Println("Google Claims:")
	fmt.Println("  - email: string")
	fmt.Println("  - name: string")
	fmt.Println("  - picture: string (URL)")
	fmt.Println("  - email_verified: bool")
	fmt.Println()
	fmt.Println("GitHub Claims:")
	fmt.Println("  - email: string (primary verified email)")
	fmt.Println("  - login: string (username)")
	fmt.Println("  - name: string")
	fmt.Println("  - avatar_url: string (URL)")
	fmt.Println()
	fmt.Println("Facebook Claims:")
	fmt.Println("  - email: string")
	fmt.Println("  - name: string")
	fmt.Println("  - picture: string (URL)")
	fmt.Println()

	// Security notes
	fmt.Println("🔒 Security Best Practices:")
	fmt.Println()
	fmt.Println("1. Always use HTTPS for redirect URLs in production")
	fmt.Println("2. Validate the state parameter to prevent CSRF attacks")
	fmt.Println("3. Store client secrets securely (environment variables)")
	fmt.Println("4. Implement proper error handling")
	fmt.Println("5. Verify email_verified claim before trusting email")
	fmt.Println("6. Use appropriate OAuth2 scopes (minimal required)")
	fmt.Println()

	fmt.Println("✅ OAuth2 Authenticator Ready!")
	fmt.Println("   Supported Providers: Google, GitHub, Facebook")
	fmt.Println("   Type:", auth.Type())
}
