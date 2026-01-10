package tenant

import (
	"context"
	"errors"
	"fmt"

	coredomain "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
)

// Example: How to enforce max concurrent sessions limit during login

var (
	ErrMaxSessionsReached = errors.New("maximum concurrent sessions reached")
)

// EnforceSessionLimit checks and enforces max concurrent session limit
// Call this during login flow before creating new session
func EnforceSessionLimit(
	ctx context.Context,
	refreshTokenStore store.RefreshTokenStore,
	tenantConfig *coredomain.TenantConfig,
	tenantID, userID string,
) error {
	// Skip if concurrent sessions not limited
	if tenantConfig == nil ||
		tenantConfig.Security == nil ||
		tenantConfig.Security.MaxConcurrentSessions <= 0 {
		return nil
	}

	maxSessions := tenantConfig.Security.MaxConcurrentSessions
	policy := tenantConfig.Security.SessionRevocationPolicy

	// Count active sessions (via active refresh tokens)
	activeCount, err := refreshTokenStore.CountActiveByUser(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to count active sessions: %w", err)
	}

	// Check if limit exceeded
	if activeCount >= maxSessions {
		switch policy {
		case coredomain.SessionRevocationPolicyOldest:
			// Auto-revoke oldest session to make room for new one
			if err := refreshTokenStore.RevokeOldestByUser(ctx, tenantID, userID); err != nil {
				return fmt.Errorf("failed to revoke oldest session: %w", err)
			}
			return nil

		case coredomain.SessionRevocationPolicyNewest, coredomain.SessionRevocationPolicyReject:
			// Reject new login attempt
			return ErrMaxSessionsReached

		default:
			// Default: reject
			return ErrMaxSessionsReached
		}
	}

	return nil
}

// Example usage in login flow:
//
// func (s *AuthService) Login(ctx context.Context, credentials *Credentials) (*TokenPair, error) {
//     // 1. Authenticate user
//     user, err := s.authenticateUser(ctx, credentials)
//     if err != nil {
//         return nil, err
//     }
//
//     // 2. Get tenant config
//     tenant, err := s.tenantStore.Get(ctx, user.TenantID)
//     if err != nil {
//         return nil, err
//     }
//
//     // 3. Enforce session limit BEFORE creating new session
//     if err := EnforceSessionLimit(ctx, s.refreshTokenStore, tenant.Config, user.TenantID, user.ID); err != nil {
//         if errors.Is(err, ErrMaxSessionsReached) {
//             return nil, fmt.Errorf("you have reached the maximum number of concurrent sessions (%d). Please logout from another device first",
//                 tenant.Config.Security.MaxConcurrentSessions)
//         }
//         return nil, err
//     }
//
//     // 4. Create new session + tokens (now we have room)
//     session := s.createSession(user)
//     refreshToken := s.createRefreshToken(user, session)
//     accessToken := s.createAccessToken(user)
//
//     return &TokenPair{
//         AccessToken:  accessToken,
//         RefreshToken: refreshToken,
//     }, nil
// }

// Example tenant config setup:
//
// Option 1: Strict - Max 5 sessions, reject new logins
//   {
//     "security": {
//       "allow_concurrent_sessions": true,
//       "max_concurrent_sessions": 5,
//       "session_revocation_policy": "reject"
//     }
//   }
//
// Option 2: Auto-revoke - Max 3 sessions, auto logout oldest
//   {
//     "security": {
//       "allow_concurrent_sessions": true,
//       "max_concurrent_sessions": 3,
//       "session_revocation_policy": "oldest"
//     }
//   }
//
// Option 3: Unlimited - No limit
//   {
//     "security": {
//       "allow_concurrent_sessions": true,
//       "max_concurrent_sessions": 0
//     }
//   }
//
// Option 4: Single session only
//   {
//     "security": {
//       "allow_concurrent_sessions": false,
//       "max_concurrent_sessions": 1,
//       "session_revocation_policy": "oldest"
//     }
//   }
