package postgres

import (
	"fmt"

	authzdomain "github.com/primadi/lokstra-auth/pkg/domain/authz"
	"github.com/primadi/lokstra-auth/pkg/store"
	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

// Query constants for policy operations
const (
	policyColumns = `id, tenant_id, app_id, name, description, effect, subjects, resources, actions, conditions, status, metadata, created_at, updated_at`

	queryPolicyInsert = `
		INSERT INTO policies (
			id, tenant_id, app_id, name, description, effect, subjects, resources, actions, conditions, status, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	queryPolicySelect = `SELECT ` + policyColumns + ` FROM policies`

	queryPolicyGet       = queryPolicySelect + ` WHERE tenant_id = $1 AND app_id = $2 AND id = $3`
	queryPolicyGetByName = queryPolicySelect + ` WHERE tenant_id = $1 AND app_id = $2 AND name = $3`
	queryPolicyList      = queryPolicySelect + ` WHERE tenant_id = $1 AND app_id = $2 ORDER BY created_at DESC`

	queryPolicyFindBySubject  = queryPolicySelect + ` WHERE tenant_id = $1 AND app_id = $2 AND subjects @> $3::jsonb ORDER BY created_at DESC`
	queryPolicyFindByResource = queryPolicySelect + ` WHERE tenant_id = $1 AND app_id = $2 AND (resources @> $3::jsonb OR resources @> '["*"]'::jsonb) ORDER BY created_at DESC`

	queryPolicyUpdate = `
		UPDATE policies
		SET name = $1, description = $2, effect = $3, subjects = $4, resources = $5, actions = $6, conditions = $7, status = $8, metadata = $9, updated_at = NOW()
		WHERE tenant_id = $10 AND app_id = $11 AND id = $12`

	queryPolicyDelete = `DELETE FROM policies WHERE tenant_id = $1 AND app_id = $2 AND id = $3`

	queryPolicyExists = `SELECT 1 FROM policies WHERE tenant_id = $1 AND app_id = $2 AND id = $3`
)

// @Service "pg-policy-store"
type pgPolicyStore struct {
	// @Inject "@auth.db_auth"
	dbPool serviceapi.DbPool
}

// Create implements [store.PolicyStore].
func (p *pgPolicyStore) Create(ctx *request.Context, policy *authzdomain.Policy) error {
	subjects, err := json.Marshal(policy.Subjects)
	if err != nil {
		return fmt.Errorf("marshal subjects: %w", err)
	}
	resources, err := json.Marshal(policy.Resources)
	if err != nil {
		return fmt.Errorf("marshal resources: %w", err)
	}
	actions, err := json.Marshal(policy.Actions)
	if err != nil {
		return fmt.Errorf("marshal actions: %w", err)
	}
	conditions, err := json.Marshal(policy.Conditions)
	if err != nil {
		return fmt.Errorf("marshal conditions: %w", err)
	}
	metadata, err := json.Marshal(policy.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.dbPool.Exec(ctx, queryPolicyInsert,
		policy.ID, policy.TenantID, policy.AppID, policy.Name, policy.Description,
		policy.Effect, subjects, resources, actions, conditions, policy.Status, metadata,
	)
	return err
}

// Get implements [store.PolicyStore].
func (p *pgPolicyStore) Get(ctx *request.Context, tenantID, appID, policyID string) (*authzdomain.Policy, error) {
	policy := &authzdomain.Policy{}
	var subjects, resources, actions, conditions, metadata []byte

	err := p.dbPool.QueryRow(ctx, queryPolicyGet, tenantID, appID, policyID).Scan(
		&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
		&policy.Effect, &subjects, &resources, &actions, &conditions, &policy.Status, &metadata,
		&policy.CreatedAt, &policy.UpdatedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("policy not found: tenant=%s, app=%s, policy=%s", tenantID, appID, policyID)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalPolicyFields(policy, subjects, resources, actions, conditions, metadata); err != nil {
		return nil, err
	}

	return policy, nil
}

// GetByName implements [store.PolicyStore].
func (p *pgPolicyStore) GetByName(ctx *request.Context, tenantID, appID, name string) (*authzdomain.Policy, error) {
	policy := &authzdomain.Policy{}
	var subjects, resources, actions, conditions, metadata []byte

	err := p.dbPool.QueryRow(ctx, queryPolicyGetByName, tenantID, appID, name).Scan(
		&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
		&policy.Effect, &subjects, &resources, &actions, &conditions, &policy.Status, &metadata,
		&policy.CreatedAt, &policy.UpdatedAt,
	)

	if p.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("policy not found: tenant=%s, app=%s, name=%s", tenantID, appID, name)
	}
	if err != nil {
		return nil, err
	}

	if err := p.unmarshalPolicyFields(policy, subjects, resources, actions, conditions, metadata); err != nil {
		return nil, err
	}

	return policy, nil
}

// Update implements [store.PolicyStore].
func (p *pgPolicyStore) Update(ctx *request.Context, policy *authzdomain.Policy) error {
	subjects, err := json.Marshal(policy.Subjects)
	if err != nil {
		return fmt.Errorf("marshal subjects: %w", err)
	}
	resources, err := json.Marshal(policy.Resources)
	if err != nil {
		return fmt.Errorf("marshal resources: %w", err)
	}
	actions, err := json.Marshal(policy.Actions)
	if err != nil {
		return fmt.Errorf("marshal actions: %w", err)
	}
	conditions, err := json.Marshal(policy.Conditions)
	if err != nil {
		return fmt.Errorf("marshal conditions: %w", err)
	}
	metadata, err := json.Marshal(policy.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	result, err := p.dbPool.Exec(ctx, queryPolicyUpdate,
		policy.Name, policy.Description, policy.Effect, subjects, resources, actions,
		conditions, policy.Status, metadata, policy.TenantID, policy.AppID, policy.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("policy not found: tenant=%s, app=%s, policy=%s", policy.TenantID, policy.AppID, policy.ID)
	}
	return nil
}

// Delete implements [store.PolicyStore].
func (p *pgPolicyStore) Delete(ctx *request.Context, tenantID, appID, policyID string) error {
	result, err := p.dbPool.Exec(ctx, queryPolicyDelete, tenantID, appID, policyID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("policy not found: tenant=%s, app=%s, policy=%s", tenantID, appID, policyID)
	}
	return nil
}

// List implements [store.PolicyStore].
func (p *pgPolicyStore) List(ctx *request.Context, tenantID, appID string) ([]*authzdomain.Policy, error) {
	rows, err := p.dbPool.Query(ctx, queryPolicyList, tenantID, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanPolicies(rows)
}

// FindBySubject implements [store.PolicyStore].
func (p *pgPolicyStore) FindBySubject(ctx *request.Context, tenantID, appID, subjectID string) ([]*authzdomain.Policy, error) {
	subjectJSON := fmt.Sprintf(`["%s"]`, subjectID)
	rows, err := p.dbPool.Query(ctx, queryPolicyFindBySubject, tenantID, appID, subjectJSON)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanPolicies(rows)
}

// FindByResource implements [store.PolicyStore].
func (p *pgPolicyStore) FindByResource(ctx *request.Context, tenantID, appID, resourceType, resourceID string) ([]*authzdomain.Policy, error) {
	resourcePattern := fmt.Sprintf(`["%s:%s"]`, resourceType, resourceID)
	rows, err := p.dbPool.Query(ctx, queryPolicyFindByResource, tenantID, appID, resourcePattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanPolicies(rows)
}

// Exists implements [store.PolicyStore].
func (p *pgPolicyStore) Exists(ctx *request.Context, tenantID, appID, policyID string) (bool, error) {
	return p.dbPool.IsExists(ctx, queryPolicyExists, tenantID, appID, policyID)
}

var _ store.PolicyStore = (*pgPolicyStore)(nil)

// unmarshalPolicyFields unmarshals JSON fields into policy struct
func (p *pgPolicyStore) unmarshalPolicyFields(policy *authzdomain.Policy, subjects, resources, actions, conditions, metadata []byte) error {
	if len(subjects) > 0 {
		if err := json.Unmarshal(subjects, &policy.Subjects); err != nil {
			return fmt.Errorf("failed to unmarshal subjects: %w", err)
		}
	}
	if len(resources) > 0 {
		if err := json.Unmarshal(resources, &policy.Resources); err != nil {
			return fmt.Errorf("failed to unmarshal resources: %w", err)
		}
	}
	if len(actions) > 0 {
		if err := json.Unmarshal(actions, &policy.Actions); err != nil {
			return fmt.Errorf("failed to unmarshal actions: %w", err)
		}
	}
	if len(conditions) > 0 {
		var c map[string]any
		if err := json.Unmarshal(conditions, &c); err != nil {
			return fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
		policy.Conditions = &c
	}
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		policy.Metadata = &m
	}
	return nil
}

func (p *pgPolicyStore) scanPolicies(rows serviceapi.Rows) ([]*authzdomain.Policy, error) {
	policies := make([]*authzdomain.Policy, 0, 10)

	for rows.Next() {
		policy := &authzdomain.Policy{}
		var subjects, resources, actions, conditions, metadata []byte

		err := rows.Scan(
			&policy.ID, &policy.TenantID, &policy.AppID, &policy.Name, &policy.Description,
			&policy.Effect, &subjects, &resources, &actions, &conditions, &policy.Status, &metadata,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := p.unmarshalPolicyFields(policy, subjects, resources, actions, conditions, metadata); err != nil {
			return nil, err
		}

		policies = append(policies, policy)
	}

	return policies, rows.Err()
}
