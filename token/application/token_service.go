package application

import (
	"fmt"

	"github.com/primadi/lokstra-auth/token"
	"github.com/primadi/lokstra/core/request"
)

// TokenService handles token validation, refresh, and revocation via HTTP.
// @RouterService name="token-service", prefix="${api-auth-prefix:/api/auth}/token", middlewares=["recovery", "request_logger"]
type TokenService struct {
	// @Inject "token-manager"
	TokenManager token.TokenManager
}

// ValidateRequest represents a token validation request
type ValidateRequest struct {
	Token string `json:"token" validate:"required"`
}

// ValidateResponse represents a token validation response
type ValidateResponse struct {
	Valid  bool           `json:"valid"`
	Claims map[string]any `json:"claims,omitempty"`
	Error  string         `json:"error,omitempty"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse represents a token refresh response
type RefreshResponse struct {
	Success      bool   `json:"success"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	Error        string `json:"error,omitempty"`
}

// RevokeRequest represents a token revocation request
type RevokeRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RevokeResponse represents a token revocation response
type RevokeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// IntrospectRequest represents a token introspection request
type IntrospectRequest struct {
	Token string `json:"token" validate:"required"`
}

// IntrospectResponse represents a token introspection response
type IntrospectResponse struct {
	Active    bool           `json:"active"`
	TokenType string         `json:"token_type,omitempty"`
	Subject   string         `json:"sub,omitempty"`
	TenantID  string         `json:"tenant_id,omitempty"`
	AppID     string         `json:"app_id,omitempty"`
	ExpiresAt int64          `json:"exp,omitempty"`
	IssuedAt  int64          `json:"iat,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
}

// Validate validates a JWT access token
// @Route "POST /validate"
func (s *TokenService) Validate(ctx *request.Context, req *ValidateRequest) (*ValidateResponse, error) {
	if req.Token == "" {
		return &ValidateResponse{
			Valid: false,
			Error: "token is required",
		}, nil
	}

	// Verify token
	result, err := s.TokenManager.Verify(ctx, req.Token)
	if err != nil {
		return &ValidateResponse{
			Valid: false,
			Error: fmt.Sprintf("verification failed: %v", err),
		}, nil
	}

	if !result.Valid {
		return &ValidateResponse{
			Valid: false,
			Error: fmt.Sprintf("invalid token: %v", result.Error),
		}, nil
	}

	return &ValidateResponse{
		Valid:  true,
		Claims: result.Claims,
	}, nil
}

// Refresh generates a new access token using a refresh token
// Implements refresh token rotation for enhanced security
// @Route "POST /refresh"
func (s *TokenService) Refresh(ctx *request.Context, req *RefreshRequest) (*RefreshResponse, error) {
	if req.RefreshToken == "" {
		return &RefreshResponse{
			Success: false,
			Error:   "refresh_token is required",
		}, nil
	}

	// Verify refresh token to extract claims
	result, err := s.TokenManager.Verify(ctx, req.RefreshToken)
	if err != nil {
		return &RefreshResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to verify refresh token: %v", err),
		}, nil
	}

	if !result.Valid {
		return &RefreshResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid refresh token: %v", result.Error),
		}, nil
	}

	// Check if this is actually a refresh token
	tokenType, _ := result.Claims.GetString("type")
	if tokenType != "refresh" {
		return &RefreshResponse{
			Success: false,
			Error:   "provided token is not a refresh token",
		}, nil
	}

	// Generate new access token from claims
	newAccessToken, err := s.TokenManager.Generate(ctx, result.Claims)
	if err != nil {
		return &RefreshResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to generate access token: %v", err),
		}, nil
	}

	// Generate NEW refresh token (rotation) from same claims
	newRefreshToken, err := s.TokenManager.GenerateRefreshToken(ctx, result.Claims)
	if err != nil {
		return &RefreshResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to generate new refresh token: %v", err),
		}, nil
	}

	// Revoke old refresh token (security: prevent reuse)
	if err := s.TokenManager.Revoke(ctx, req.RefreshToken); err != nil {
		// Log error but don't fail the request
		// Old token might already be expired/revoked
	}

	return &RefreshResponse{
		Success:      true,
		AccessToken:  newAccessToken.Value,
		RefreshToken: newRefreshToken.Value,
		TokenType:    newAccessToken.Type,
		ExpiresIn:    int64(newAccessToken.ExpiresAt.Sub(newAccessToken.IssuedAt).Seconds()),
	}, nil
}

// Revoke revokes a refresh token (logout)
// @Route "POST /revoke"
func (s *TokenService) Revoke(ctx *request.Context, req *RevokeRequest) (*RevokeResponse, error) {
	if req.RefreshToken == "" {
		return &RevokeResponse{
			Success: false,
			Error:   "refresh_token is required",
		}, nil
	}

	// Revoke refresh token
	err := s.TokenManager.Revoke(ctx, req.RefreshToken)
	if err != nil {
		return &RevokeResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to revoke token: %v", err),
		}, nil
	}

	return &RevokeResponse{
		Success: true,
		Message: "token revoked successfully",
	}, nil
}

// Introspect provides detailed information about a token
// @Route "POST /introspect"
func (s *TokenService) Introspect(ctx *request.Context, req *IntrospectRequest) (*IntrospectResponse, error) {
	if req.Token == "" {
		return &IntrospectResponse{
			Active: false,
		}, nil
	}

	// Verify token
	result, err := s.TokenManager.Verify(ctx, req.Token)
	if err != nil {
		return &IntrospectResponse{
			Active: false,
		}, nil
	}

	if !result.Valid {
		return &IntrospectResponse{
			Active: false,
		}, nil
	}

	// Extract standard claims
	subject, _ := result.Claims.GetString("sub")
	tenantID, _ := result.Claims.GetString("tenant_id")
	appID, _ := result.Claims.GetString("app_id")
	tokenType, _ := result.Claims.GetString("type")

	// Get exp and iat
	exp, _ := result.Claims.GetInt64("exp")
	iat, _ := result.Claims.GetInt64("iat")

	// If no explicit type, it's an access token
	if tokenType == "" {
		tokenType = "access"
	}

	return &IntrospectResponse{
		Active:    true,
		TokenType: tokenType,
		Subject:   subject,
		TenantID:  tenantID,
		AppID:     appID,
		ExpiresAt: exp,
		IssuedAt:  iat,
		Claims:    result.Claims,
	}, nil
}
