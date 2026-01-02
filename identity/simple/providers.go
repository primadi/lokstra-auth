package simple

import (
	"context"
	"sync"

	identity "github.com/primadi/lokstra-auth/identity"
)

// InMemoryRoleProvider provides roles from in-memory storage with tenant+app isolation
type InMemoryRoleProvider struct {
	mu    sync.RWMutex
	roles map[string][]string // key: "tenantID:appID:subjectID"
}

// NewInMemoryRoleProvider creates a new in-memory role provider
func NewInMemoryRoleProvider() *InMemoryRoleProvider {
	return &InMemoryRoleProvider{
		roles: make(map[string][]string),
	}
}

// GetRoles retrieves roles for a subject in a specific tenant and app
func (p *InMemoryRoleProvider) GetRoles(ctx context.Context, tenantID, appID string, sub *identity.Subject) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	key := tenantID + ":" + appID + ":" + sub.ID
	if roles, ok := p.roles[key]; ok {
		return roles, nil
	}
	return []string{}, nil
}

// SetRoles sets roles for a subject in a specific tenant and app
func (p *InMemoryRoleProvider) SetRoles(tenantID, appID, subjectID string, roles []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + appID + ":" + subjectID
	p.roles[key] = roles
}

// AddRole adds a role to a subject in a specific tenant and app
func (p *InMemoryRoleProvider) AddRole(tenantID, appID, subjectID, role string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + appID + ":" + subjectID
	if _, ok := p.roles[key]; !ok {
		p.roles[key] = []string{}
	}
	p.roles[key] = append(p.roles[key], role)
}

// RemoveRole removes a role from a subject in a specific tenant and app
func (p *InMemoryRoleProvider) RemoveRole(tenantID, appID, subjectID, role string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + appID + ":" + subjectID
	if roles, ok := p.roles[key]; ok {
		newRoles := make([]string, 0, len(roles))
		for _, r := range roles {
			if r != role {
				newRoles = append(newRoles, r)
			}
		}
		p.roles[key] = newRoles
	}
}

// InMemoryPermissionProvider provides permissions from in-memory storage with tenant+app isolation
type InMemoryPermissionProvider struct {
	mu          sync.RWMutex
	permissions map[string][]string // key: "tenantID:appID:subjectID"
}

// NewInMemoryPermissionProvider creates a new in-memory permission provider
func NewInMemoryPermissionProvider() *InMemoryPermissionProvider {
	return &InMemoryPermissionProvider{
		permissions: make(map[string][]string),
	}
}

// GetPermissions retrieves permissions for a subject in a specific tenant and app
func (p *InMemoryPermissionProvider) GetPermissions(ctx context.Context, tenantID, appID string, sub *identity.Subject) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	key := tenantID + ":" + appID + ":" + sub.ID
	if permissions, ok := p.permissions[key]; ok {
		return permissions, nil
	}
	return []string{}, nil
}

// SetPermissions sets permissions for a subject in a specific tenant and app
func (p *InMemoryPermissionProvider) SetPermissions(tenantID, appID, subjectID string, permissions []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + appID + ":" + subjectID
	p.permissions[key] = permissions
}

// AddPermission adds a permission to a subject in a specific tenant and app
func (p *InMemoryPermissionProvider) AddPermission(tenantID, appID, subjectID, permission string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + appID + ":" + subjectID
	if _, ok := p.permissions[key]; !ok {
		p.permissions[key] = []string{}
	}
	p.permissions[key] = append(p.permissions[key], permission)
}

// RemovePermission removes a permission from a subject in a specific tenant and app
func (p *InMemoryPermissionProvider) RemovePermission(tenantID, appID, subjectID, permission string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + appID + ":" + subjectID
	if permissions, ok := p.permissions[key]; ok {
		newPermissions := make([]string, 0, len(permissions))
		for _, p := range permissions {
			if p != permission {
				newPermissions = append(newPermissions, p)
			}
		}
		p.permissions[key] = newPermissions
	}
}

// InMemoryGroupProvider provides groups from in-memory storage with tenant isolation
type InMemoryGroupProvider struct {
	mu     sync.RWMutex
	groups map[string][]string // key: "tenantID:subjectID"
}

// NewInMemoryGroupProvider creates a new in-memory group provider
func NewInMemoryGroupProvider() *InMemoryGroupProvider {
	return &InMemoryGroupProvider{
		groups: make(map[string][]string),
	}
}

// GetGroups retrieves groups for a subject in a specific tenant
func (p *InMemoryGroupProvider) GetGroups(ctx context.Context, tenantID string, sub *identity.Subject) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	key := tenantID + ":" + sub.ID
	if groups, ok := p.groups[key]; ok {
		return groups, nil
	}
	return []string{}, nil
}

// SetGroups sets groups for a subject in a specific tenant
func (p *InMemoryGroupProvider) SetGroups(tenantID, subjectID string, groups []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + subjectID
	p.groups[key] = groups
}

// AddGroup adds a group to a subject in a specific tenant
func (p *InMemoryGroupProvider) AddGroup(tenantID, subjectID, group string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + subjectID
	if _, ok := p.groups[key]; !ok {
		p.groups[key] = []string{}
	}
	p.groups[key] = append(p.groups[key], group)
}

// RemoveGroup removes a group from a subject in a specific tenant
func (p *InMemoryGroupProvider) RemoveGroup(tenantID, subjectID, group string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + subjectID
	if groups, ok := p.groups[key]; ok {
		newGroups := make([]string, 0, len(groups))
		for _, g := range groups {
			if g != group {
				newGroups = append(newGroups, g)
			}
		}
		p.groups[key] = newGroups
	}
}

// InMemoryProfileProvider provides profiles from in-memory storage with tenant isolation
type InMemoryProfileProvider struct {
	mu       sync.RWMutex
	profiles map[string]map[string]any // key: "tenantID:subjectID"
}

// NewInMemoryProfileProvider creates a new in-memory profile provider
func NewInMemoryProfileProvider() *InMemoryProfileProvider {
	return &InMemoryProfileProvider{
		profiles: make(map[string]map[string]any),
	}
}

// GetProfile retrieves profile information for a subject in a specific tenant
func (p *InMemoryProfileProvider) GetProfile(ctx context.Context, tenantID string, sub *identity.Subject) (map[string]any, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	key := tenantID + ":" + sub.ID
	if profile, ok := p.profiles[key]; ok {
		return profile, nil
	}
	return make(map[string]any), nil
}

// SetProfile sets profile information for a subject in a specific tenant
func (p *InMemoryProfileProvider) SetProfile(tenantID, subjectID string, profile map[string]any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + subjectID
	p.profiles[key] = profile
}

// UpdateProfile updates profile information for a subject in a specific tenant
func (p *InMemoryProfileProvider) UpdateProfile(tenantID, subjectID string, updates map[string]any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := tenantID + ":" + subjectID
	if _, ok := p.profiles[key]; !ok {
		p.profiles[key] = make(map[string]any)
	}
	for k, v := range updates {
		p.profiles[key][k] = v
	}
}
