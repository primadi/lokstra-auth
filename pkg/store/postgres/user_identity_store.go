package postgres

import (
	"context"
	"fmt"

	domaincore "github.com/primadi/lokstra-auth/pkg/domain/core"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for user identity operations
const (
	userIdentityColumns = `id, user_id, tenant_id, provider, provider_id, email, username, verified, metadata, created_at, updated_at`

	queryUserIdentityInsert = `
		INSERT INTO user_identities (
			id, user_id, tenant_id, provider, provider_id, email, username, verified, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	queryUserIdentitySelect = `SELECT ` + userIdentityColumns + ` FROM user_identities`

	queryUserIdentityGet           = queryUserIdentitySelect + ` WHERE tenant_id = $1 AND user_id = $2 AND id = $3`
	queryUserIdentityGetByProvider = queryUserIdentitySelect + ` WHERE tenant_id = $1 AND user_id = $2 AND provider = $3`
	queryUserIdentityList          = queryUserIdentitySelect + ` WHERE tenant_id = $1 AND user_id = $2 ORDER BY created_at DESC`
	queryUserIdentityExists        = `SELECT 1 FROM user_identities WHERE tenant_id = $1 AND user_id = $2 AND provider = $3`

	queryUserIdentityFindUser = `
		SELECT u.id, u.tenant_id, u.username, u.email, u.email_verified_at, u.phone_number, u.phone_verified_at, u.full_name, u.password_hash, u.status, u.is_tenant_owner,
		       u.failed_login_attempts, u.last_failed_login_at, u.locked_at, u.locked_until, u.lockout_count, u.metadata, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		INNER JOIN user_identities ui ON u.tenant_id = ui.tenant_id AND u.id = ui.user_id
		WHERE ui.tenant_id = $1 AND ui.provider = $2 AND ui.provider_id = $3`

	queryUserIdentityUpdate = `
		UPDATE user_identities
		SET email = $1, username = $2, verified = $3, metadata = $4, updated_at = NOW()
		WHERE tenant_id = $5 AND user_id = $6 AND id = $7`

	queryUserIdentityDelete = `DELETE FROM user_identities WHERE tenant_id = $1 AND user_id = $2 AND id = $3`
)

// @Service "pg-user-identity-store"
type pgUserIdentityStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.UserIdentityStore].
func (p *pgUserIdentityStore) Create(ctx context.Context, identity *domaincore.UserIdentity) error {
	metadata, err := json.Marshal(identity.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryUserIdentityInsert,
		identity.ID, identity.UserID, identity.TenantID, identity.Provider, identity.ProviderID,
		identity.Email, identity.Username, identity.Verified, metadata,
	)
	return err
}

// Get implements [store.UserIdentityStore].
func (p *pgUserIdentityStore) Get(ctx context.Context, tenantID, userID, identityID string) (*domaincore.UserIdentity, error) {
	identity := &domaincore.UserIdentity{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryUserIdentityGet, tenantID, userID, identityID).Scan(
		&identity.ID, &identity.UserID, &identity.TenantID, &identity.Provider, &identity.ProviderID,
		&identity.Email, &identity.Username, &identity.Verified, &metadata,
		&identity.CreatedAt, &identity.UpdatedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("identity not found: tenant=%s, user=%s, identity=%s", tenantID, userID, identityID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalIdentityFields(identity, metadata); err != nil {
		return nil, err
	}

	return identity, nil
}

// GetByProvider implements [store.UserIdentityStore].
func (p *pgUserIdentityStore) GetByProvider(ctx context.Context, tenantID, userID string, provider domaincore.IdentityProvider) (*domaincore.UserIdentity, error) {
	identity := &domaincore.UserIdentity{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryUserIdentityGetByProvider, tenantID, userID, provider).Scan(
		&identity.ID, &identity.UserID, &identity.TenantID, &identity.Provider, &identity.ProviderID,
		&identity.Email, &identity.Username, &identity.Verified, &metadata,
		&identity.CreatedAt, &identity.UpdatedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("identity not found: tenant=%s, user=%s, provider=%s", tenantID, userID, provider)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalIdentityFields(identity, metadata); err != nil {
		return nil, err
	}

	return identity, nil
}

// Update implements [store.UserIdentityStore].
func (p *pgUserIdentityStore) Update(ctx context.Context, identity *domaincore.UserIdentity) error {
	metadata, err := json.Marshal(identity.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryUserIdentityUpdate,
		identity.Email, identity.Username, identity.Verified, metadata,
		identity.TenantID, identity.UserID, identity.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("identity not found: tenant=%s, user=%s, identity=%s", identity.TenantID, identity.UserID, identity.ID)
	}
	return nil
}

// Delete implements [store.UserIdentityStore].
func (p *pgUserIdentityStore) Delete(ctx context.Context, tenantID, userID, identityID string) error {
	result, err := p.dbPool.Exec(ctx, queryUserIdentityDelete, tenantID, userID, identityID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("identity not found: tenant=%s, user=%s, identity=%s", tenantID, userID, identityID)
	}
	return nil
}

// List implements [store.UserIdentityStore].
func (p *pgUserIdentityStore) List(ctx context.Context, tenantID, userID string) ([]*domaincore.UserIdentity, error) {
	rows, err := p.dbPool.Query(ctx, queryUserIdentityList, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanIdentities(rows)
}

// FindUserByProvider implements [store.UserIdentityStore].
func (p *pgUserIdentityStore) FindUserByProvider(ctx context.Context, tenantID string, provider domaincore.IdentityProvider, providerID string) (*domaincore.User, error) {
	user := &domaincore.User{}
	var metadata []byte

	err := p.dbPool.QueryRow(ctx, queryUserIdentityFindUser, tenantID, provider, providerID).Scan(
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.EmailVerifiedAt, &user.PhoneNumber, &user.PhoneVerifiedAt,
		&user.FullName, &user.PasswordHash, &user.Status, &user.IsTenantOwner,
		&user.FailedLoginAttempts, &user.LastFailedLoginAt, &user.LockedAt, &user.LockedUntil, &user.LockoutCount,
		&metadata, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("user not found for provider: tenant=%s, provider=%s, providerID=%s", tenantID, provider, providerID)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		user.Metadata = &m
	}

	return user, nil
}

// Exists implements [store.UserIdentityStore].
func (p *pgUserIdentityStore) Exists(ctx context.Context, tenantID, userID string, provider domaincore.IdentityProvider) (bool, error) {
	return p.dbPool.IsExists(ctx, queryUserIdentityExists, tenantID, userID, provider)
}

var _ store.UserIdentityStore = (*pgUserIdentityStore)(nil)

// unmarshalIdentityFields unmarshals JSON fields into identity struct
func (p *pgUserIdentityStore) unmarshalIdentityFields(identity *domaincore.UserIdentity, metadata []byte) error {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		identity.Metadata = &m
	}
	return nil
}

func (p *pgUserIdentityStore) scanIdentities(rows serviceapi.Rows) ([]*domaincore.UserIdentity, error) {
	identities := make([]*domaincore.UserIdentity, 0, 10)

	for rows.Next() {
		identity := &domaincore.UserIdentity{}
		var metadata []byte

		err := rows.Scan(
			&identity.ID, &identity.UserID, &identity.TenantID, &identity.Provider, &identity.ProviderID,
			&identity.Email, &identity.Username, &identity.Verified, &metadata,
			&identity.CreatedAt, &identity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalIdentityFields(identity, metadata); err != nil {
			return nil, err
		}

		identities = append(identities, identity)
	}

	return identities, rows.Err()
}
