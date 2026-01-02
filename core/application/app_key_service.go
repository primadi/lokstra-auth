package application

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/core/infrastructure/crypto"
	"github.com/primadi/lokstra-auth/core/infrastructure/idgen"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/core/request"
)

// AppKeyService manages API keys for service-to-service authentication
// @RouterService name="app-key-service", prefix="${api-auth-prefix:/api/auth}/core/tenants/{tenant_id}/apps/{app_id}/keys", middlewares=["recovery", "request_logger", "auth"]
type AppKeyService struct {
	// @Inject "@store.app-key-store"
	Store repository.AppKeyStore
	// @Inject "app-service"
	AppService *AppService
}

// GenerateKey generates a new API key for an app
// @Route "POST /"
func (s *AppKeyService) GenerateKey(ctx *request.Context, req *domain.GenerateAppKeyRequest) (*domain.AppKeyResponse, error) {
	// Validate app exists and is active
	app, err := s.AppService.GetApp(ctx, &domain.GetAppRequest{
		TenantID: req.TenantID,
		ID:       req.AppID,
	})
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	if app.Status != domain.AppStatusActive {
		return nil, fmt.Errorf("app is not active: %s", app.Status)
	}

	// Generate key ID
	keyID := idgen.GenerateID("appkey")

	// Generate a cryptographically secure secret (32 bytes = 256 bits)
	secret, err := crypto.GenerateSecureSecret(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure secret: %w", err)
	}

	// Hash the secret before storing (SHA3-256)
	secretHash, err := crypto.HashSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash secret: %w", err)
	}

	// Determine environment
	env := "live"
	if req.Environment == "test" {
		env = "test"
	}

	// Generate the full key string (this includes the plain secret - shown ONLY once!)
	keyString := fmt.Sprintf("%s_%s.%s", req.AppID, keyID, secret)

	// Create API key
	apiKey := &domain.AppKey{
		ID:          idgen.GenerateID("id"),
		TenantID:    req.TenantID,
		AppID:       req.AppID,
		KeyID:       keyID,
		SecretHash:  secretHash, // ✅ Store HASHED secret (SHA3-256), not plain text!
		Prefix:      fmt.Sprintf("sk_%s", env),
		KeyType:     "secret",
		Environment: env,
		UserID:      fmt.Sprintf("app:%s", req.AppID),
		Name:        req.Name,
		Scopes:      req.Scopes,
		Metadata: map[string]any{
			"key_owner_type": "app",
			"app_name":       app.Name,
			"purpose":        req.Purpose,
		},
		CreatedAt: time.Now(),
	}

	if req.Description != "" {
		apiKey.Metadata["description"] = req.Description
	}

	if req.ExpiresIn != nil {
		expiresAt := time.Now().Add(*req.ExpiresIn)
		apiKey.ExpiresAt = &expiresAt
	}

	// Save to store
	if err := s.Store.Store(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to store app key: %w", err)
	}

	return &domain.AppKeyResponse{
		KeyID:     keyID,
		KeyString: keyString, // Only shown once!
		AppKey:    apiKey,
	}, nil
}

// GetKey retrieves an API key by ID (sanitized - no secret hash)
// @Route "GET /{key_id}"
func (s *AppKeyService) GetKey(ctx *request.Context, req *domain.GetAppKeyRequest) (*domain.AppKeyInfo, error) {
	// Query with full tenant+app+key scope for multi-tenant isolation
	key, err := s.Store.GetByKeyID(ctx, req.TenantID, req.AppID, req.KeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app key: %w", err)
	}

	// Return sanitized info (no secret hash!)
	return domain.ToAppKeyInfo(key), nil
}

// ListKeys lists all API keys for an app (sanitized - no secret hashes)
// @Route "GET /"
func (s *AppKeyService) ListKeys(ctx *request.Context, req *domain.ListAppKeysRequest) ([]*domain.AppKeyInfo, error) {
	keys, err := s.Store.ListByApp(ctx, req.TenantID, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to list app keys: %w", err)
	}

	// Return sanitized list (no secret hashes!)
	return domain.ToAppKeyInfoList(keys), nil
}

// RevokeKey revokes an API key
// @Route "POST /{key_id}/revoke"
func (s *AppKeyService) RevokeKey(ctx *request.Context, req *domain.RevokeAppKeyRequest) error {
	// Get the key
	key, err := s.Store.GetByKeyID(ctx, req.TenantID, req.AppID, req.KeyID)
	if err != nil {
		return fmt.Errorf("key not found: %w", err)
	}

	// Check if already revoked
	if key.Revoked {
		return fmt.Errorf("key is already revoked")
	}

	// Revoke the key
	if err := s.Store.Revoke(ctx, key.TenantID, key.AppID, key.KeyID); err != nil {
		return fmt.Errorf("failed to revoke app key: %w", err)
	}

	return nil
}

// DeleteKey permanently deletes an API key
// @Route "DELETE /{key_id}"
func (s *AppKeyService) DeleteKey(ctx *request.Context, req *domain.DeleteAppKeyRequest) error {
	// Verify key exists and belongs to this app
	key, err := s.Store.GetByKeyID(ctx, req.TenantID, req.AppID, req.KeyID)
	if err != nil {
		return fmt.Errorf("key not found: %w", err)
	}

	// Delete the key
	if err := s.Store.Delete(ctx, key.TenantID, key.AppID, key.KeyID); err != nil {
		return fmt.Errorf("failed to delete app key: %w", err)
	}

	return nil
}

// RotateKey rotates an API key (revokes old, generates new with same permissions)
// @Route "POST /{key_id}/rotate"
func (s *AppKeyService) RotateKey(ctx *request.Context, req *domain.RotateAppKeyRequest) (*domain.AppKeyResponse, error) {
	// Get old key
	oldKey, err := s.Store.GetByKeyID(ctx, req.TenantID, req.AppID, req.KeyID)
	if err != nil {
		return nil, fmt.Errorf("old key not found: %w", err)
	}

	if oldKey.Revoked {
		return nil, fmt.Errorf("cannot rotate revoked key")
	}

	// Generate new key with same properties
	newKeyReq := &domain.GenerateAppKeyRequest{
		TenantID:    oldKey.TenantID,
		AppID:       oldKey.AppID,
		Name:        fmt.Sprintf("%s (rotated)", oldKey.Name),
		Purpose:     "rotation",
		Description: fmt.Sprintf("Rotated from key %s", oldKey.KeyID),
		Environment: string(oldKey.Environment),
		Scopes:      oldKey.Scopes,
	}

	// Copy expiration logic
	if oldKey.ExpiresAt != nil {
		duration := time.Until(*oldKey.ExpiresAt)
		if duration > 0 {
			newKeyReq.ExpiresIn = &duration
		}
	}

	// Generate new key
	newKeyResp, err := s.GenerateKey(ctx, newKeyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new key: %w", err)
	}

	// Add rotation metadata
	newKeyResp.AppKey.Metadata["rotated_from"] = oldKey.KeyID
	newKeyResp.AppKey.Metadata["rotated_at"] = time.Now()

	if err := s.Store.Update(ctx, newKeyResp.AppKey); err != nil {
		return nil, fmt.Errorf("failed to update rotation metadata: %w", err)
	}

	// Revoke old key
	if err := s.Store.Revoke(ctx, oldKey.TenantID, oldKey.AppID, oldKey.KeyID); err != nil {
		return nil, fmt.Errorf("failed to revoke old key: %w", err)
	}

	return newKeyResp, nil
}
