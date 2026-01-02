package simple

import (
	"context"
	"fmt"

	identity "github.com/primadi/lokstra-auth/identity"
)

// ContextBuilder is a simple identity context builder
type ContextBuilder struct {
	roleProvider       identity.RoleProvider
	permissionProvider identity.PermissionProvider
	groupProvider      identity.GroupProvider
	profileProvider    identity.ProfileProvider
}

// NewContextBuilder creates a new simple identity context builder
func NewContextBuilder(
	roleProvider identity.RoleProvider,
	permissionProvider identity.PermissionProvider,
	groupProvider identity.GroupProvider,
	profileProvider identity.ProfileProvider,
) *ContextBuilder {
	return &ContextBuilder{
		roleProvider:       roleProvider,
		permissionProvider: permissionProvider,
		groupProvider:      groupProvider,
		profileProvider:    profileProvider,
	}
}

// Build creates an IdentityContext from a subject
func (b *ContextBuilder) Build(ctx context.Context, sub *identity.Subject) (*identity.IdentityContext, error) {
	// Extract app_id from subject attributes (set by token claims)
	appID, _ := sub.Attributes["app_id"].(string)
	if appID == "" {
		return nil, fmt.Errorf("missing app_id in subject attributes")
	}

	identity := &identity.IdentityContext{
		Subject:  sub,
		TenantID: sub.TenantID,
		AppID:    appID,
		Metadata: make(map[string]any),
	}

	// Load roles (scoped to tenant+app)
	if b.roleProvider != nil {
		roles, err := b.roleProvider.GetRoles(ctx, sub.TenantID, appID, sub)
		if err != nil {
			return nil, err
		}
		identity.Roles = roles
	}

	// Load permissions (scoped to tenant+app)
	if b.permissionProvider != nil {
		permissions, err := b.permissionProvider.GetPermissions(ctx, sub.TenantID, appID, sub)
		if err != nil {
			return nil, err
		}
		identity.Permissions = permissions
	}

	// Load groups (scoped to tenant)
	if b.groupProvider != nil {
		groups, err := b.groupProvider.GetGroups(ctx, sub.TenantID, sub)
		if err != nil {
			return nil, err
		}
		identity.Groups = groups
	}

	// Load profile (scoped to tenant)
	if b.profileProvider != nil {
		profile, err := b.profileProvider.GetProfile(ctx, sub.TenantID, sub)
		if err != nil {
			return nil, err
		}
		identity.Profile = profile
	}

	return identity, nil
}

// StaticRoleProvider provides a static list of roles
type StaticRoleProvider struct {
	roles map[string][]string
}

// NewStaticRoleProvider creates a new static role provider
func NewStaticRoleProvider(roles map[string][]string) *StaticRoleProvider {
	return &StaticRoleProvider{
		roles: roles,
	}
}

// GetRoles retrieves roles for a subject (with tenant and app scoping)
func (p *StaticRoleProvider) GetRoles(ctx context.Context, tenantID, appID string, sub *identity.Subject) ([]string, error) {
	// Use composite key for tenant+app+user isolation
	key := tenantID + ":" + appID + ":" + sub.ID
	if roles, ok := p.roles[key]; ok {
		return roles, nil
	}
	return []string{}, nil
}

// StaticPermissionProvider provides a static list of permissions
type StaticPermissionProvider struct {
	permissions map[string][]string
}

// NewStaticPermissionProvider creates a new static permission provider
func NewStaticPermissionProvider(permissions map[string][]string) *StaticPermissionProvider {
	return &StaticPermissionProvider{
		permissions: permissions,
	}
}

// GetPermissions retrieves permissions for a subject (with tenant and app scoping)
func (p *StaticPermissionProvider) GetPermissions(ctx context.Context, tenantID, appID string, sub *identity.Subject) ([]string, error) {
	// Use composite key for tenant+app+user isolation
	key := tenantID + ":" + appID + ":" + sub.ID
	if permissions, ok := p.permissions[key]; ok {
		return permissions, nil
	}
	return []string{}, nil
}

// StaticGroupProvider provides a static list of groups
type StaticGroupProvider struct {
	groups map[string][]string
}

// NewStaticGroupProvider creates a new static group provider
func NewStaticGroupProvider(groups map[string][]string) *StaticGroupProvider {
	return &StaticGroupProvider{
		groups: groups,
	}
}

// GetGroups retrieves groups for a subject (with tenant scoping)
func (p *StaticGroupProvider) GetGroups(ctx context.Context, tenantID string, sub *identity.Subject) ([]string, error) {
	// Use composite key for tenant+user isolation
	key := tenantID + ":" + sub.ID
	if groups, ok := p.groups[key]; ok {
		return groups, nil
	}
	return []string{}, nil
}

// StaticProfileProvider provides a static profile
type StaticProfileProvider struct {
	profiles map[string]map[string]any
}

// NewStaticProfileProvider creates a new static profile provider
func NewStaticProfileProvider(profiles map[string]map[string]any) *StaticProfileProvider {
	return &StaticProfileProvider{
		profiles: profiles,
	}
}

// GetProfile retrieves profile information for a subject (with tenant scoping)
func (p *StaticProfileProvider) GetProfile(ctx context.Context, tenantID string, sub *identity.Subject) (map[string]any, error) {
	// Use composite key for tenant+user isolation
	key := tenantID + ":" + sub.ID
	if profile, ok := p.profiles[key]; ok {
		return profile, nil
	}
	return make(map[string]any), nil
}
