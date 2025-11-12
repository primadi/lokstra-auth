package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	credential "github.com/primadi/lokstra-auth/01_credential"
)

var (
	ErrInvalidToken    = errors.New("invalid OAuth2 token")
	ErrTokenExpired    = errors.New("OAuth2 token expired")
	ErrInvalidProvider = errors.New("invalid OAuth2 provider")
	ErrUserInfoFailed  = errors.New("failed to fetch user info")
)

// Provider represents an OAuth2 provider (Google, GitHub, Facebook, etc.)
type Provider string

const (
	ProviderGoogle    Provider = "google"
	ProviderGithub    Provider = "github"
	ProviderFacebook  Provider = "facebook"
	ProviderMicrosoft Provider = "microsoft"
)

// Credentials represents OAuth2 credentials
type Credentials struct {
	Provider    Provider
	AccessToken string
	IDToken     string // For OIDC providers like Google
}

func (c *Credentials) Type() string {
	return "oauth2"
}

func (c *Credentials) Validate() error {
	if c.Provider == "" {
		return errors.New("provider is required")
	}
	if c.AccessToken == "" && c.IDToken == "" {
		return errors.New("access_token or id_token is required")
	}
	return nil
}

// UserInfo represents user information from OAuth2 provider
type UserInfo struct {
	ID            string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	Provider      Provider
	RawData       map[string]interface{}
}

// ProviderConfig holds configuration for an OAuth2 provider
type ProviderConfig struct {
	Name         Provider
	UserInfoURL  string
	ValidateFunc func(ctx context.Context, token string) (*UserInfo, error)
}

// Authenticator handles OAuth2 authentication
type Authenticator struct {
	providers map[Provider]*ProviderConfig
	client    *http.Client
}

// Config holds configuration for OAuth2 authenticator
type Config struct {
	// HTTPClient is the HTTP client for making requests
	HTTPClient *http.Client

	// Timeout for HTTP requests
	Timeout time.Duration

	// Custom provider configurations
	CustomProviders map[Provider]*ProviderConfig
}

// DefaultConfig returns default OAuth2 configuration
func DefaultConfig() *Config {
	return &Config{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Timeout: 10 * time.Second,
	}
}

// NewAuthenticator creates a new OAuth2 authenticator
func NewAuthenticator(config *Config) *Authenticator {
	if config == nil {
		config = DefaultConfig()
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	auth := &Authenticator{
		providers: make(map[Provider]*ProviderConfig),
		client:    config.HTTPClient,
	}

	// Register built-in providers
	auth.registerBuiltinProviders()

	// Register custom providers
	if config.CustomProviders != nil {
		for provider, cfg := range config.CustomProviders {
			auth.providers[provider] = cfg
		}
	}

	return auth
}

// registerBuiltinProviders registers default OAuth2 providers
func (a *Authenticator) registerBuiltinProviders() {
	// Google OAuth2
	a.providers[ProviderGoogle] = &ProviderConfig{
		Name:        ProviderGoogle,
		UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
		ValidateFunc: func(ctx context.Context, token string) (*UserInfo, error) {
			return a.fetchGoogleUserInfo(ctx, token)
		},
	}

	// GitHub OAuth2
	a.providers[ProviderGithub] = &ProviderConfig{
		Name:        ProviderGithub,
		UserInfoURL: "https://api.github.com/user",
		ValidateFunc: func(ctx context.Context, token string) (*UserInfo, error) {
			return a.fetchGithubUserInfo(ctx, token)
		},
	}

	// Facebook OAuth2
	a.providers[ProviderFacebook] = &ProviderConfig{
		Name:        ProviderFacebook,
		UserInfoURL: "https://graph.facebook.com/me?fields=id,name,email,picture",
		ValidateFunc: func(ctx context.Context, token string) (*UserInfo, error) {
			return a.fetchFacebookUserInfo(ctx, token)
		},
	}
}

// Authenticate verifies OAuth2 credentials
func (a *Authenticator) Authenticate(ctx context.Context, creds credential.Credentials) (*credential.AuthenticationResult, error) {
	oauth2Creds, ok := creds.(*Credentials)
	if !ok {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   errors.New("invalid credentials type"),
		}, nil
	}

	if err := oauth2Creds.Validate(); err != nil {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	// Get provider config
	providerCfg, ok := a.providers[oauth2Creds.Provider]
	if !ok {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   fmt.Errorf("%w: %s", ErrInvalidProvider, oauth2Creds.Provider),
		}, nil
	}

	// Validate token and get user info
	userInfo, err := providerCfg.ValidateFunc(ctx, oauth2Creds.AccessToken)
	if err != nil {
		return &credential.AuthenticationResult{
			Success: false,
			Error:   err,
		}, nil
	}

	// Build claims from user info
	claims := map[string]interface{}{
		"sub":            userInfo.ID,
		"email":          userInfo.Email,
		"email_verified": userInfo.EmailVerified,
		"name":           userInfo.Name,
		"picture":        userInfo.Picture,
		"provider":       string(userInfo.Provider),
	}

	// Add raw data
	for key, value := range userInfo.RawData {
		if _, exists := claims[key]; !exists {
			claims[key] = value
		}
	}

	return &credential.AuthenticationResult{
		Success: true,
		Subject: userInfo.ID,
		Claims:  claims,
	}, nil
}

// Type returns the authenticator type
func (a *Authenticator) Type() string {
	return "oauth2"
}

// fetchGoogleUserInfo fetches user info from Google
func (a *Authenticator) fetchGoogleUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code %d", ErrInvalidToken, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:            getString(data, "id"),
		Email:         getString(data, "email"),
		EmailVerified: getBool(data, "verified_email"),
		Name:          getString(data, "name"),
		Picture:       getString(data, "picture"),
		Provider:      ProviderGoogle,
		RawData:       data,
	}, nil
}

// fetchGithubUserInfo fetches user info from GitHub
func (a *Authenticator) fetchGithubUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code %d", ErrInvalidToken, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	// GitHub doesn't provide email in user endpoint by default
	// Need to fetch from /user/emails endpoint
	email := getString(data, "email")
	if email == "" {
		email, _ = a.fetchGithubEmail(ctx, token)
	}

	return &UserInfo{
		ID:            fmt.Sprintf("%v", data["id"]),
		Email:         email,
		EmailVerified: true, // GitHub emails are verified
		Name:          getString(data, "name"),
		Picture:       getString(data, "avatar_url"),
		Provider:      ProviderGithub,
		RawData:       data,
	}, nil
}

// fetchGithubEmail fetches primary email from GitHub
func (a *Authenticator) fetchGithubEmail(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []map[string]interface{}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	// Find primary email
	for _, emailData := range emails {
		if getBool(emailData, "primary") {
			return getString(emailData, "email"), nil
		}
	}

	// Return first email if no primary found
	if len(emails) > 0 {
		return getString(emails[0], "email"), nil
	}

	return "", nil
}

// fetchFacebookUserInfo fetches user info from Facebook
func (a *Authenticator) fetchFacebookUserInfo(ctx context.Context, token string) (*UserInfo, error) {
	url := fmt.Sprintf("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=%s", token)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code %d", ErrInvalidToken, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	pictureURL := ""
	if pictureData, ok := data["picture"].(map[string]interface{}); ok {
		if dataMap, ok := pictureData["data"].(map[string]interface{}); ok {
			pictureURL = getString(dataMap, "url")
		}
	}

	return &UserInfo{
		ID:            getString(data, "id"),
		Email:         getString(data, "email"),
		EmailVerified: true, // Facebook emails are verified
		Name:          getString(data, "name"),
		Picture:       pictureURL,
		Provider:      ProviderFacebook,
		RawData:       data,
	}, nil
}

// Helper functions

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
