package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for refresh token operations
const (
	refreshTokenColumns = `id, tenant_id, user_id, app_id, session_id, token_hash, token_family, rotated_from_id, status, created_at, expires_at, used_at, revoked_at, ip_address, user_agent, metadata`

	queryRefreshTokenInsert = `
		INSERT INTO refresh_tokens (
			id, tenant_id, user_id, app_id, session_id, token_hash, token_family, rotated_from_id, status, expires_at, ip_address, user_agent, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	queryRefreshTokenSelect = `SELECT ` + refreshTokenColumns + ` FROM refresh_tokens`

	queryRefreshTokenGet           = queryRefreshTokenSelect + ` WHERE id = $1`
	queryRefreshTokenGetByToken    = queryRefreshTokenSelect + ` WHERE token_hash = $1`
	queryRefreshTokenListByUser    = queryRefreshTokenSelect + ` WHERE tenant_id = $1 AND user_id = $2 ORDER BY created_at DESC`
	queryRefreshTokenListBySession = queryRefreshTokenSelect + ` WHERE session_id = $1 ORDER BY created_at DESC`
	queryRefreshTokenExists        = `SELECT 1 FROM refresh_tokens WHERE id = $1`

	queryRefreshTokenUpdate = `
		UPDATE refresh_tokens
		SET status = $1, metadata = $2
		WHERE id = $3`

	queryRefreshTokenMarkUsed = `
		UPDATE refresh_tokens
		SET status = 'used', used_at = NOW()
		WHERE id = $1`

	queryRefreshTokenRevoke = `
		UPDATE refresh_tokens
		SET status = 'revoked', revoked_at = NOW()
		WHERE id = $1`

	queryRefreshTokenRevokeFamily = `
		UPDATE refresh_tokens
		SET status = 'revoked', revoked_at = NOW()
		WHERE token_family = $1 AND status = 'active'`

	queryRefreshTokenRevokeAllByUser = `
		UPDATE refresh_tokens
		SET status = 'revoked', revoked_at = NOW()
		WHERE tenant_id = $1 AND user_id = $2 AND status = 'active'`

	queryRefreshTokenCountActiveByUser = `
		SELECT COUNT(*) FROM refresh_tokens
		WHERE tenant_id = $1 AND user_id = $2 AND status = 'active' AND expires_at > NOW()`

	queryRefreshTokenRevokeOldestByUser = `
		UPDATE refresh_tokens
		SET status = 'revoked', revoked_at = NOW()
		WHERE id = (
			SELECT id FROM refresh_tokens
			WHERE tenant_id = $1 AND user_id = $2 AND status = 'active' AND expires_at > NOW()
			ORDER BY created_at ASC
			LIMIT 1
		)`

	queryRefreshTokenDelete = `DELETE FROM refresh_tokens WHERE id = $1`

	queryRefreshTokenCleanupExpired = `
		DELETE FROM refresh_tokens
		WHERE status IN ('active', 'used') AND expires_at < NOW()`
)

// @Service "pg-refresh-token-store"
type pgRefreshTokenStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) Create(ctx context.Context, token *domaincore.RefreshToken) error {
	metadata, err := json.Marshal(token.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryRefreshTokenInsert,
		token.ID, token.TenantID, token.UserID, token.AppID, token.SessionID,
		token.TokenHash, token.TokenFamily, token.RotatedFromID, token.Status, token.ExpiresAt,
		token.IPAddress, token.UserAgent, metadata,
	)
	return err
}

// Get implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) Get(ctx context.Context, id uuid.UUID) (*domaincore.RefreshToken, error) {
	token := &domaincore.RefreshToken{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryRefreshTokenGet, id).Scan(
		&token.ID, &token.TenantID, &token.UserID, &token.AppID, &token.SessionID,
		&token.TokenHash, &token.TokenFamily, &token.RotatedFromID, &token.Status, &token.CreatedAt,
		&token.ExpiresAt, &token.UsedAt, &token.RevokedAt,
		&token.IPAddress, &token.UserAgent, &metadata,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("refresh token not found: id=%s", id)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalRefreshTokenFields(token, metadata); err != nil {
		return nil, err
	}

	return token, nil
}

// GetByToken implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) GetByToken(ctx context.Context, tokenHash string) (*domaincore.RefreshToken, error) {
	token := &domaincore.RefreshToken{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryRefreshTokenGetByToken, tokenHash).Scan(
		&token.ID, &token.TenantID, &token.UserID, &token.AppID, &token.SessionID,
		&token.TokenHash, &token.TokenFamily, &token.RotatedFromID, &token.Status, &token.CreatedAt,
		&token.ExpiresAt, &token.UsedAt, &token.RevokedAt,
		&token.IPAddress, &token.UserAgent, &metadata,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("refresh token not found by token")
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalRefreshTokenFields(token, metadata); err != nil {
		return nil, err
	}

	return token, nil
}

// Update implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) Update(ctx context.Context, token *domaincore.RefreshToken) error {
	metadata, err := json.Marshal(token.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryRefreshTokenUpdate,
		token.Status, metadata, token.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found: id=%s", token.ID)
	}
	return nil
}

// MarkUsed implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) MarkUsed(ctx context.Context, id uuid.UUID) error {
	result, err := p.dbPool.Exec(ctx, queryRefreshTokenMarkUsed, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found: id=%s", id)
	}
	return nil
}

// Revoke implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) Revoke(ctx context.Context, id uuid.UUID) error {
	result, err := p.dbPool.Exec(ctx, queryRefreshTokenRevoke, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found: id=%s", id)
	}
	return nil
}

// RevokeFamily implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) RevokeFamily(ctx context.Context, tokenFamily string) error {
	_, err := p.dbPool.Exec(ctx, queryRefreshTokenRevokeFamily, tokenFamily)
	return err
}

// Delete implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := p.dbPool.Exec(ctx, queryRefreshTokenDelete, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("refresh token not found: id=%s", id)
	}
	return nil
}

// ListByUser implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) ListByUser(ctx context.Context, tenantID, userID string) ([]*domaincore.RefreshToken, error) {
	rows, err := p.dbPool.Query(ctx, queryRefreshTokenListByUser, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanRefreshTokens(rows)
}

// ListBySession implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) ListBySession(ctx context.Context, sessionID string) ([]*domaincore.RefreshToken, error) {
	rows, err := p.dbPool.Query(ctx, queryRefreshTokenListBySession, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanRefreshTokens(rows)
}

// RevokeAllByUser implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) RevokeAllByUser(ctx context.Context, tenantID, userID string) error {
	_, err := p.dbPool.Exec(ctx, queryRefreshTokenRevokeAllByUser, tenantID, userID)
	return err
}

// CountActiveByUser implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) CountActiveByUser(ctx context.Context, tenantID, userID string) (int, error) {
	var count int
	err := p.dbPool.QueryRow(ctx, queryRefreshTokenCountActiveByUser, tenantID, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// RevokeOldestByUser implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) RevokeOldestByUser(ctx context.Context, tenantID, userID string) error {
	result, err := p.dbPool.Exec(ctx, queryRefreshTokenRevokeOldestByUser, tenantID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no active refresh token found for user: tenant_id=%s user_id=%s", tenantID, userID)
	}
	return nil
}

// CleanupExpired implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) CleanupExpired(ctx context.Context) (int, error) {
	result, err := p.dbPool.Exec(ctx, queryRefreshTokenCleanupExpired)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// Exists implements [store.RefreshTokenStore].
func (p *pgRefreshTokenStore) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return p.dbPool.IsExists(ctx, queryRefreshTokenExists, id)
}

var _ store.RefreshTokenStore = (*pgRefreshTokenStore)(nil)

// unmarshalRefreshTokenFields unmarshals JSON fields into token struct
func (p *pgRefreshTokenStore) unmarshalRefreshTokenFields(token *domaincore.RefreshToken, metadata []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		token.Metadata = &m
	}
	return nil
}

func (p *pgRefreshTokenStore) scanRefreshTokens(rows serviceapi.Rows) ([]*domaincore.RefreshToken, error) {
	tokens := make([]*domaincore.RefreshToken, 0, 10)

	for rows.Next() {
		token := &domaincore.RefreshToken{}
		var metadata []byte

		err := rows.Scan(
			&token.ID, &token.TenantID, &token.UserID, &token.AppID, &token.SessionID,
			&token.TokenHash, &token.TokenFamily, &token.RotatedFromID, &token.Status, &token.CreatedAt,
			&token.ExpiresAt, &token.UsedAt, &token.RevokedAt,
			&token.IPAddress, &token.UserAgent, &metadata,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalRefreshTokenFields(token, metadata); err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}
