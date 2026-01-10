package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for user operations
const (
	userColumns = `id, tenant_id, username, email, email_verified_at, phone_number, phone_verified_at, full_name, password_hash, status, is_tenant_owner, 
	               failed_login_attempts, last_failed_login_at, locked_at, locked_until, lockout_count, metadata, created_at, updated_at, deleted_at`

	queryUserInsert = `
		INSERT INTO users (
			id, tenant_id, username, email, email_verified_at, phone_number, phone_verified_at, full_name, password_hash, status, is_tenant_owner,
			failed_login_attempts, last_failed_login_at, locked_at, locked_until, lockout_count, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	queryUserSelect = `SELECT ` + userColumns + ` FROM users`

	queryUserGet           = queryUserSelect + ` WHERE tenant_id = $1 AND id = $2`
	queryUserGetByUsername = queryUserSelect + ` WHERE tenant_id = $1 AND username = $2`
	queryUserGetByEmail    = queryUserSelect + ` WHERE tenant_id = $1 AND email = $2`
	queryUserList          = queryUserSelect + ` WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	queryUserListByApp     = `
		SELECT ` + userColumns + `
		FROM users u
		INNER JOIN user_apps ua ON u.tenant_id = ua.tenant_id AND u.id = ua.user_id
		WHERE u.tenant_id = $1 AND ua.app_id = $2 AND ua.status = 'active' AND u.deleted_at IS NULL
		ORDER BY u.created_at DESC`
	queryUserExists = `SELECT 1 FROM users WHERE tenant_id = $1 AND id = $2`

	queryUserUpdate = `
		UPDATE users
		SET username = $1, email = $2, email_verified_at = $3, phone_number = $4, phone_verified_at = $5, 
		    full_name = $6, status = $7, is_tenant_owner = $8, metadata = $9, updated_at = NOW()
		WHERE tenant_id = $10 AND id = $11`

	queryUserSetPassword    = `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE tenant_id = $2 AND id = $3`
	queryUserRemovePassword = `UPDATE users SET password_hash = NULL, updated_at = NOW() WHERE tenant_id = $1 AND id = $2`
	queryUserDelete         = `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE tenant_id = $1 AND id = $2`
)

// @Service "pg-user-store"
type pgUserStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.UserStore].
func (p *pgUserStore) Create(ctx context.Context, user *domaincore.User) error {
	metadata, err := json.Marshal(user.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryUserInsert,
		user.ID, user.TenantID, user.Username, user.Email, user.EmailVerifiedAt, user.PhoneNumber, user.PhoneVerifiedAt,
		user.FullName, user.PasswordHash, user.Status, user.IsTenantOwner,
		user.FailedLoginAttempts, user.LastFailedLoginAt, user.LockedAt, user.LockedUntil, user.LockoutCount,
		metadata,
	)
	return err
}

// Get implements [store.UserStore].
func (p *pgUserStore) Get(ctx context.Context, tenantID, userID string) (*domaincore.User, error) {
	user := &domaincore.User{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryUserGet, tenantID, userID).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.PhoneNumber, &user.PhoneVerifiedAt,
		&user.FullName, &user.PasswordHash, &user.Status, &user.IsTenantOwner,
		&user.FailedLoginAttempts, &user.LastFailedLoginAt, &user.LockedAt, &user.LockedUntil, &user.LockoutCount,
		&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user not found: tenant=%s, user=%s", tenantID, userID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalUserFields(user, metadata); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByUsername implements [store.UserStore].
func (p *pgUserStore) GetByUsername(ctx context.Context, tenantID, username string) (*domaincore.User, error) {
	user := &domaincore.User{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryUserGetByUsername, tenantID, username).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.PhoneNumber, &user.PhoneVerifiedAt,
		&user.FullName, &user.PasswordHash, &user.Status, &user.IsTenantOwner,
		&user.FailedLoginAttempts, &user.LastFailedLoginAt, &user.LockedAt, &user.LockedUntil, &user.LockoutCount,
		&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user not found with username: %s", username)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalUserFields(user, metadata); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail implements [store.UserStore].
func (p *pgUserStore) GetByEmail(ctx context.Context, tenantID, email string) (*domaincore.User, error) {
	user := &domaincore.User{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryUserGetByEmail, tenantID, email).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.PhoneNumber, &user.PhoneVerifiedAt,
		&user.FullName, &user.PasswordHash, &user.Status, &user.IsTenantOwner,
		&user.FailedLoginAttempts, &user.LastFailedLoginAt, &user.LockedAt, &user.LockedUntil, &user.LockoutCount,
		&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user not found with email: %s", email)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalUserFields(user, metadata); err != nil {
		return nil, err
	}

	return user, nil
}

// Update implements [store.UserStore].
func (p *pgUserStore) Update(ctx context.Context, user *domaincore.User) error {
	metadata, err := json.Marshal(user.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryUserUpdate,
		user.Username, user.Email, user.EmailVerifiedAt, user.PhoneNumber, user.PhoneVerifiedAt,
		user.FullName, user.Status, user.IsTenantOwner, metadata,
		user.TenantID, user.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: tenant=%s, user=%s", user.TenantID, user.ID)
	}
	return nil
}

// Delete implements [store.UserStore].
func (p *pgUserStore) Delete(ctx context.Context, tenantID, userID string) error {
	result, err := p.dbPool.Exec(ctx, queryUserDelete, tenantID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: tenant=%s, user=%s", tenantID, userID)
	}
	return nil
}

// List implements [store.UserStore].
func (p *pgUserStore) List(ctx context.Context, tenantID string) ([]*domaincore.User, error) {
	rows, err := p.dbPool.Query(ctx, queryUserList, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanUsers(rows)
}

// ListByApp implements [store.UserStore].
func (p *pgUserStore) ListByApp(ctx context.Context, tenantID, appID string) ([]*domaincore.User, error) {
	rows, err := p.dbPool.Query(ctx, queryUserListByApp, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanUsers(rows)
}

// SetPassword implements [store.UserStore].
func (p *pgUserStore) SetPassword(ctx context.Context, tenantID, userID, passwordHash string) error {
	result, err := p.dbPool.Exec(ctx, queryUserSetPassword, passwordHash, tenantID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: tenant=%s, user=%s", tenantID, userID)
	}
	return nil
}

// RemovePassword implements [store.UserStore].
func (p *pgUserStore) RemovePassword(ctx context.Context, tenantID, userID string) error {
	result, err := p.dbPool.Exec(ctx, queryUserRemovePassword, tenantID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: tenant=%s, user=%s", tenantID, userID)
	}
	return nil
}

// Exists implements [store.UserStore].
func (p *pgUserStore) Exists(ctx context.Context, tenantID, userID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryUserExists, tenantID, userID)
}

var _ store.UserStore = (*pgUserStore)(nil)

// unmarshalUserFields unmarshals JSON fields into user struct
func (p *pgUserStore) unmarshalUserFields(user *domaincore.User, metadata []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		user.Metadata = &m
	}
	return nil
}

func (p *pgUserStore) scanUsers(rows serviceapi.Rows) ([]*domaincore.User, error) {
	users := make([]*domaincore.User, 0, 10)

	for rows.Next() {
		user := &domaincore.User{}
		var metadata []byte

		err := rows.Scan(
			&user.ID, &user.TenantID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.PhoneNumber, &user.PhoneVerifiedAt,
			&user.FullName, &user.PasswordHash, &user.Status, &user.IsTenantOwner,
			&user.FailedLoginAttempts, &user.LastFailedLoginAt, &user.LockedAt, &user.LockedUntil, &user.LockoutCount,
			&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalUserFields(user, metadata); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, rows.Err()
}
