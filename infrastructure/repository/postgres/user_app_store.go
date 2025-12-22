package postgres

import (
	"context"
	"time"

	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-user-app-store"
type PostgresUserAppStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.UserAppStore = (*PostgresUserAppStore)(nil)

func (s *PostgresUserAppStore) GrantAccess(ctx context.Context, tenantID, appID, userID string) error {
	query := `
		INSERT INTO user_apps (tenant_id, app_id, user_id, status, granted_at)
		VALUES ($1, $2, $3, 'active', $4)
		ON CONFLICT (tenant_id, app_id, user_id)
		DO UPDATE SET status = 'active', granted_at = $4, revoked_at = NULL
	`

	now := time.Now()
	_, err := s.dbPool.Exec(ctx, query, tenantID, appID, userID, now)

	return err
}

func (s *PostgresUserAppStore) RevokeAccess(ctx context.Context, tenantID, appID, userID string) error {
	query := `
		UPDATE user_apps
		SET status = 'revoked', revoked_at = $1
		WHERE tenant_id = $2 AND app_id = $3 AND user_id = $4
	`

	now := time.Now()

	_, err := s.dbPool.Exec(ctx, query, now, tenantID, appID, userID)
	return err
}

func (s *PostgresUserAppStore) HasAccess(ctx context.Context, tenantID, appID, userID string) (bool, error) {
	query := `
		SELECT 1 FROM user_apps
			WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3 AND status = 'active'
	`

	return s.dbPool.IsExists(ctx, query, tenantID, appID, userID)
}

func (s *PostgresUserAppStore) ListUserApps(ctx context.Context, tenantID, userID string) ([]string, error) {
	query := `
		SELECT app_id
		FROM user_apps
		WHERE tenant_id = $1 AND user_id = $2 AND status = 'active'
		ORDER BY granted_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, userID)
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

	rows, err := s.dbPool.Query(ctx, query, tenantID, appID)
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
