package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for app operations
const (
	appColumns = `id, tenant_id, name, type, status, config, metadata, created_at, updated_at, deleted_at`

	queryAppInsert = `
		INSERT INTO apps (
			id, tenant_id, name, type, status, config, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	queryAppSelect = `SELECT ` + appColumns + ` FROM apps`

	queryAppGet        = queryAppSelect + ` WHERE tenant_id = $1 AND id = $2`
	queryAppGetByName  = queryAppSelect + ` WHERE tenant_id = $1 AND name = $2`
	queryAppList       = queryAppSelect + ` WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	queryAppListByType = queryAppSelect + ` WHERE tenant_id = $1 AND type = $2 AND deleted_at IS NULL ORDER BY created_at DESC`
	queryAppExists     = `SELECT 1 FROM apps WHERE tenant_id = $1 AND id = $2`

	queryAppUpdate = `
		UPDATE apps
		SET name = $1, type = $2, status = $3, config = $4, metadata = $5, updated_at = NOW()
		WHERE tenant_id = $6 AND id = $7`

	queryAppDelete = `UPDATE apps SET deleted_at = NOW(), updated_at = NOW() WHERE tenant_id = $1 AND id = $2`
)

// @Service "pg-app-store"
type pgAppStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.AppStore].
func (p *pgAppStore) Create(ctx context.Context, app *domaincore.App) error {
	config, err := json.Marshal(app.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	metadata, err := json.Marshal(app.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryAppInsert,
		app.ID, app.TenantID, app.Name, app.Type, app.Status,
		config, metadata,
	)
	return err
}

// Get implements [store.AppStore].
func (p *pgAppStore) Get(ctx context.Context, tenantID, appID string) (*domaincore.App, error) {
	app := &domaincore.App{}
	var config, metadata []byte

	err := p.dbPool.QueryRow(ctx, queryAppGet, tenantID, appID).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.Type, &app.Status,
		&config, &metadata, &app.CreatedAt, &app.UpdatedAt, &app.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("app not found: tenant=%s, app=%s", tenantID, appID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalAppFields(app, config, metadata); err != nil {
		return nil, err
	}

	return app, nil
}

// GetByName implements [store.AppStore].
func (p *pgAppStore) GetByName(ctx context.Context, tenantID, name string) (*domaincore.App, error) {
	app := &domaincore.App{}
	var config, metadata []byte

	err := p.dbPool.QueryRow(ctx, queryAppGetByName, tenantID, name).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.Type, &app.Status,
		&config, &metadata, &app.CreatedAt, &app.UpdatedAt, &app.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("app not found with name: %s", name)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalAppFields(app, config, metadata); err != nil {
		return nil, err
	}

	return app, nil
}

// Update implements [store.AppStore].
func (p *pgAppStore) Update(ctx context.Context, app *domaincore.App) error {
	config, err := json.Marshal(app.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	metadata, err := json.Marshal(app.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryAppUpdate,
		app.Name, app.Type, app.Status, config, metadata,
		app.TenantID, app.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("app not found: tenant=%s, app=%s", app.TenantID, app.ID)
	}
	return nil
}

// Delete implements [store.AppStore].
func (p *pgAppStore) Delete(ctx context.Context, tenantID, appID string) error {
	result, err := p.dbPool.Exec(ctx, queryAppDelete, tenantID, appID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("app not found: tenant=%s, app=%s", tenantID, appID)
	}
	return nil
}

// List implements [store.AppStore].
func (p *pgAppStore) List(ctx context.Context, tenantID string) ([]*domaincore.App, error) {
	rows, err := p.dbPool.Query(ctx, queryAppList, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanApps(rows)
}

// ListByType implements [store.AppStore].
func (p *pgAppStore) ListByType(ctx context.Context, tenantID string, appType domaincore.AppType) ([]*domaincore.App, error) {
	rows, err := p.dbPool.Query(ctx, queryAppListByType, tenantID, appType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanApps(rows)
}

// Exists implements [store.AppStore].
func (p *pgAppStore) Exists(ctx context.Context, tenantID, appID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryAppExists, tenantID, appID)
}

var _ store.AppStore = (*pgAppStore)(nil)

// unmarshalAppFields unmarshals JSON fields into app struct
func (p *pgAppStore) unmarshalAppFields(app *domaincore.App, config, metadata []byte) error {
	if len(config) > 0 {
		app.Config = &domaincore.AppConfig{}
		if err := json.Unmarshal(config, app.Config); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		app.Metadata = &m
	}
	return nil
}

func (p *pgAppStore) scanApps(rows serviceapi.Rows) ([]*domaincore.App, error) {
	apps := make([]*domaincore.App, 0, 10)

	for rows.Next() {
		app := &domaincore.App{}
		var config, metadata []byte

		err := rows.Scan(
			&app.ID, &app.TenantID, &app.Name, &app.Type, &app.Status,
			&config, &metadata, &app.CreatedAt, &app.UpdatedAt, &app.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalAppFields(app, config, metadata); err != nil {
			return nil, err
		}

		apps = append(apps, app)
	}

	return apps, rows.Err()
}
