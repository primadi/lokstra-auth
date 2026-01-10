package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for session operations
const (
	sessionColumns = `id, tenant_id, user_id, app_id, session_token_hash, ip_address, user_agent, status, created_at, last_activity_at, expires_at, revoked_at, metadata`

	querySessionInsert = `
		INSERT INTO sessions (
			id, tenant_id, user_id, app_id, session_token_hash, ip_address, user_agent, status, expires_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	querySessionSelect = `SELECT ` + sessionColumns + ` FROM sessions`

	querySessionGet        = querySessionSelect + ` WHERE id = $1`
	querySessionGetByToken = querySessionSelect + ` WHERE session_token_hash = $1`
	querySessionListByUser = querySessionSelect + ` WHERE tenant_id = $1 AND user_id = $2 ORDER BY created_at DESC`
	querySessionListActive = querySessionSelect + ` WHERE tenant_id = $1 AND user_id = $2 AND status = 'active' AND expires_at > NOW() ORDER BY last_activity_at DESC`
	querySessionExists     = `SELECT 1 FROM sessions WHERE id = $1`

	querySessionUpdate = `
		UPDATE sessions
		SET status = $1, metadata = $2, updated_at = NOW()
		WHERE id = $3`

	querySessionUpdateLastActivity = `
		UPDATE sessions
		SET last_activity_at = NOW()
		WHERE id = $1`

	querySessionRevoke = `
		UPDATE sessions
		SET status = 'revoked', revoked_at = NOW()
		WHERE id = $1`

	querySessionRevokeAllByUser = `
		UPDATE sessions
		SET status = 'revoked', revoked_at = NOW()
		WHERE tenant_id = $1 AND user_id = $2 AND status = 'active'`

	querySessionDelete = `DELETE FROM sessions WHERE id = $1`

	querySessionCleanupExpired = `
		DELETE FROM sessions
		WHERE status = 'active' AND expires_at < NOW()`
)

// @Service "pg-session-store"
type pgSessionStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.SessionStore].
func (p *pgSessionStore) Create(ctx context.Context, session *domaincore.Session) error {
	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, querySessionInsert,
		session.ID, session.TenantID, session.UserID, session.AppID,
		session.SessionTokenHash, session.IPAddress, session.UserAgent,
		session.Status, session.ExpiresAt, metadata,
	)
	return err
}

// Get implements [store.SessionStore].
func (p *pgSessionStore) Get(ctx context.Context, id string) (*domaincore.Session, error) {
	session := &domaincore.Session{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, querySessionGet, id).Scan(
		&session.ID, &session.TenantID, &session.UserID, &session.AppID,
		&session.SessionTokenHash, &session.IPAddress, &session.UserAgent,
		&session.Status, &session.CreatedAt, &session.LastActivityAt,
		&session.ExpiresAt, &session.RevokedAt, &metadata,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("session not found: id=%s", id)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalSessionFields(session, metadata); err != nil {
		return nil, err
	}

	return session, nil
}

// GetByToken implements [store.SessionStore].
func (p *pgSessionStore) GetByToken(ctx context.Context, tokenHash string) (*domaincore.Session, error) {
	session := &domaincore.Session{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, querySessionGetByToken, tokenHash).Scan(
		&session.ID, &session.TenantID, &session.UserID, &session.AppID,
		&session.SessionTokenHash, &session.IPAddress, &session.UserAgent,
		&session.Status, &session.CreatedAt, &session.LastActivityAt,
		&session.ExpiresAt, &session.RevokedAt, &metadata,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("session not found by token")
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalSessionFields(session, metadata); err != nil {
		return nil, err
	}

	return session, nil
}

// Update implements [store.SessionStore].
func (p *pgSessionStore) Update(ctx context.Context, session *domaincore.Session) error {
	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, querySessionUpdate,
		session.Status, metadata, session.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: id=%s", session.ID)
	}
	return nil
}

// UpdateLastActivity implements [store.SessionStore].
func (p *pgSessionStore) UpdateLastActivity(ctx context.Context, id string) error {
	result, err := p.dbPool.Exec(ctx, querySessionUpdateLastActivity, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: id=%s", id)
	}
	return nil
}

// Revoke implements [store.SessionStore].
func (p *pgSessionStore) Revoke(ctx context.Context, id string) error {
	result, err := p.dbPool.Exec(ctx, querySessionRevoke, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: id=%s", id)
	}
	return nil
}

// Delete implements [store.SessionStore].
func (p *pgSessionStore) Delete(ctx context.Context, id string) error {
	result, err := p.dbPool.Exec(ctx, querySessionDelete, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: id=%s", id)
	}
	return nil
}

// ListByUser implements [store.SessionStore].
func (p *pgSessionStore) ListByUser(ctx context.Context, tenantID, userID string) ([]*domaincore.Session, error) {
	rows, err := p.dbPool.Query(ctx, querySessionListByUser, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanSessions(rows)
}

// ListActive implements [store.SessionStore].
func (p *pgSessionStore) ListActive(ctx context.Context, tenantID, userID string) ([]*domaincore.Session, error) {
	rows, err := p.dbPool.Query(ctx, querySessionListActive, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanSessions(rows)
}

// RevokeAllByUser implements [store.SessionStore].
func (p *pgSessionStore) RevokeAllByUser(ctx context.Context, tenantID, userID string) error {
	_, err := p.dbPool.Exec(ctx, querySessionRevokeAllByUser, tenantID, userID)
	return err
}

// CleanupExpired implements [store.SessionStore].
func (p *pgSessionStore) CleanupExpired(ctx context.Context) (int, error) {
	result, err := p.dbPool.Exec(ctx, querySessionCleanupExpired)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// Exists implements [store.SessionStore].
func (p *pgSessionStore) Exists(ctx context.Context, id string) (bool, error) {
	return p.dbPool.IsExists(ctx, querySessionExists, id)
}

var _ store.SessionStore = (*pgSessionStore)(nil)

// unmarshalSessionFields unmarshals JSON fields into session struct
func (p *pgSessionStore) unmarshalSessionFields(session *domaincore.Session, metadata []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		session.Metadata = &m
	}
	return nil
}

func (p *pgSessionStore) scanSessions(rows serviceapi.Rows) ([]*domaincore.Session, error) {
	sessions := make([]*domaincore.Session, 0, 10)

	for rows.Next() {
		session := &domaincore.Session{}
		var metadata []byte

		err := rows.Scan(
			&session.ID, &session.TenantID, &session.UserID, &session.AppID,
			&session.SessionTokenHash, &session.IPAddress, &session.UserAgent,
			&session.Status, &session.CreatedAt, &session.LastActivityAt,
			&session.ExpiresAt, &session.RevokedAt, &metadata,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalSessionFields(session, metadata); err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}
