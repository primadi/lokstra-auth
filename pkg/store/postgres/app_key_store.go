package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for app_key operations
const (
	appKeyColumns = `id, tenant_id, app_id, branch_id, key_id, prefix, secret_hash, key_type, environment, name, description, status, scopes, expires_at, last_used_at, created_by, metadata, created_at, updated_at, deleted_at`

	queryAppKeyInsert = `
		INSERT INTO app_keys (
			id, tenant_id, app_id, branch_id, key_id, prefix, secret_hash, key_type, environment, name, description, status, scopes, expires_at, created_by, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	queryAppKeySelect = `SELECT ` + appKeyColumns + ` FROM app_keys`

	queryAppKeyGet        = queryAppKeySelect + ` WHERE id = $1 AND deleted_at IS NULL`
	queryAppKeyGetByKeyID = queryAppKeySelect + ` WHERE key_id = $1 AND deleted_at IS NULL`
	queryAppKeyListByApp  = queryAppKeySelect + ` WHERE tenant_id = $1 AND app_id = $2 AND deleted_at IS NULL ORDER BY created_at DESC`

	queryAppKeyUpdate = `
		UPDATE app_keys
		SET name = $1, description = $2, status = $3, metadata = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL`

	queryAppKeyRevoke = `
		UPDATE app_keys SET status = 'revoked', updated_at = NOW(), deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
)

// @Service "pg-app-key-store"
type pgAppKeyStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Store implements [store.AppKeyStore].
func (p *pgAppKeyStore) Store(ctx context.Context, appKey *domaincore.AppKey) error {
	metadata, err := json.Marshal(appKey.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryAppKeyInsert,
		appKey.ID, appKey.TenantID, appKey.AppID, appKey.BranchID, appKey.KeyID, appKey.Prefix,
		appKey.SecretHash, appKey.KeyType, appKey.Environment, appKey.Name,
		appKey.Description, appKey.Status, appKey.Scopes, appKey.ExpiresAt, appKey.CreatedBy, metadata,
	)
	return err
}

// GetByID implements [store.AppKeyStore].
func (p *pgAppKeyStore) GetByID(ctx context.Context, id string) (*domaincore.AppKey, error) {
	appKey := &domaincore.AppKey{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryAppKeyGet, id).Scan(
		&appKey.ID, &appKey.TenantID, &appKey.AppID, &appKey.BranchID, &appKey.KeyID, &appKey.Prefix,
		&appKey.SecretHash, &appKey.KeyType, &appKey.Environment, &appKey.Name,
		&appKey.Description, &appKey.Status, &appKey.Scopes, &appKey.ExpiresAt, &appKey.LastUsedAt,
		&appKey.CreatedBy, &metadata, &appKey.CreatedAt, &appKey.UpdatedAt, &appKey.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("app key not found: id=%s", id)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalAppKeyFields(appKey, metadata); err != nil {
		return nil, err
	}

	return appKey, nil
}

// GetByKeyID implements [store.AppKeyStore].
func (p *pgAppKeyStore) GetByKeyID(ctx context.Context, tenantID, appID, keyID string) (*domaincore.AppKey, error) {
	appKey := &domaincore.AppKey{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryAppKeyGetByKeyID, keyID).Scan(
		&appKey.ID, &appKey.TenantID, &appKey.AppID, &appKey.BranchID, &appKey.KeyID, &appKey.Prefix,
		&appKey.SecretHash, &appKey.KeyType, &appKey.Environment, &appKey.Name,
		&appKey.Description, &appKey.Status, &appKey.Scopes, &appKey.ExpiresAt, &appKey.LastUsedAt,
		&appKey.CreatedBy, &metadata, &appKey.CreatedAt, &appKey.UpdatedAt, &appKey.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("app key not found: key_id=%s", keyID)
	}
	if err != nil {
		return nil, err
	}

	// Verify tenant and app match
	if appKey.TenantID != tenantID || appKey.AppID != appID {
		return nil, fmt.Errorf("app key not found: tenant=%s, app=%s, key_id=%s", tenantID, appID, keyID)
	}

	if err := p.unmarshalAppKeyFields(appKey, metadata); err != nil {
		return nil, err
	}

	return appKey, nil
}

// Update updates an API key
func (p *pgAppKeyStore) Update(ctx context.Context, appKey *domaincore.AppKey) error {
	metadata, err := json.Marshal(appKey.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryAppKeyUpdate,
		appKey.Name, appKey.Description, appKey.Status, metadata, appKey.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("app key not found: id=%s", appKey.ID)
	}
	return nil
}

// ListByApp implements [store.AppKeyStore].
func (p *pgAppKeyStore) ListByApp(ctx context.Context, tenantID, appID string) ([]*domaincore.AppKey, error) {
	rows, err := p.dbPool.Query(ctx, queryAppKeyListByApp, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanAppKeys(rows)
}

// ListByTenant implements [store.AppKeyStore].
func (p *pgAppKeyStore) ListByTenant(ctx context.Context, tenantID string) ([]*domaincore.AppKey, error) {
	query := queryAppKeySelect + ` WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := p.dbPool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanAppKeys(rows)
}

// Revoke implements [store.AppKeyStore].
func (p *pgAppKeyStore) Revoke(ctx context.Context, tenantID, appID, keyID string) error {
	// Use queryAppKeyDelete for soft delete (revoke)
	result, err := p.dbPool.Exec(ctx, queryAppKeyRevoke, tenantID, appID, keyID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("app key not found: tenant=%s, app=%s, key_id=%s", tenantID, appID, keyID)
	}
	return nil
}

var _ store.AppKeyStore = (*pgAppKeyStore)(nil)

// unmarshalAppKeyFields unmarshals JSON fields into appKey struct
func (p *pgAppKeyStore) unmarshalAppKeyFields(appKey *domaincore.AppKey, metadata []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		appKey.Metadata = &m
	}
	return nil
}

func (p *pgAppKeyStore) scanAppKeys(rows serviceapi.Rows) ([]*domaincore.AppKey, error) {
	appKeys := make([]*domaincore.AppKey, 0, 10)

	for rows.Next() {
		appKey := &domaincore.AppKey{}
		var metadata []byte

		err := rows.Scan(
			&appKey.ID, &appKey.TenantID, &appKey.AppID, &appKey.BranchID, &appKey.KeyID, &appKey.Prefix,
			&appKey.SecretHash, &appKey.KeyType, &appKey.Environment, &appKey.Name,
			&appKey.Description, &appKey.Status, &appKey.Scopes, &appKey.ExpiresAt, &appKey.LastUsedAt,
			&appKey.CreatedBy, &metadata, &appKey.CreatedAt, &appKey.UpdatedAt, &appKey.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalAppKeyFields(appKey, metadata); err != nil {
			return nil, err
		}

		appKeys = append(appKeys, appKey)
	}

	return appKeys, rows.Err()
}
