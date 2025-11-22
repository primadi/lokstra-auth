package postgres

import (
	"context"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

// ============================================================
// PostgreSQL User-App Store
// ============================================================

type PostgresUserAppStore struct {
	dbPool serviceapi.DbPoolWithSchema
}

func NewUserAppStore(dbPool serviceapi.DbPoolWithSchema) *PostgresUserAppStore {
	return &PostgresUserAppStore{dbPool: dbPool}
}

func (s *PostgresUserAppStore) GrantAccess(ctx context.Context, tenantID, appID, userID string) error {
	query := `
		INSERT INTO user_apps (tenant_id, app_id, user_id, status, granted_at)
		VALUES ($1, $2, $3, 'active', $4)
		ON CONFLICT (tenant_id, app_id, user_id)
		DO UPDATE SET status = 'active', granted_at = $4, revoked_at = NULL
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	now := time.Now()
	_, err = cn.Exec(ctx, query, tenantID, appID, userID, now)

	return err
}

func (s *PostgresUserAppStore) RevokeAccess(ctx context.Context, tenantID, appID, userID string) error {
	query := `
		UPDATE user_apps
		SET status = 'revoked', revoked_at = $1
		WHERE tenant_id = $2 AND app_id = $3 AND user_id = $4
	`

	now := time.Now()

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query, now, tenantID, appID, userID)
	return err
}

func (s *PostgresUserAppStore) HasAccess(ctx context.Context, tenantID, appID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_apps
			WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3 AND status = 'active'
		)
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer cn.Release()

	var hasAccess bool
	err = cn.QueryRow(ctx, query, tenantID, appID, userID).Scan(&hasAccess)

	return hasAccess, err
}

func (s *PostgresUserAppStore) ListUserApps(ctx context.Context, tenantID, userID string) ([]string, error) {
	query := `
		SELECT app_id
		FROM user_apps
		WHERE tenant_id = $1 AND user_id = $2 AND status = 'active'
		ORDER BY granted_at DESC
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appIDs := make([]string, 0)
	for rows.Next() {
		var appID string
		if err := rows.Scan(&appID); err != nil {
			return nil, err
		}
		appIDs = append(appIDs, appID)
	}

	return appIDs, rows.Err()
}

func (s *PostgresUserAppStore) ListAppUsers(ctx context.Context, tenantID, appID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM user_apps
		WHERE tenant_id = $1 AND app_id = $2 AND status = 'active'
		ORDER BY granted_at DESC
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}
