package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for user_app operations
const (
	userAppColumns = `tenant_id, app_id, user_id, status, granted_at, granted_by, revoked_at, revoked_by, metadata, created_at, updated_at`

	queryUserAppInsert = `
		INSERT INTO user_apps (
			tenant_id, app_id, user_id, status, granted_by, metadata
		) VALUES ($1, $2, $3, $4, $5, $6)`

	queryUserAppSelect = `SELECT ` + userAppColumns + ` FROM user_apps`

	queryUserAppGet        = queryUserAppSelect + ` WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3`
	queryUserAppListByUser = queryUserAppSelect + ` WHERE tenant_id = $1 AND user_id = $2 ORDER BY granted_at DESC`
	queryUserAppListByApp  = queryUserAppSelect + ` WHERE tenant_id = $1 AND app_id = $2 ORDER BY granted_at DESC`
	queryUserAppExists     = `SELECT 1 FROM user_apps WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3`

	queryUserAppUpdate = `
		UPDATE user_apps
		SET status = $1, metadata = $2, updated_at = NOW()
		WHERE tenant_id = $3 AND app_id = $4 AND user_id = $5`

	queryUserAppRevoke = `
		UPDATE user_apps
		SET status = 'revoked', revoked_at = NOW(), revoked_by = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND app_id = $3 AND user_id = $4`

	queryUserAppDelete = `DELETE FROM user_apps WHERE tenant_id = $1 AND app_id = $2 AND user_id = $3`
)

// @Service "pg-user-app-store"
type pgUserAppStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.UserAppStore].
func (p *pgUserAppStore) Create(ctx context.Context, userApp *domaincore.UserApp) error {
	metadata, err := json.Marshal(userApp.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryUserAppInsert,
		userApp.TenantID, userApp.AppID, userApp.UserID, userApp.Status,
		userApp.GrantedBy, metadata,
	)
	return err
}

// Get implements [store.UserAppStore].
func (p *pgUserAppStore) Get(ctx context.Context, tenantID, appID, userID string) (*domaincore.UserApp, error) {
	userApp := &domaincore.UserApp{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryUserAppGet, tenantID, appID, userID).Scan(
		&userApp.TenantID, &userApp.AppID, &userApp.UserID, &userApp.Status,
		&userApp.GrantedAt, &userApp.GrantedBy, &userApp.RevokedAt, &userApp.RevokedBy,
		&metadata, &userApp.CreatedAt, &userApp.UpdatedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user app not found: tenant=%s, app=%s, user=%s", tenantID, appID, userID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalUserAppFields(userApp, metadata); err != nil {
		return nil, err
	}

	return userApp, nil
}

// Update implements [store.UserAppStore].
func (p *pgUserAppStore) Update(ctx context.Context, userApp *domaincore.UserApp) error {
	metadata, err := json.Marshal(userApp.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryUserAppUpdate,
		userApp.Status, metadata,
		userApp.TenantID, userApp.AppID, userApp.UserID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user app not found: tenant=%s, app=%s, user=%s", userApp.TenantID, userApp.AppID, userApp.UserID)
	}
	return nil
}

// Delete implements [store.UserAppStore].
func (p *pgUserAppStore) Delete(ctx context.Context, tenantID, appID, userID string) error {
	result, err := p.dbPool.Exec(ctx, queryUserAppDelete, tenantID, appID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user app not found: tenant=%s, app=%s, user=%s", tenantID, appID, userID)
	}
	return nil
}

// ListByUser implements [store.UserAppStore].
func (p *pgUserAppStore) ListByUser(ctx context.Context, tenantID, userID string) ([]*domaincore.UserApp, error) {
	rows, err := p.dbPool.Query(ctx, queryUserAppListByUser, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanUserApps(rows)
}

// ListByApp implements [store.UserAppStore].
func (p *pgUserAppStore) ListByApp(ctx context.Context, tenantID, appID string) ([]*domaincore.UserApp, error) {
	rows, err := p.dbPool.Query(ctx, queryUserAppListByApp, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanUserApps(rows)
}

// GrantAccess implements [store.UserAppStore].
func (p *pgUserAppStore) GrantAccess(ctx context.Context, tenantID, appID, userID string) error {
	userApp := &domaincore.UserApp{
		TenantID: tenantID,
		AppID:    appID,
		UserID:   userID,
		Status:   domaincore.UserAppStatusActive,
	}
	return p.Create(ctx, userApp)
}

// RevokeAccess implements [store.UserAppStore].
func (p *pgUserAppStore) RevokeAccess(ctx context.Context, tenantID, appID, userID string) error {
	result, err := p.dbPool.Exec(ctx, queryUserAppRevoke, nil, tenantID, appID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user app not found: tenant=%s, app=%s, user=%s", tenantID, appID, userID)
	}
	return nil
}

// HasAccess implements [store.UserAppStore].
func (p *pgUserAppStore) HasAccess(ctx context.Context, tenantID, appID, userID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryUserAppExists, tenantID, appID, userID)
}

// ListUserApps implements [store.UserAppStore].
func (p *pgUserAppStore) ListUserApps(ctx context.Context, tenantID, userID string) ([]string, error) {
	rows, err := p.dbPool.Query(ctx, `SELECT app_id FROM user_apps WHERE tenant_id = $1 AND user_id = $2 AND status = 'active'`, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appIDs := make([]string, 0, 10)
	for rows.Next() {
		var appID string
		if err := rows.Scan(&appID); err != nil {
			return nil, err
		}
		appIDs = append(appIDs, appID)
	}

	return appIDs, rows.Err()
}

// ListAppUsers implements [store.UserAppStore].
func (p *pgUserAppStore) ListAppUsers(ctx context.Context, tenantID, appID string) ([]string, error) {
	rows, err := p.dbPool.Query(ctx, `SELECT user_id FROM user_apps WHERE tenant_id = $1 AND app_id = $2 AND status = 'active'`, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]string, 0, 10)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

var _ store.UserAppStore = (*pgUserAppStore)(nil)

// unmarshalUserAppFields unmarshals JSON fields into userApp struct
func (p *pgUserAppStore) unmarshalUserAppFields(userApp *domaincore.UserApp, metadata []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		userApp.Metadata = &m
	}
	return nil
}

func (p *pgUserAppStore) scanUserApps(rows serviceapi.Rows) ([]*domaincore.UserApp, error) {
	userApps := make([]*domaincore.UserApp, 0, 10)

	for rows.Next() {
		userApp := &domaincore.UserApp{}
		var metadata []byte

		err := rows.Scan(
			&userApp.TenantID, &userApp.AppID, &userApp.UserID, &userApp.Status,
			&userApp.GrantedAt, &userApp.GrantedBy, &userApp.RevokedAt, &userApp.RevokedBy,
			&metadata, &userApp.CreatedAt, &userApp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalUserAppFields(userApp, metadata); err != nil {
			return nil, err
		}

		userApps = append(userApps, userApp)
	}

	return userApps, rows.Err()
}
