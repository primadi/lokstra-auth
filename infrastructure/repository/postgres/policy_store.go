package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/primadi/lokstra-auth/authz/domain"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

// PostgresPolicyStore implements PolicyStore using PostgreSQL
// @Service "postgres-policy-store"
type PostgresPolicyStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

// Create creates a new policy
func (s *PostgresPolicyStore) Create(ctx *request.Context, policy *domain.Policy) error {
	query := `
		INSERT INTO policies (
			id, tenant_id, app_id, name, description, effect, 
			subjects, resources, actions, conditions, status, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	conditionsJSON, err := json.Marshal(policy.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	metadataJSON, err := json.Marshal(policy.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	subjectsJSON, err := json.Marshal(policy.Subjects)
	if err != nil {
		return fmt.Errorf("failed to marshal subjects: %w", err)
	}

	resourcesJSON, err := json.Marshal(policy.Resources)
	if err != nil {
		return fmt.Errorf("failed to marshal resources: %w", err)
	}

	actionsJSON, err := json.Marshal(policy.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(context.Background(), query,
		policy.ID, policy.TenantID, policy.AppID, policy.Name, policy.Description,
		policy.Effect, subjectsJSON, resourcesJSON, actionsJSON,
		conditionsJSON, policy.Status, metadataJSON,
		policy.CreatedAt, policy.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// Get retrieves a policy by ID
func (s *PostgresPolicyStore) Get(ctx *request.Context, tenantID, appID, policyID string) (*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, effect,
		       subjects, resources, actions, conditions, status, metadata,
		       created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3
	`

	var policy domain.Policy
	var conditionsJSON, metadataJSON, subjectsJSON, resourcesJSON, actionsJSON []byte

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	err = cn.QueryRow(ctx, query, tenantID, appID, policyID).Scan(
		&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
		&policy.Effect, &subjectsJSON, &resourcesJSON, &actionsJSON,
		&conditionsJSON, &policy.Status, &metadataJSON,
		&policy.CreatedAt, &policy.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	// Unmarshal JSON fields
	if len(subjectsJSON) > 0 {
		if err := json.Unmarshal(subjectsJSON, &policy.Subjects); err != nil {
			return nil, fmt.Errorf("failed to unmarshal subjects: %w", err)
		}
	}

	if len(resourcesJSON) > 0 {
		if err := json.Unmarshal(resourcesJSON, &policy.Resources); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resources: %w", err)
		}
	}

	if len(actionsJSON) > 0 {
		if err := json.Unmarshal(actionsJSON, &policy.Actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}
	}

	if len(conditionsJSON) > 0 && string(conditionsJSON) != "null" {
		var conditions map[string]any
		if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
		policy.Conditions = &conditions
	}

	if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
		var metadata map[string]any
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		policy.Metadata = &metadata
	}

	return &policy, nil
}

// GetByName retrieves a policy by name within tenant+app scope
func (s *PostgresPolicyStore) GetByName(ctx *request.Context, tenantID, appID, name string) (*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, effect,
		       subjects, resources, actions, conditions, status, metadata,
		       created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND app_id = $2 AND name = $3
	`
	var policy domain.Policy
	var conditionsJSON, metadataJSON, subjectsJSON, resourcesJSON, actionsJSON []byte

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	err = cn.QueryRow(ctx, query, tenantID, appID, name).Scan(
		&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
		&policy.Effect, &subjectsJSON, &resourcesJSON, &actionsJSON,
		&conditionsJSON, &policy.Status, &metadataJSON,
		&policy.CreatedAt, &policy.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get policy by name: %w", err)
	}

	// Unmarshal JSON fields
	if len(subjectsJSON) > 0 {
		if err := json.Unmarshal(subjectsJSON, &policy.Subjects); err != nil {
			return nil, fmt.Errorf("failed to unmarshal subjects: %w", err)
		}
	}

	if len(resourcesJSON) > 0 {
		if err := json.Unmarshal(resourcesJSON, &policy.Resources); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resources: %w", err)
		}
	}

	if len(actionsJSON) > 0 {
		if err := json.Unmarshal(actionsJSON, &policy.Actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}
	}

	if len(conditionsJSON) > 0 && string(conditionsJSON) != "null" {
		var conditions map[string]any
		if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
		policy.Conditions = &conditions
	}

	if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
		var metadata map[string]any
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		policy.Metadata = &metadata
	}

	return &policy, nil
}

// Update updates an existing policy
func (s *PostgresPolicyStore) Update(ctx *request.Context, policy *domain.Policy) error {
	query := `
		UPDATE policies
		SET name = $4, description = $5, effect = $6,
		    subjects = $7, resources = $8, actions = $9,
		    conditions = $10, status = $11, metadata = $12,
		    updated_at = $13
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3
	`

	conditionsJSON, err := json.Marshal(policy.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	metadataJSON, err := json.Marshal(policy.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	subjectsJSON, err := json.Marshal(policy.Subjects)
	if err != nil {
		return fmt.Errorf("failed to marshal subjects: %w", err)
	}

	resourcesJSON, err := json.Marshal(policy.Resources)
	if err != nil {
		return fmt.Errorf("failed to marshal resources: %w", err)
	}

	actionsJSON, err := json.Marshal(policy.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query,
		policy.TenantID, policy.AppID, policy.ID,
		policy.Name, policy.Description, policy.Effect,
		subjectsJSON, resourcesJSON, actionsJSON,
		conditionsJSON, policy.Status, metadataJSON,
		policy.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}

// Delete deletes a policy (hard delete for policies)
func (s *PostgresPolicyStore) Delete(ctx *request.Context, tenantID, appID, policyID string) error {
	query := `
		DELETE FROM policies
		WHERE tenant_id = $1 AND app_id = $2 AND id = $3
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer cn.Release()

	_, err = cn.Exec(ctx, query, tenantID, appID, policyID)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}

// List lists all policies for tenant+app
func (s *PostgresPolicyStore) List(ctx *request.Context, tenantID, appID string) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, effect,
		       subjects, resources, actions, conditions, status, metadata,
		       created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND app_id = $2
		ORDER BY created_at DESC
	`
	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		var conditionsJSON, metadataJSON, subjectsJSON, resourcesJSON, actionsJSON []byte

		err := rows.Scan(
			&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
			&policy.Effect, &subjectsJSON, &resourcesJSON, &actionsJSON,
			&conditionsJSON, &policy.Status, &metadataJSON,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		// Unmarshal JSON fields
		if len(subjectsJSON) > 0 {
			if err := json.Unmarshal(subjectsJSON, &policy.Subjects); err != nil {
				return nil, fmt.Errorf("failed to unmarshal subjects: %w", err)
			}
		}

		if len(resourcesJSON) > 0 {
			if err := json.Unmarshal(resourcesJSON, &policy.Resources); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resources: %w", err)
			}
		}

		if len(actionsJSON) > 0 {
			if err := json.Unmarshal(actionsJSON, &policy.Actions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
			}
		}

		if len(conditionsJSON) > 0 && string(conditionsJSON) != "null" {
			var conditions map[string]any
			if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
			policy.Conditions = &conditions
		}

		if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			policy.Metadata = &metadata
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}

// ListWithFilters lists policies with filters
func (s *PostgresPolicyStore) ListWithFilters(ctx *request.Context, req *domain.ListPoliciesRequest) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, effect,
		       subjects, resources, actions, conditions, status, metadata,
		       created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND app_id = $2
	`
	args := []any{req.TenantID, req.AppID}
	argPos := 3

	if req.Effect != nil {
		query += fmt.Sprintf(" AND effect = $%d", argPos)
		args = append(args, *req.Effect)
		argPos++
	}

	if req.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *req.Status)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, req.Limit)
		argPos++
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, req.Offset)
	}

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies with filters: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		var conditionsJSON, metadataJSON, subjectsJSON, resourcesJSON, actionsJSON []byte

		err := rows.Scan(
			&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
			&policy.Effect, &subjectsJSON, &resourcesJSON, &actionsJSON,
			&conditionsJSON, &policy.Status, &metadataJSON,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		// Unmarshal JSON fields
		if len(subjectsJSON) > 0 {
			if err := json.Unmarshal(subjectsJSON, &policy.Subjects); err != nil {
				return nil, fmt.Errorf("failed to unmarshal subjects: %w", err)
			}
		}

		if len(resourcesJSON) > 0 {
			if err := json.Unmarshal(resourcesJSON, &policy.Resources); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resources: %w", err)
			}
		}

		if len(actionsJSON) > 0 {
			if err := json.Unmarshal(actionsJSON, &policy.Actions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
			}
		}

		if len(conditionsJSON) > 0 && string(conditionsJSON) != "null" {
			var conditions map[string]any
			if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
			policy.Conditions = &conditions
		}

		if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			policy.Metadata = &metadata
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}

// FindBySubject finds policies that match a subject
func (s *PostgresPolicyStore) FindBySubject(ctx *request.Context, tenantID, appID, subjectID string) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, effect,
		       subjects, resources, actions, conditions, status, metadata,
		       created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND app_id = $2 
		  AND status = 'active'
		  AND (subjects @> $3 OR subjects @> '["*"]')
		ORDER BY created_at DESC
	`

	subjectJSON := fmt.Sprintf(`["%s"]`, subjectID)

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, tenantID, appID, subjectJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to find policies by subject: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		var conditionsJSON, metadataJSON, subjectsJSON, resourcesJSON, actionsJSON []byte

		err := rows.Scan(
			&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
			&policy.Effect, &subjectsJSON, &resourcesJSON, &actionsJSON,
			&conditionsJSON, &policy.Status, &metadataJSON,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		// Unmarshal JSON fields
		if len(subjectsJSON) > 0 {
			if err := json.Unmarshal(subjectsJSON, &policy.Subjects); err != nil {
				return nil, fmt.Errorf("failed to unmarshal subjects: %w", err)
			}
		}

		if len(resourcesJSON) > 0 {
			if err := json.Unmarshal(resourcesJSON, &policy.Resources); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resources: %w", err)
			}
		}

		if len(actionsJSON) > 0 {
			if err := json.Unmarshal(actionsJSON, &policy.Actions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
			}
		}

		if len(conditionsJSON) > 0 && string(conditionsJSON) != "null" {
			var conditions map[string]any
			if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
			policy.Conditions = &conditions
		}

		if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			policy.Metadata = &metadata
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}

// FindByResource finds policies that match a resource pattern
func (s *PostgresPolicyStore) FindByResource(ctx *request.Context, tenantID, appID, resourceType, resourceID string) ([]*domain.Policy, error) {
	query := `
		SELECT id, tenant_id, app_id, name, description, effect,
		       subjects, resources, actions, conditions, status, metadata,
		       created_at, updated_at
		FROM policies
		WHERE tenant_id = $1 AND app_id = $2 
		  AND status = 'active'
	`
	args := []any{tenantID, appID}

	// Build resource pattern matching
	if resourceID != "" {
		resourcePattern := fmt.Sprintf(`["%s:%s"]`, resourceType, resourceID)
		query += ` AND (resources @> $3 OR resources @> '["*"]')`
		args = append(args, resourcePattern)
	} else {
		resourcePattern := fmt.Sprintf(`["%s:*"]`, resourceType)
		query += ` AND (resources @> $3 OR resources @> '["*"]')`
		args = append(args, resourcePattern)
	}

	query += " ORDER BY created_at DESC"

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer cn.Release()

	rows, err := cn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find policies by resource: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var policy domain.Policy
		var conditionsJSON, metadataJSON, subjectsJSON, resourcesJSON, actionsJSON []byte

		err := rows.Scan(
			&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
			&policy.Effect, &subjectsJSON, &resourcesJSON, &actionsJSON,
			&conditionsJSON, &policy.Status, &metadataJSON,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		// Unmarshal JSON fields
		if len(subjectsJSON) > 0 {
			if err := json.Unmarshal(subjectsJSON, &policy.Subjects); err != nil {
				return nil, fmt.Errorf("failed to unmarshal subjects: %w", err)
			}
		}

		if len(resourcesJSON) > 0 {
			if err := json.Unmarshal(resourcesJSON, &policy.Resources); err != nil {
				return nil, fmt.Errorf("failed to unmarshal resources: %w", err)
			}
		}

		if len(actionsJSON) > 0 {
			if err := json.Unmarshal(actionsJSON, &policy.Actions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
			}
		}

		if len(conditionsJSON) > 0 && string(conditionsJSON) != "null" {
			var conditions map[string]any
			if err := json.Unmarshal(conditionsJSON, &conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
			policy.Conditions = &conditions
		}

		if len(metadataJSON) > 0 && string(metadataJSON) != "null" {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			policy.Metadata = &metadata
		}

		policies = append(policies, &policy)
	}

	return policies, nil
}

// Exists checks if a policy exists
func (s *PostgresPolicyStore) Exists(ctx *request.Context, tenantID, appID, policyID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM policies
			WHERE tenant_id = $1 AND app_id = $2 AND id = $3
		)
	`

	cn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer cn.Release()

	var exists bool
	err = cn.QueryRow(ctx, query, tenantID, appID, policyID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check policy existence: %w", err)
	}

	return exists, nil
}
