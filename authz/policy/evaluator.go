package policy

import (
	"context"
	"fmt"
	"strings"

	authz "github.com/primadi/lokstra-auth/authz"
	identity "github.com/primadi/lokstra-auth/identity"
)

// Evaluator is a policy-based authorization evaluator
type Evaluator struct {
	store            authz.PolicyStore
	combineAlgorithm string // "deny-overrides", "allow-overrides", "first-applicable"
}

// NewEvaluator creates a new policy evaluator
func NewEvaluator(store authz.PolicyStore, combineAlgorithm string) *Evaluator {
	if combineAlgorithm == "" {
		combineAlgorithm = "deny-overrides" // Default: deny wins
	}

	return &Evaluator{
		store:            store,
		combineAlgorithm: combineAlgorithm,
	}
}

// Evaluate evaluates policies for an authorization request
func (e *Evaluator) Evaluate(ctx context.Context, request *authz.AuthorizationRequest) (*authz.AuthorizationDecision, error) {
	// Extract tenant and app from identity context
	tenantID := request.Subject.TenantID
	appID := request.Subject.AppID

	// Validate tenant+app match resource
	if request.Resource.TenantID != "" && request.Resource.TenantID != tenantID {
		return &authz.AuthorizationDecision{
			Allowed: false,
			Reason:  "resource tenant mismatch",
		}, nil
	}
	if request.Resource.AppID != "" && request.Resource.AppID != appID {
		return &authz.AuthorizationDecision{
			Allowed: false,
			Reason:  "resource app mismatch",
		}, nil
	}

	// Find applicable policies (tenant+app scoped)
	policies, err := e.findApplicablePolicies(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(policies) == 0 {
		return &authz.AuthorizationDecision{
			Allowed: false,
			Reason:  "no applicable policies found",
		}, nil
	}

	// Combine policy decisions based on algorithm
	return e.combinePolicies(policies, request), nil
}

// findApplicablePolicies finds policies applicable to the request (tenant+app scoped)
func (e *Evaluator) findApplicablePolicies(ctx context.Context, request *authz.AuthorizationRequest) ([]*authz.Policy, error) {
	tenantID := request.Subject.TenantID
	appID := request.Subject.AppID
	subjectID := request.Subject.Subject.ID

	// Get policies for subject (tenant+app scoped)
	subjectPolicies, err := e.store.FindBySubject(ctx, tenantID, appID, subjectID)
	if err != nil {
		return nil, err
	}

	// Get policies for resource (tenant+app scoped)
	resourcePolicies, err := e.store.FindByResource(ctx, tenantID, appID, request.Resource.Type, request.Resource.ID)
	if err != nil {
		return nil, err
	}

	// Combine and filter applicable policies (only for this tenant+app)
	policyMap := make(map[string]*authz.Policy)

	for _, policy := range subjectPolicies {
		// Extra check: ensure policy belongs to correct tenant+app
		if policy.TenantID == tenantID && (policy.AppID == "" || policy.AppID == appID) {
			if e.policyApplies(policy, request) {
				policyMap[policy.ID] = policy
			}
		}
	}

	for _, policy := range resourcePolicies {
		// Extra check: ensure policy belongs to correct tenant+app
		if policy.TenantID == tenantID && (policy.AppID == "" || policy.AppID == appID) {
			if e.policyApplies(policy, request) {
				policyMap[policy.ID] = policy
			}
		}
	}

	// Convert map to slice
	applicable := make([]*authz.Policy, 0, len(policyMap))
	for _, policy := range policyMap {
		applicable = append(applicable, policy)
	}

	return applicable, nil
}

// policyApplies checks if a policy applies to the request
func (e *Evaluator) policyApplies(policy *authz.Policy, request *authz.AuthorizationRequest) bool {
	// Check subject
	subjectMatches := false
	for _, sub := range policy.Subjects {
		if sub == "*" || sub == request.Subject.Subject.ID {
			subjectMatches = true
			break
		}
		// Check if subject has matching role
		for _, role := range request.Subject.Roles {
			if sub == fmt.Sprintf("role:%s", role) {
				subjectMatches = true
				break
			}
		}
	}

	if !subjectMatches {
		return false
	}

	// Check resource
	resourceMatches := false
	resourcePattern := fmt.Sprintf("%s:%s", request.Resource.Type, request.Resource.ID)
	for _, res := range policy.Resources {
		if res == "*" || res == resourcePattern || res == fmt.Sprintf("%s:*", request.Resource.Type) {
			resourceMatches = true
			break
		}
	}

	if !resourceMatches {
		return false
	}

	// Check action
	actionMatches := false
	for _, action := range policy.Actions {
		if action == "*" || action == request.Action {
			actionMatches = true
			break
		}
	}

	if !actionMatches {
		return false
	}

	// Check conditions (if any)
	if len(policy.Conditions) > 0 {
		return e.evaluateConditions(policy.Conditions, request)
	}

	return true
}

// evaluateConditions evaluates policy conditions
func (e *Evaluator) evaluateConditions(conditions map[string]any, request *authz.AuthorizationRequest) bool {
	// Simple condition evaluation
	for key, expectedValue := range conditions {
		// Check in request context
		if actualValue, ok := request.Context[key]; ok {
			if actualValue != expectedValue {
				return false
			}
		} else {
			// Check in resource attributes
			if request.Resource.Attributes != nil {
				if actualValue, ok := request.Resource.Attributes[key]; ok {
					if actualValue != expectedValue {
						return false
					}
				} else {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}

// combinePolicies combines multiple policy decisions
func (e *Evaluator) combinePolicies(policies []*authz.Policy, _ *authz.AuthorizationRequest) *authz.AuthorizationDecision {
	switch e.combineAlgorithm {
	case "deny-overrides":
		return e.denyOverrides(policies)
	case "allow-overrides":
		return e.allowOverrides(policies)
	case "first-applicable":
		return e.firstApplicable(policies)
	default:
		return e.denyOverrides(policies)
	}
}

// denyOverrides: if any policy denies, result is deny
func (e *Evaluator) denyOverrides(policies []*authz.Policy) *authz.AuthorizationDecision {
	hasAllow := false
	var allowPolicy *authz.Policy

	for _, policy := range policies {
		if strings.ToLower(policy.Effect) == "deny" {
			return &authz.AuthorizationDecision{
				Allowed: false,
				Reason:  fmt.Sprintf("policy '%s' denies access", policy.ID),
				Metadata: map[string]any{
					"policy_id": policy.ID,
					"algorithm": "deny-overrides",
				},
			}
		}
		if strings.ToLower(policy.Effect) == "allow" {
			hasAllow = true
			allowPolicy = policy
		}
	}

	if hasAllow {
		return &authz.AuthorizationDecision{
			Allowed: true,
			Reason:  fmt.Sprintf("policy '%s' allows access", allowPolicy.ID),
			Metadata: map[string]any{
				"policy_id": allowPolicy.ID,
				"algorithm": "deny-overrides",
			},
		}
	}

	return &authz.AuthorizationDecision{
		Allowed: false,
		Reason:  "no explicit allow policy found",
	}
}

// allowOverrides: if any policy allows, result is allow
func (e *Evaluator) allowOverrides(policies []*authz.Policy) *authz.AuthorizationDecision {
	hasDeny := false
	var denyPolicy *authz.Policy

	for _, policy := range policies {
		if strings.ToLower(policy.Effect) == "allow" {
			return &authz.AuthorizationDecision{
				Allowed: true,
				Reason:  fmt.Sprintf("policy '%s' allows access", policy.ID),
				Metadata: map[string]any{
					"policy_id": policy.ID,
					"algorithm": "allow-overrides",
				},
			}
		}
		if strings.ToLower(policy.Effect) == "deny" {
			hasDeny = true
			denyPolicy = policy
		}
	}

	if hasDeny {
		return &authz.AuthorizationDecision{
			Allowed: false,
			Reason:  fmt.Sprintf("policy '%s' denies access", denyPolicy.ID),
			Metadata: map[string]any{
				"policy_id": denyPolicy.ID,
				"algorithm": "allow-overrides",
			},
		}
	}

	return &authz.AuthorizationDecision{
		Allowed: false,
		Reason:  "no applicable policies found",
	}
}

// firstApplicable: first policy that applies wins
func (e *Evaluator) firstApplicable(policies []*authz.Policy) *authz.AuthorizationDecision {
	if len(policies) == 0 {
		return &authz.AuthorizationDecision{
			Allowed: false,
			Reason:  "no applicable policies found",
		}
	}

	policy := policies[0]
	allowed := strings.ToLower(policy.Effect) == "allow"

	return &authz.AuthorizationDecision{
		Allowed: allowed,
		Reason:  fmt.Sprintf("first applicable policy '%s' %s access", policy.ID, policy.Effect),
		Metadata: map[string]any{
			"policy_id": policy.ID,
			"algorithm": "first-applicable",
		},
	}
}

// HasPermission checks if the subject has a specific permission
func (e *Evaluator) HasPermission(ctx context.Context, identity *identity.IdentityContext, permission string) (bool, error) {
	// Check if identity has the permission directly
	return identity.HasPermission(permission), nil
}

// HasAnyPermission checks if the subject has any of the specified permissions
func (e *Evaluator) HasAnyPermission(ctx context.Context, identity *identity.IdentityContext, permissions ...string) (bool, error) {
	for _, perm := range permissions {
		if identity.HasPermission(perm) {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if the subject has all of the specified permissions
func (e *Evaluator) HasAllPermissions(ctx context.Context, identity *identity.IdentityContext, permissions ...string) (bool, error) {
	for _, perm := range permissions {
		if !identity.HasPermission(perm) {
			return false, nil
		}
	}
	return true, nil
}

// HasRole checks if the subject has a specific role
func (e *Evaluator) HasRole(ctx context.Context, identity *identity.IdentityContext, role string) (bool, error) {
	return identity.HasRole(role), nil
}

// HasAnyRole checks if the subject has any of the specified roles
func (e *Evaluator) HasAnyRole(ctx context.Context, identity *identity.IdentityContext, roles ...string) (bool, error) {
	return identity.HasAnyRole(roles...), nil
}

// HasAllRoles checks if the subject has all of the specified roles
func (e *Evaluator) HasAllRoles(ctx context.Context, identity *identity.IdentityContext, roles ...string) (bool, error) {
	return identity.HasAllRoles(roles...), nil
}
