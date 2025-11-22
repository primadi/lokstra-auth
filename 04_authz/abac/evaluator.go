package abac

import (
	"context"
	"fmt"
	"strings"

	subject "github.com/primadi/lokstra-auth/03_subject"
	authz "github.com/primadi/lokstra-auth/04_authz"
)

// Evaluator is an ABAC (Attribute-Based Access Control) policy evaluator
type Evaluator struct {
	rules             []*Rule
	attributeProvider authz.AttributeProvider
	defaultDecision   bool
}

// Rule represents an ABAC rule (tenant and app scoped)
type Rule struct {
	ID          string
	TenantID    string // Tenant this rule belongs to (required)
	AppID       string // App this rule belongs to (optional, empty = all apps in tenant)
	Description string
	Effect      string // "allow" or "deny"
	Conditions  []Condition
	Priority    int // Higher priority rules evaluated first
}

// Condition represents a condition in an ABAC rule
type Condition struct {
	Type     string // "subject", "resource", "environment", "action"
	Key      string
	Operator string // "eq", "ne", "in", "not_in", "gt", "lt", "contains", "matches"
	Value    any
}

// NewEvaluator creates a new ABAC evaluator
func NewEvaluator(attributeProvider authz.AttributeProvider, defaultDecision bool) *Evaluator {
	return &Evaluator{
		rules:             make([]*Rule, 0),
		attributeProvider: attributeProvider,
		defaultDecision:   defaultDecision,
	}
}

// AddRule adds a rule to the evaluator
func (e *Evaluator) AddRule(rule *Rule) {
	e.rules = append(e.rules, rule)
	// Sort by priority (higher first)
	e.sortRules()
}

// sortRules sorts rules by priority
func (e *Evaluator) sortRules() {
	// Simple bubble sort (fine for small rule sets)
	for i := 0; i < len(e.rules); i++ {
		for j := i + 1; j < len(e.rules); j++ {
			if e.rules[j].Priority > e.rules[i].Priority {
				e.rules[i], e.rules[j] = e.rules[j], e.rules[i]
			}
		}
	}
}

// Evaluate evaluates ABAC policies for an authorization request
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

	// Collect attributes
	subjectAttrs := e.getSubjectAttributes(request.Subject)

	// Add tenant and app to subject attributes for rule evaluation
	subjectAttrs["tenant_id"] = tenantID
	subjectAttrs["app_id"] = appID

	resourceAttrs := request.Resource.Attributes
	if resourceAttrs == nil {
		resourceAttrs = make(map[string]any)
	}

	// Add resource type, ID, tenant, and app to attributes
	resourceAttrs["type"] = request.Resource.Type
	resourceAttrs["id"] = request.Resource.ID
	resourceAttrs["tenant_id"] = request.Resource.TenantID
	resourceAttrs["app_id"] = request.Resource.AppID

	envAttrs := request.Context
	if envAttrs == nil {
		envAttrs = make(map[string]any)
	}

	// Evaluate rules in priority order (only rules for this tenant+app)
	for _, rule := range e.rules {
		// Skip rules for other tenants
		if rule.TenantID != tenantID {
			continue
		}

		// Skip rules for other apps (unless rule applies to all apps)
		if rule.AppID != "" && rule.AppID != appID {
			continue
		}

		matches, err := e.evaluateRule(rule, request.Action, subjectAttrs, resourceAttrs, envAttrs)
		if err != nil {
			return nil, err
		}

		if matches {
			allowed := rule.Effect == "allow"
			return &authz.AuthorizationDecision{
				Allowed: allowed,
				Reason:  fmt.Sprintf("rule '%s' matched (%s)", rule.ID, rule.Effect),
				Metadata: map[string]any{
					"rule_id": rule.ID,
					"effect":  rule.Effect,
				},
			}, nil
		}
	}

	// No rules matched, return default decision
	return &authz.AuthorizationDecision{
		Allowed: e.defaultDecision,
		Reason:  "no matching rules, using default decision",
	}, nil
}

// evaluateRule checks if all conditions in a rule match
func (e *Evaluator) evaluateRule(rule *Rule, action authz.Action, subjectAttrs, resourceAttrs, envAttrs map[string]any) (bool, error) {
	for _, condition := range rule.Conditions {
		matches, err := e.evaluateCondition(condition, action, subjectAttrs, resourceAttrs, envAttrs)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}
	return true, nil
}

// evaluateCondition evaluates a single condition
func (e *Evaluator) evaluateCondition(condition Condition, action authz.Action, subjectAttrs, resourceAttrs, envAttrs map[string]any) (bool, error) {
	var actualValue any

	switch condition.Type {
	case "action":
		actualValue = string(action)
	case "subject":
		actualValue = subjectAttrs[condition.Key]
	case "resource":
		actualValue = resourceAttrs[condition.Key]
	case "environment":
		actualValue = envAttrs[condition.Key]
	default:
		return false, fmt.Errorf("unknown condition type: %s", condition.Type)
	}

	return e.compareValues(actualValue, condition.Operator, condition.Value)
}

// compareValues compares two values using an operator
func (e *Evaluator) compareValues(actual any, operator string, expected any) (bool, error) {
	switch operator {
	case "eq":
		return actual == expected, nil

	case "ne":
		return actual != expected, nil

	case "in":
		// Check if actual is in expected (slice)
		expectedSlice, ok := expected.([]any)
		if !ok {
			return false, fmt.Errorf("'in' operator requires slice value")
		}
		for _, v := range expectedSlice {
			if actual == v {
				return true, nil
			}
		}
		return false, nil

	case "not_in":
		// Check if actual is not in expected (slice)
		expectedSlice, ok := expected.([]any)
		if !ok {
			return false, fmt.Errorf("'not_in' operator requires slice value")
		}
		for _, v := range expectedSlice {
			if actual == v {
				return false, nil
			}
		}
		return true, nil

	case "contains":
		// Check if actual (string) contains expected (string)
		actualStr, ok1 := actual.(string)
		expectedStr, ok2 := expected.(string)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("'contains' operator requires string values")
		}
		return strings.Contains(actualStr, expectedStr), nil

	case "gt":
		// Greater than (numbers)
		actualNum, ok1 := toFloat64(actual)
		expectedNum, ok2 := toFloat64(expected)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("'gt' operator requires numeric values")
		}
		return actualNum > expectedNum, nil

	case "lt":
		// Less than (numbers)
		actualNum, ok1 := toFloat64(actual)
		expectedNum, ok2 := toFloat64(expected)
		if !ok1 || !ok2 {
			return false, fmt.Errorf("'lt' operator requires numeric values")
		}
		return actualNum < expectedNum, nil

	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// toFloat64 converts a value to float64
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

// getSubjectAttributes extracts attributes from identity context (includes tenant+app)
func (e *Evaluator) getSubjectAttributes(identity *subject.IdentityContext) map[string]any {
	attrs := make(map[string]any)

	// Add tenant and app context
	attrs["tenant_id"] = identity.TenantID
	attrs["app_id"] = identity.AppID

	if identity.Subject != nil {
		attrs["id"] = identity.Subject.ID
		attrs["type"] = identity.Subject.Type
		attrs["principal"] = identity.Subject.Principal
		attrs["subject_tenant_id"] = identity.Subject.TenantID

		// Add all subject attributes
		for k, v := range identity.Subject.Attributes {
			attrs[k] = v
		}
	}

	// Add metadata (commonly used for attributes in examples)
	if identity.Metadata != nil {
		for k, v := range identity.Metadata {
			attrs[k] = v
		}
	}

	// Add roles
	attrs["roles"] = identity.Roles

	// Add groups
	attrs["groups"] = identity.Groups

	// Add profile data
	if identity.Profile != nil {
		for k, v := range identity.Profile {
			attrs[k] = v
		}
	}

	return attrs
}

// HasPermission checks if the subject has a specific permission (not applicable for pure ABAC)
func (e *Evaluator) HasPermission(ctx context.Context, identity *subject.IdentityContext, permission string) (bool, error) {
	// ABAC doesn't use traditional permissions, but we can check using a dummy request
	return false, fmt.Errorf("HasPermission not applicable for pure ABAC, use Evaluate instead")
}

// HasAnyPermission checks if the subject has any of the specified permissions
func (e *Evaluator) HasAnyPermission(ctx context.Context, identity *subject.IdentityContext, permissions ...string) (bool, error) {
	return false, fmt.Errorf("HasAnyPermission not applicable for pure ABAC, use Evaluate instead")
}

// HasAllPermissions checks if the subject has all of the specified permissions
func (e *Evaluator) HasAllPermissions(ctx context.Context, identity *subject.IdentityContext, permissions ...string) (bool, error) {
	return false, fmt.Errorf("HasAllPermissions not applicable for pure ABAC, use Evaluate instead")
}

// HasRole checks if the subject has a specific role
func (e *Evaluator) HasRole(ctx context.Context, identity *subject.IdentityContext, role string) (bool, error) {
	return identity.HasRole(role), nil
}

// HasAnyRole checks if the subject has any of the specified roles
func (e *Evaluator) HasAnyRole(ctx context.Context, identity *subject.IdentityContext, roles ...string) (bool, error) {
	return identity.HasAnyRole(roles...), nil
}

// HasAllRoles checks if the subject has all of the specified roles
func (e *Evaluator) HasAllRoles(ctx context.Context, identity *subject.IdentityContext, roles ...string) (bool, error) {
	return identity.HasAllRoles(roles...), nil
}

// RemoveRule removes a rule by ID
func (e *Evaluator) RemoveRule(ruleID string) bool {
	for i, rule := range e.rules {
		if rule.ID == ruleID {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			return true
		}
	}
	return false
}

// GetRules returns all rules
func (e *Evaluator) GetRules() []*Rule {
	result := make([]*Rule, len(e.rules))
	copy(result, e.rules)
	return result
}
