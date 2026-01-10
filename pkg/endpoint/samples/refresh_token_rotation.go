package tenant

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	coredomain "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/service/token"
	"github.com/primadi/lokstra-auth/pkg/store"
)

// RefreshTokenRotationService handles refresh token rotation with grace period
type RefreshTokenRotationService struct {
	refreshTokenStore store.RefreshTokenStore
	tokenManager      token.TokenManager
}

var (
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")
	ErrRefreshTokenReused  = errors.New("refresh token reuse detected - possible token theft")
)

// RefreshAccessToken refreshes access token with rotation and grace period support
func (s *RefreshTokenRotationService) RefreshAccessToken(
	ctx context.Context,
	refreshTokenValue string,
	tenantConfig *coredomain.TenantSecurityConfig,
) (*TokenPair, error) {

	// 1. Hash the incoming refresh token to lookup in DB
	tokenHash := hashToken(refreshTokenValue)

	// 2. Get refresh token from store
	storedToken, err := s.refreshTokenStore.GetByToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 3. Check expiry
	if storedToken.IsExpired() {
		return nil, ErrRefreshTokenExpired
	}

	// 4. Check revoked
	if storedToken.IsRevoked() {
		return nil, ErrRefreshTokenRevoked
	}

	// 5. Handle token status with grace period
	gracePeriod := 0
	if tenantConfig != nil {
		gracePeriod = tenantConfig.RefreshTokenRotationGracePeriod
	}

	switch storedToken.Status {
	case coredomain.RefreshTokenStatusActive:
		// Token is active - can be used
		// Mark as used if rotation is enabled
		if tenantConfig != nil && tenantConfig.EnableRefreshTokenRotation {
			if err := s.refreshTokenStore.MarkUsed(ctx, storedToken.ID); err != nil {
				return nil, fmt.Errorf("failed to mark token as used: %w", err)
			}
		}

	case coredomain.RefreshTokenStatusUsed:
		// Token already used - check grace period
		if gracePeriod <= 0 {
			// No grace period - this is a reuse attempt
			return s.handleTokenReuse(ctx, storedToken, tenantConfig)
		}

		// Check if within grace period
		if !storedToken.IsWithinGracePeriod(gracePeriod) {
			// Outside grace period - this is suspicious
			return s.handleTokenReuse(ctx, storedToken, tenantConfig)
		}

		// Within grace period - allow it (for multi-tab scenario)
		// Don't mark as used again, keep the original used_at timestamp

	default:
		return nil, fmt.Errorf("refresh token cannot be used (status: %s)", storedToken.Status)
	}

	// 6. Generate new access token
	claims := token.Claims{
		"sub":       storedToken.UserID,
		"tenant_id": storedToken.TenantID,
		"app_id":    storedToken.AppID,
	}

	if storedToken.SessionID != nil {
		claims["session_id"] = *storedToken.SessionID
	}

	accessToken, err := s.tokenManager.Generate(ctx, claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 7. Generate new refresh token if rotation enabled
	var newRefreshToken *token.Token
	if tenantConfig != nil && tenantConfig.EnableRefreshTokenRotation {
		newRefreshToken, err = s.generateRotatedRefreshToken(ctx, storedToken, claims)
		if err != nil {
			return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
		}
	} else {
		// No rotation - return same refresh token
		newRefreshToken = &token.Token{
			Value:     refreshTokenValue,
			Type:      "Bearer",
			TenantID:  storedToken.TenantID,
			AppID:     *storedToken.AppID,
			ExpiresAt: storedToken.ExpiresAt,
		}
	}

	return &TokenPair{
		AccessToken:  accessToken.Value,
		RefreshToken: newRefreshToken.Value,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(accessToken.ExpiresAt).Seconds()),
	}, nil
}

// handleTokenReuse handles suspected token reuse (possible theft)
func (s *RefreshTokenRotationService) handleTokenReuse(
	ctx context.Context,
	reusedToken *coredomain.RefreshToken,
	tenantConfig *coredomain.TenantSecurityConfig,
) (*TokenPair, error) {

	// If reuse detection is enabled, revoke entire token family
	if tenantConfig != nil && tenantConfig.DetectRefreshTokenReuse && reusedToken.TokenFamily != nil {
		// Revoke all tokens in this family (security breach!)
		if err := s.refreshTokenStore.RevokeFamily(ctx, *reusedToken.TokenFamily); err != nil {
			// Log error but still return error to user
			fmt.Printf("ERROR: Failed to revoke token family: %v\n", err)
		}
	}

	return nil, ErrRefreshTokenReused
}

// generateRotatedRefreshToken creates a new refresh token from the old one
func (s *RefreshTokenRotationService) generateRotatedRefreshToken(
	ctx context.Context,
	oldToken *coredomain.RefreshToken,
	claims token.Claims,
) (*token.Token, error) {

	// Generate new refresh token via TokenManager
	newToken, err := s.tokenManager.GenerateRefreshToken(ctx, claims)
	if err != nil {
		return nil, err
	}

	// Store in database with link to old token
	refreshTokenRecord := &coredomain.RefreshToken{
		ID:            generateTokenID(),
		TenantID:      oldToken.TenantID,
		UserID:        oldToken.UserID,
		AppID:         oldToken.AppID,
		SessionID:     oldToken.SessionID,
		TokenHash:     hashToken(newToken.Value),
		TokenFamily:   oldToken.TokenFamily,
		RotatedFromID: &oldToken.ID, // Link to previous token
		Status:        coredomain.RefreshTokenStatusActive,
		CreatedAt:     time.Now(),
		ExpiresAt:     newToken.ExpiresAt,
		IPAddress:     oldToken.IPAddress,
		UserAgent:     oldToken.UserAgent,
	}

	if err := s.refreshTokenStore.Create(ctx, refreshTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	return newToken, nil
}

// hashToken creates SHA-256 hash of token for storage
func hashToken(tokenValue string) string {
	hash := sha256.Sum256([]byte(tokenValue))
	return hex.EncodeToString(hash[:])
}

// generateTokenID generates a unique ID for token
func generateTokenID() uuid.UUID {
	// Implement your ID generation logic (e.g., UUID)
	// This is placeholder
	return uuid.Must(uuid.NewV7())
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Example tenant config for multi-tab support:
//
// {
//   "security": {
//     "enable_refresh_token_rotation": true,
//     "refresh_token_rotation_grace_period": 30,  // 30 seconds grace period
//     "detect_refresh_token_reuse": true          // Revoke family if reuse detected
//   }
// }
//
// How it works:
//
// Time 00:00 - Tab 1 & Tab 2 both have refresh_token_A (active)
// Time 00:05 - Tab 1 access token expires → refresh using refresh_token_A
//              → Backend marks refresh_token_A as "used" (used_at = 00:05)
//              → Backend returns new refresh_token_B
//              → Tab 1 now has refresh_token_B ✅
//
// Time 00:10 - Tab 2 access token expires → refresh using refresh_token_A (old)
//              → Backend finds refresh_token_A with status="used", used_at=00:05
//              → Check grace period: now - used_at = 5 seconds
//              → 5 seconds < 30 seconds grace period ✅
//              → Allow refresh, return NEW refresh_token_C
//              → Tab 2 now has refresh_token_C ✅
//
// Time 00:40 - Attacker tries to use refresh_token_A (stolen)
//              → Backend finds refresh_token_A with status="used", used_at=00:05
//              → Check grace period: now - used_at = 35 seconds
//              → 35 seconds > 30 seconds grace period ❌
//              → Token reuse detected! Revoke entire token family
//              → All tabs logged out (security measure)
