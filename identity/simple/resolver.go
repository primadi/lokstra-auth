package simple

import (
	"context"
	"fmt"

	identity "github.com/primadi/lokstra-auth/identity"
)

// Resolver is a simple subject resolver that creates subjects from claims
type Resolver struct {
	// SubjectIDClaim is the claim key for subject ID (default: "sub")
	SubjectIDClaim string

	// SubjectTypeClaim is the claim key for subject type (default: "type")
	SubjectTypeClaim string

	// PrincipalClaim is the claim key for principal (default: "username" or "email")
	PrincipalClaim string

	// DefaultSubjectType is the default subject type if not specified in claims
	DefaultSubjectType string
}

// NewResolver creates a new simple subject resolver
func NewResolver() *Resolver {
	return &Resolver{
		SubjectIDClaim:     "sub",
		SubjectTypeClaim:   "type",
		PrincipalClaim:     "username",
		DefaultSubjectType: "user",
	}
}

// Resolve creates a Subject from claims
func (r *Resolver) Resolve(ctx context.Context, claims map[string]any) (*identity.Subject, error) {
	// Extract subject ID
	subID, ok := r.getStringClaim(claims, r.SubjectIDClaim)
	if !ok || subID == "" {
		return nil, fmt.Errorf("missing or invalid subject ID claim: %s", r.SubjectIDClaim)
	}

	// Extract tenant ID (REQUIRED for multi-tenant)
	tenantID, ok := r.getStringClaim(claims, "tenant_id")
	if !ok || tenantID == "" {
		return nil, fmt.Errorf("missing or invalid tenant_id claim")
	}

	// Extract subject type
	subType, ok := r.getStringClaim(claims, r.SubjectTypeClaim)
	if !ok || subType == "" {
		subType = r.DefaultSubjectType
	}

	// Extract principal
	principal, _ := r.getStringClaim(claims, r.PrincipalClaim)
	if principal == "" {
		// Fallback to email if username not found
		principal, _ = r.getStringClaim(claims, "email")
	}
	if principal == "" {
		// Final fallback to subject ID
		principal = subID
	}

	// Extract all other claims as attributes
	attributes := make(map[string]any)
	for k, v := range claims {
		// Skip the claims we already extracted
		if k != r.SubjectIDClaim && k != r.SubjectTypeClaim && k != r.PrincipalClaim && k != "tenant_id" {
			attributes[k] = v
		}
	}

	return &identity.Subject{
		ID:         subID,
		TenantID:   tenantID,
		Type:       subType,
		Principal:  principal,
		Attributes: attributes,
	}, nil
}

func (r *Resolver) getStringClaim(claims map[string]any, key string) (string, bool) {
	val, ok := claims[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}
