package lokstraauth

import (
	"context"
	"errors"
	"fmt"

	credential "github.com/primadi/lokstra-auth/01_credential/domain"
	token "github.com/primadi/lokstra-auth/02_token"
	subject "github.com/primadi/lokstra-auth/03_subject"
	authz "github.com/primadi/lokstra-auth/04_authz"
)

var (
	ErrNoAuthenticator         = errors.New("no authenticator configured")
	ErrNoTokenManager          = errors.New("no token manager configured")
	ErrNoSubjectResolver       = errors.New("no subject resolver configured")
	ErrNoContextBuilder        = errors.New("no identity context builder configured")
	ErrNoAuthorizer            = errors.New("no authorizer configured")
	ErrAuthenticationFailed    = errors.New("authentication failed")
	ErrTokenGenerationFailed   = errors.New("token generation failed")
	ErrSubjectResolutionFailed = errors.New("subject resolution failed")
	ErrAuthorizationFailed     = errors.New("authorization check failed")
)

// Auth is the main runtime object for Lokstra Auth framework
// It orchestrates all 4 layers and provides a unified API
type Auth struct {
	// Layer 1: Credential Input
	authenticators map[string]credential.Authenticator

	// Layer 2: Token Management
	tokenManager token.TokenManager

	// Layer 3: Subject Resolution
	subjectResolver subject.SubjectResolver
	contextBuilder  subject.IdentityContextBuilder

	// Layer 4: Authorization
	authorizer authz.Authorizer

	// Configuration
	config *Config
}

// Config holds the configuration for Auth runtime
type Config struct {
	// DefaultAuthenticatorType is the default authenticator to use
	DefaultAuthenticatorType string

	// IssueRefreshToken indicates whether to issue refresh tokens
	IssueRefreshToken bool

	// SessionManagement indicates whether to manage sessions
	SessionManagement bool

	// Metadata contains additional runtime metadata
	Metadata map[string]any
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultAuthenticatorType: "basic",
		IssueRefreshToken:        true,
		SessionManagement:        false,
		Metadata:                 make(map[string]any),
	}
}

// New creates a new Auth runtime instance
func New(config *Config) *Auth {
	if config == nil {
		config = DefaultConfig()
	}

	return &Auth{
		authenticators: make(map[string]credential.Authenticator),
		config:         config,
	}
}

// RegisterAuthenticator registers an authenticator for a specific type
func (a *Auth) RegisterAuthenticator(authType string, authenticator credential.Authenticator) {
	a.authenticators[authType] = authenticator
}

// SetTokenManager sets the token manager
func (a *Auth) SetTokenManager(manager token.TokenManager) {
	a.tokenManager = manager
}

// SetSubjectResolver sets the subject resolver
func (a *Auth) SetSubjectResolver(resolver subject.SubjectResolver) {
	a.subjectResolver = resolver
}

// SetIdentityContextBuilder sets the identity context builder
func (a *Auth) SetIdentityContextBuilder(builder subject.IdentityContextBuilder) {
	a.contextBuilder = builder
}

// SetAuthorizer sets the authorizer
func (a *Auth) SetAuthorizer(authorizer authz.Authorizer) {
	a.authorizer = authorizer
}

// GetAuthorizer returns the configured authorizer
func (a *Auth) GetAuthorizer() authz.Authorizer {
	return a.authorizer
}

// LoginRequest represents a login request
type LoginRequest struct {
	// AuthContext contains tenant and app context (REQUIRED for multi-tenant)
	AuthContext *credential.AuthContext

	// Credentials contains the credentials to authenticate
	Credentials credential.Credentials

	// Metadata contains additional request metadata
	Metadata map[string]any
}

// LoginResponse represents a login response
type LoginResponse struct {
	// AccessToken is the issued access token
	AccessToken *token.Token

	// RefreshToken is the issued refresh token (if enabled)
	RefreshToken *token.Token

	// Identity is the resolved identity context
	Identity *subject.IdentityContext

	// Metadata contains additional response metadata
	Metadata map[string]any
}

// Login performs the complete authentication flow
// Layer 1 -> Layer 2 -> Layer 3
func (a *Auth) Login(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	// Validate auth context
	if request.AuthContext == nil {
		return nil, errors.New("auth context is required")
	}
	if err := request.AuthContext.Validate(); err != nil {
		return nil, fmt.Errorf("invalid auth context: %w", err)
	}

	// Layer 1: Authenticate credentials
	credType := request.Credentials.Type()
	authenticator, ok := a.authenticators[credType]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoAuthenticator, credType)
	}

	authResult, err := authenticator.Authenticate(ctx, request.AuthContext, request.Credentials)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	if !authResult.Success {
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, authResult.Error)
	}

	// Layer 2: Generate tokens
	if a.tokenManager == nil {
		return nil, ErrNoTokenManager
	}

	accessToken, err := a.tokenManager.Generate(ctx, authResult.Claims)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenGenerationFailed, err)
	}

	response := &LoginResponse{
		AccessToken: accessToken,
		Metadata:    make(map[string]any),
	}

	// Generate refresh token if enabled
	if a.config.IssueRefreshToken {
		// Check if token manager supports refresh tokens
		if rtHandler, ok := a.tokenManager.(interface {
			GenerateRefreshToken(ctx context.Context, claims token.Claims) (*token.Token, error)
		}); ok {
			refreshToken, err := rtHandler.GenerateRefreshToken(ctx, authResult.Claims)
			if err == nil {
				response.RefreshToken = refreshToken
			}
		}
	}

	// Layer 3: Resolve subject and build identity context (optional)
	if a.subjectResolver != nil && a.contextBuilder != nil {
		sub, err := a.subjectResolver.Resolve(ctx, authResult.Claims)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSubjectResolutionFailed, err)
		}

		identity, err := a.contextBuilder.Build(ctx, sub)
		if err != nil {
			return nil, fmt.Errorf("identity context building error: %w", err)
		}

		response.Identity = identity
	}

	return response, nil
}

// VerifyRequest represents a token verification request
type VerifyRequest struct {
	// Token is the token to verify
	Token string

	// BuildIdentityContext indicates whether to build full identity context
	BuildIdentityContext bool

	// Metadata contains additional request metadata
	Metadata map[string]any
}

// VerifyResponse represents a token verification response
type VerifyResponse struct {
	// Valid indicates whether the token is valid
	Valid bool

	// Claims contains the extracted claims
	Claims token.Claims

	// Identity is the resolved identity context (if requested)
	Identity *subject.IdentityContext

	// Metadata contains additional response metadata
	Metadata map[string]any
}

// Verify verifies a token and optionally builds identity context
// Layer 2 -> Layer 3
func (a *Auth) Verify(ctx context.Context, request *VerifyRequest) (*VerifyResponse, error) {
	// Layer 2: Verify token
	if a.tokenManager == nil {
		return nil, ErrNoTokenManager
	}

	verifyResult, err := a.tokenManager.Verify(ctx, request.Token)
	if err != nil {
		return nil, fmt.Errorf("token verification error: %w", err)
	}

	response := &VerifyResponse{
		Valid:    verifyResult.Valid,
		Claims:   verifyResult.Claims,
		Metadata: make(map[string]any),
	}

	if !verifyResult.Valid {
		return response, nil
	}

	// Layer 3: Build identity context if requested
	if request.BuildIdentityContext && a.subjectResolver != nil && a.contextBuilder != nil {
		sub, err := a.subjectResolver.Resolve(ctx, verifyResult.Claims)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSubjectResolutionFailed, err)
		}

		identity, err := a.contextBuilder.Build(ctx, sub)
		if err != nil {
			return nil, fmt.Errorf("identity context building error: %w", err)
		}

		response.Identity = identity
	}

	return response, nil
}

// Authorize checks if a subject is authorized to perform an action on a resource
// Layer 4
func (a *Auth) Authorize(ctx context.Context, request *authz.AuthorizationRequest) (*authz.AuthorizationDecision, error) {
	if a.authorizer == nil {
		return nil, ErrNoAuthorizer
	}

	decision, err := a.authorizer.Evaluate(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthorizationFailed, err)
	}

	return decision, nil
}

// CheckPermission is a convenience method to check a simple permission
func (a *Auth) CheckPermission(ctx context.Context, identity *subject.IdentityContext, permission string) (bool, error) {
	if a.authorizer == nil {
		return false, ErrNoAuthorizer
	}

	checker, ok := a.authorizer.(authz.PermissionChecker)
	if !ok {
		return false, errors.New("authorizer does not support permission checking")
	}

	return checker.HasPermission(ctx, identity, permission)
}

// CheckRole is a convenience method to check if identity has a role
func (a *Auth) CheckRole(ctx context.Context, identity *subject.IdentityContext, role string) (bool, error) {
	if a.authorizer == nil {
		return false, ErrNoAuthorizer
	}

	checker, ok := a.authorizer.(authz.RoleChecker)
	if !ok {
		return false, errors.New("authorizer does not support role checking")
	}

	return checker.HasRole(ctx, identity, role)
}
