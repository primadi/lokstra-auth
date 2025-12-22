package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
	"github.com/primadi/lokstra/serviceapi"
)

// PostgresUserIdentityStore is a PostgreSQL implementation of UserIdentityStore
// @Service "postgres-user-identity-store"
type PostgresUserIdentityStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ repository.UserIdentityStore = (*PostgresUserIdentityStore)(nil)

// Create creates a new user identity link
func (s *PostgresUserIdentityStore) Create(ctx context.Context,
	identity *domain.UserIdentity) error {
	if err := identity.Validate(); err != nil {
		return err
	}

	// Serialize metadata
	var metadataJSON []byte
	if identity.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(*identity.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO user_identities (
			id, user_id, tenant_id, provider, provider_id,
			email, username, verified, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.dbPool.Exec(
		ctx,
		query,
		identity.ID,
		identity.UserID,
		identity.TenantID,
		identity.Provider,
		identity.ProviderID,
		identity.Email,
		identity.Username,
		identity.Verified,
		metadataJSON,
		identity.CreatedAt,
		identity.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user identity: %w", err)
	}

	return nil
}

// Get retrieves an identity by ID
func (s *PostgresUserIdentityStore) Get(ctx context.Context,
	tenantID, userID, identityID string) (*domain.UserIdentity, error) {
	query := `
		SELECT 
			id, user_id, tenant_id, provider, provider_id,
			email, username, verified, metadata,
			created_at, updated_at
		FROM user_identities
		WHERE tenant_id = $1 AND user_id = $2 AND id = $3
	`

	identity := &domain.UserIdentity{}
	var metadataJSON []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, userID, identityID).Scan(
		&identity.ID,
		&identity.UserID,
		&identity.TenantID,
		&identity.Provider,
		&identity.ProviderID,
		&identity.Email,
		&identity.Username,
		&identity.Verified,
		&metadataJSON,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, domain.ErrUserIdentityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user identity: %w", err)
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		var metadata map[string]any
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		identity.Metadata = &metadata
	}

	return identity, nil
}

// GetByProvider retrieves an identity by provider for a user
func (s *PostgresUserIdentityStore) GetByProvider(ctx context.Context,
	tenantID, userID string, provider domain.IdentityProvider) (*domain.UserIdentity, error) {
	query := `
		SELECT 
			id, user_id, tenant_id, provider, provider_id,
			email, username, verified, metadata,
			created_at, updated_at
		FROM user_identities
		WHERE tenant_id = $1 AND user_id = $2 AND provider = $3
	`

	identity := &domain.UserIdentity{}
	var metadataJSON []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, userID, provider).Scan(
		&identity.ID,
		&identity.UserID,
		&identity.TenantID,
		&identity.Provider,
		&identity.ProviderID,
		&identity.Email,
		&identity.Username,
		&identity.Verified,
		&metadataJSON,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, domain.ErrUserIdentityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user identity by provider: %w", err)
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		var metadata map[string]any
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		identity.Metadata = &metadata
	}

	return identity, nil
}

// Update updates an existing identity
func (s *PostgresUserIdentityStore) Update(ctx context.Context,
	identity *domain.UserIdentity) error {
	if err := identity.Validate(); err != nil {
		return err
	}

	// Serialize metadata
	var metadataJSON []byte
	if identity.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(*identity.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE user_identities
		SET 
			provider_id = $1,
			email = $2,
			username = $3,
			verified = $4,
			metadata = $5,
			updated_at = $6
		WHERE tenant_id = $7 AND user_id = $8 AND id = $9
	`

	result, err := s.dbPool.Exec(
		ctx,
		query,
		identity.ProviderID,
		identity.Email,
		identity.Username,
		identity.Verified,
		metadataJSON,
		identity.UpdatedAt,
		identity.TenantID,
		identity.UserID,
		identity.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user identity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserIdentityNotFound
	}

	return nil
}

// Delete removes an identity link
func (s *PostgresUserIdentityStore) Delete(ctx context.Context,
	tenantID, userID, identityID string) error {
	query := `
		DELETE FROM user_identities
		WHERE tenant_id = $1 AND user_id = $2 AND id = $3
	`

	result, err := s.dbPool.Exec(ctx, query, tenantID, userID, identityID)
	if err != nil {
		return fmt.Errorf("failed to delete user identity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUserIdentityNotFound
	}

	return nil
}

// List lists all identities for a user
func (s *PostgresUserIdentityStore) List(ctx context.Context,
	tenantID, userID string) ([]*domain.UserIdentity, error) {
	query := `
		SELECT 
			id, user_id, tenant_id, provider, provider_id,
			email, username, verified, metadata,
			created_at, updated_at
		FROM user_identities
		WHERE tenant_id = $1 AND user_id = $2
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user identities: %w", err)
	}
	defer rows.Close()

	var identities []*domain.UserIdentity

	for rows.Next() {
		identity := &domain.UserIdentity{}
		var metadataJSON []byte

		err := rows.Scan(
			&identity.ID,
			&identity.UserID,
			&identity.TenantID,
			&identity.Provider,
			&identity.ProviderID,
			&identity.Email,
			&identity.Username,
			&identity.Verified,
			&metadataJSON,
			&identity.CreatedAt,
			&identity.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user identity: %w", err)
		}

		// Deserialize metadata
		if len(metadataJSON) > 0 {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			identity.Metadata = &metadata
		}

		identities = append(identities, identity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user identities: %w", err)
	}

	return identities, nil
}

// FindUserByProvider finds a user by provider identity (for login)
func (s *PostgresUserIdentityStore) FindUserByProvider(ctx context.Context,
	tenantID string, provider domain.IdentityProvider, providerID string) (*domain.User, error) {
	query := `
		SELECT u.id, u.tenant_id, u.username, u.email, u.full_name, u.password_hash, u.status,
		       u.metadata, u.created_at, u.updated_at, u.deleted_at
		FROM user_identities ui
		JOIN users u ON u.tenant_id = ui.tenant_id AND u.id = ui.user_id
		WHERE ui.tenant_id = $1 AND ui.provider = $2 AND ui.provider_id = $3
	`

	user := &domain.User{}
	var metadataJSON []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID, provider, providerID).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.PasswordHash,
		&user.Status,
		&metadataJSON,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, domain.ErrUserIdentityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by provider: %w", err)
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		var metadata map[string]any
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user metadata: %w", err)
		}
		user.Metadata = &metadata
	}

	return user, nil
}

// Exists checks if an identity exists
func (s *PostgresUserIdentityStore) Exists(ctx context.Context, tenantID, userID string,
	provider domain.IdentityProvider) (bool, error) {
	query := `
		SELECT 1
		FROM user_identities
		WHERE tenant_id = $1 AND user_id = $2 AND provider = $3
	`

	return s.dbPool.IsExists(ctx, query, tenantID, userID, provider)
}
