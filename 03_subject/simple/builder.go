package simple

import (
	"context"

	subject "github.com/primadi/lokstra-auth/03_subject"
)

// ContextBuilder is a simple identity context builder
type ContextBuilder struct {
	roleProvider       subject.RoleProvider
	permissionProvider subject.PermissionProvider
	groupProvider      subject.GroupProvider
	profileProvider    subject.ProfileProvider
}

// NewContextBuilder creates a new simple identity context builder
func NewContextBuilder(
	roleProvider subject.RoleProvider,
	permissionProvider subject.PermissionProvider,
	groupProvider subject.GroupProvider,
	profileProvider subject.ProfileProvider,
) *ContextBuilder {
	return &ContextBuilder{
		roleProvider:       roleProvider,
		permissionProvider: permissionProvider,
		groupProvider:      groupProvider,
		profileProvider:    profileProvider,
	}
}

// Build creates an IdentityContext from a subject
func (b *ContextBuilder) Build(ctx context.Context, sub *subject.Subject) (*subject.IdentityContext, error) {
	identity := &subject.IdentityContext{
		Subject:  sub,
		Metadata: make(map[string]any),
	}

	// Load roles
	if b.roleProvider != nil {
		roles, err := b.roleProvider.GetRoles(ctx, sub)
		if err != nil {
			return nil, err
		}
		identity.Roles = roles
	}

	// Load permissions
	if b.permissionProvider != nil {
		permissions, err := b.permissionProvider.GetPermissions(ctx, sub)
		if err != nil {
			return nil, err
		}
		identity.Permissions = permissions
	}

	// Load groups
	if b.groupProvider != nil {
		groups, err := b.groupProvider.GetGroups(ctx, sub)
		if err != nil {
			return nil, err
		}
		identity.Groups = groups
	}

	// Load profile
	if b.profileProvider != nil {
		profile, err := b.profileProvider.GetProfile(ctx, sub)
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

// GetRoles retrieves roles for a subject
func (p *StaticRoleProvider) GetRoles(ctx context.Context, sub *subject.Subject) ([]string, error) {
	if roles, ok := p.roles[sub.ID]; ok {
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

// GetPermissions retrieves permissions for a subject
func (p *StaticPermissionProvider) GetPermissions(ctx context.Context, sub *subject.Subject) ([]string, error) {
	if permissions, ok := p.permissions[sub.ID]; ok {
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

// GetGroups retrieves groups for a subject
func (p *StaticGroupProvider) GetGroups(ctx context.Context, sub *subject.Subject) ([]string, error) {
	if groups, ok := p.groups[sub.ID]; ok {
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

// GetProfile retrieves profile information for a subject
func (p *StaticProfileProvider) GetProfile(ctx context.Context, sub *subject.Subject) (map[string]any, error) {
	if profile, ok := p.profiles[sub.ID]; ok {
		return profile, nil
	}
	return make(map[string]any), nil
}
