package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-app-key-store"
type PostgresAppKeyStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.AppKeyStore = (*PostgresAppKeyStore)(nil)

func (s *PostgresAppKeyStore) Store(ctx context.Context, key *domain.AppKey) error {
	metadata, _ := json.Marshal(key.Metadata)
	scopes, _ := json.Marshal(key.Scopes)

	query := `
		INSERT INTO app_keys (
			id, tenant_id, app_id, key_id, prefix, secret_hash,
			key_type, environment, user_id, name, scopes, metadata,
			created_at, expires_at, last_used, revoked, revoked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := s.dbPool.Exec(ctx, query,
		key.ID, key.TenantID, key.AppID, key.KeyID, key.Prefix, key.SecretHash,
		key.KeyType, key.Environment, key.UserID, key.Name, scopes, metadata,
		key.CreatedAt, key.ExpiresAt, key.LastUsed, key.Revoked, key.RevokedAt,
	)
	return err
}

func (s *PostgresAppKeyStore) GetByKeyID(ctx context.Context, tenantID, appID, keyID string) (*domain.AppKey, error) {
	query := `
		SELECT id, tenant_id, app_id, key_id, prefix, secret_hash,
			   key_type, environment, user_id, name, scopes, metadata,
			   created_at, expires_at, last_used, revoked, revoked_at
		FROM app_keys
		WHERE tenant_id = $1 AND app_id = $2 AND key_id = $3
	`

	key := &domain.AppKey{}
	var scopes, metadata []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, appID, keyID).Scan(
		&key.ID, &key.TenantID, &key.AppID, &key.KeyID, &key.Prefix, &key.SecretHash,
		&key.KeyType, &key.Environment, &key.UserID, &key.Name, &scopes, &metadata,
		&key.CreatedAt, &key.ExpiresAt, &key.LastUsed, &key.Revoked, &key.RevokedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(scopes, &key.Scopes)
	json.Unmarshal(metadata, &key.Metadata)
	return key, nil
}

func (s *PostgresAppKeyStore) GetByID(ctx context.Context, id string) (*domain.AppKey, error) {
	query := `
		SELECT id, tenant_id, app_id, key_id, prefix, secret_hash,
			   key_type, environment, user_id, name, scopes, metadata,
			   created_at, expires_at, last_used, revoked, revoked_at
		FROM app_keys
		WHERE id = $1
	`

	key := &domain.AppKey{}
	var scopes, metadata []byte

	err := s.dbPool.QueryRow(ctx, query, id).Scan(
		&key.ID, &key.TenantID, &key.AppID, &key.KeyID, &key.Prefix, &key.SecretHash,
		&key.KeyType, &key.Environment, &key.UserID, &key.Name, &scopes, &metadata,
		&key.CreatedAt, &key.ExpiresAt, &key.LastUsed, &key.Revoked, &key.RevokedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("key not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(scopes, &key.Scopes)
	json.Unmarshal(metadata, &key.Metadata)
	return key, nil
}

func (s *PostgresAppKeyStore) GetByPrefix(ctx context.Context, prefix string) ([]*domain.AppKey, error) {
	query := `
		SELECT id, tenant_id, app_id, key_id, prefix, secret_hash,
			   key_type, environment, user_id, name, scopes, metadata,
			   created_at, expires_at, last_used, revoked, revoked_at
		FROM app_keys
		WHERE prefix = $1
	`

	rows, err := s.dbPool.Query(ctx, query, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanAppKeys(rows)
}

func (s *PostgresAppKeyStore) ListByApp(ctx context.Context, tenantID, appID string) ([]*domain.AppKey, error) {
	query := `
		SELECT id, tenant_id, app_id, key_id, prefix, secret_hash,
			   key_type, environment, user_id, name, scopes, metadata,
			   created_at, expires_at, last_used, revoked, revoked_at
		FROM app_keys
		WHERE tenant_id = $1 AND app_id = $2
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanAppKeys(rows)
}

func (s *PostgresAppKeyStore) ListByTenant(ctx context.Context, tenantID string) ([]*domain.AppKey, error) {
	query := `
		SELECT id, tenant_id, app_id, key_id, prefix, secret_hash,
			   key_type, environment, user_id, name, scopes, metadata,
			   created_at, expires_at, last_used, revoked, revoked_at
		FROM app_keys
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanAppKeys(rows)
}

func (s *PostgresAppKeyStore) Update(ctx context.Context, key *domain.AppKey) error {
	metadata, _ := json.Marshal(key.Metadata)
	scopes, _ := json.Marshal(key.Scopes)

	query := `
		UPDATE app_keys
		SET secret_hash = $1, key_type = $2, environment = $3, name = $4,
		    scopes = $5, metadata = $6, expires_at = $7, last_used = $8,
		    revoked = $9, revoked_at = $10
		WHERE id = $11
	`

	result, err := s.dbPool.Exec(ctx, query,
		key.SecretHash, key.KeyType, key.Environment, key.Name,
		scopes, metadata, key.ExpiresAt, key.LastUsed,
		key.Revoked, key.RevokedAt, key.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("key not found: %s", key.ID)
	}
	return nil
}

func (s *PostgresAppKeyStore) Revoke(ctx context.Context, tenantID, appID, keyID string) error {
	query := `
		UPDATE app_keys
		SET revoked = true, revoked_at = $1
		WHERE tenant_id = $2 AND app_id = $3 AND key_id = $4
	`

	now := time.Now()
	result, err := s.dbPool.Exec(ctx, query, now, tenantID, appID, keyID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("key not found: %s", keyID)
	}
	return nil
}

func (s *PostgresAppKeyStore) Delete(ctx context.Context, tenantID, appID, keyID string) error {
	query := `DELETE FROM app_keys WHERE tenant_id = $1 AND app_id = $2 AND key_id = $3`

	result, err := s.dbPool.Exec(ctx, query, tenantID, appID, keyID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("key not found: %s", keyID)
	}
	return nil
}

func (s *PostgresAppKeyStore) UpdateLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	query := `UPDATE app_keys SET last_used = $1 WHERE id = $2`

	_, err := s.dbPool.Exec(ctx, query, lastUsed, id)
	return err
}

func (s *PostgresAppKeyStore) scanAppKeys(rows serviceapi.Rows) ([]*domain.AppKey, error) {
	keys := make([]*domain.AppKey, 0)

	for rows.Next() {
		key := &domain.AppKey{}
		var scopes, metadata []byte

		err := rows.Scan(
			&key.ID, &key.TenantID, &key.AppID, &key.KeyID, &key.Prefix, &key.SecretHash,
			&key.KeyType, &key.Environment, &key.UserID, &key.Name, &scopes, &metadata,
			&key.CreatedAt, &key.ExpiresAt, &key.LastUsed, &key.Revoked, &key.RevokedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(scopes, &key.Scopes)
		json.Unmarshal(metadata, &key.Metadata)
		keys = append(keys, key)
	}

	return keys, rows.Err()
}
