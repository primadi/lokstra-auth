package acl

import (
	"context"
	"fmt"
	"strings"
	"sync"

	subject "github.com/primadi/lokstra-auth/03_subject"
	authz "github.com/primadi/lokstra-auth/04_authz"
)

// ACLEntry represents a single access control entry
type ACLEntry struct {
	SubjectID   string   // User or role ID
	SubjectType string   // "user" or "role"
	Permissions []string // List of allowed permissions
}

// Manager manages access control lists for resources (multi-tenant aware)
type Manager struct {
	// acls: map[tenantID:appID:resourceType:resourceID] -> ACL entries
	// Composite key ensures tenant+app isolation
	acls map[string][]*ACLEntry
	mu   sync.RWMutex
}

// NewManager creates a new ACL manager
func NewManager() *Manager {
	return &Manager{
		acls: make(map[string][]*ACLEntry),
	}
}

// resourceKey creates a unique key for a resource (with tenant+app scoping)
func (m *Manager) resourceKey(tenantID, appID, resourceType, resourceID string) string {
	return fmt.Sprintf("%s:%s:%s:%s", tenantID, appID, strings.ToLower(resourceType), resourceID)
}

// Grant grants access to a resource (implements AccessControlList interface)
func (m *Manager) Grant(ctx context.Context, tenantID, appID, subjectID string, resource *authz.Resource, action authz.Action) error {
	return m.grantPermissions(ctx, tenantID, appID, resource.Type, resource.ID, subjectID, "user", string(action))
}

// Revoke revokes access to a resource (implements AccessControlList interface)
func (m *Manager) Revoke(ctx context.Context, tenantID, appID, subjectID string, resource *authz.Resource, action authz.Action) error {
	return m.revokePermissions(ctx, tenantID, appID, resource.Type, resource.ID, subjectID, "user", string(action))
}

// Check checks if a subject has access to a resource (implements AccessControlList interface)
func (m *Manager) Check(ctx context.Context, tenantID, appID, subjectID string, resource *authz.Resource, action authz.Action) (bool, error) {
	return m.checkPermission(ctx, tenantID, appID, resource.Type, resource.ID, subjectID, string(action), nil)
}

// List lists all permissions for a subject on a resource (implements AccessControlList interface)
func (m *Manager) List(ctx context.Context, tenantID, appID, subjectID string, resource *authz.Resource) ([]authz.Action, error) {
	permissions, err := m.GetPermissions(ctx, tenantID, appID, resource.Type, resource.ID, subjectID, nil)
	if err != nil {
		return nil, err
	}

	actions := make([]authz.Action, 0, len(permissions))
	for _, perm := range permissions {
		actions = append(actions, authz.Action(perm))
	}
	return actions, nil
}

// grantPermissions grants permissions to a subject for a resource (internal method)
func (m *Manager) grantPermissions(_ context.Context, tenantID, appID, resourceType, resourceID, subjectID, subjectType string, permissions ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)

	// Find or create ACL entry
	var entry *ACLEntry
	for _, e := range m.acls[key] {
		if e.SubjectID == subjectID && e.SubjectType == subjectType {
			entry = e
			break
		}
	}

	if entry == nil {
		entry = &ACLEntry{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			Permissions: []string{},
		}
		m.acls[key] = append(m.acls[key], entry)
	}

	// Add permissions (avoid duplicates)
	for _, perm := range permissions {
		if !contains(entry.Permissions, perm) {
			entry.Permissions = append(entry.Permissions, perm)
		}
	}

	return nil
}

// revokePermissions removes permissions from a subject for a resource (internal method)
func (m *Manager) revokePermissions(_ context.Context, tenantID, appID, resourceType, resourceID, subjectID, subjectType string, permissions ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)

	// Find ACL entry
	for _, entry := range m.acls[key] {
		if entry.SubjectID == subjectID && entry.SubjectType == subjectType {
			// Remove permissions
			newPerms := []string{}
			for _, perm := range entry.Permissions {
				if !contains(permissions, perm) {
					newPerms = append(newPerms, perm)
				}
			}
			entry.Permissions = newPerms
			break
		}
	}

	return nil
}

// RevokeAll removes all permissions from a subject for a resource
func (m *Manager) RevokeAll(ctx context.Context, tenantID, appID, resourceType, resourceID, subjectID, subjectType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)

	// Remove entry
	newACL := []*ACLEntry{}
	for _, entry := range m.acls[key] {
		if entry.SubjectID != subjectID || entry.SubjectType != subjectType {
			newACL = append(newACL, entry)
		}
	}
	m.acls[key] = newACL

	return nil
}

// checkPermission checks if a subject has permission on a resource (internal method)
func (m *Manager) checkPermission(_ context.Context, tenantID, appID, resourceType, resourceID, subjectID, permission string, identity *subject.IdentityContext) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)
	entries := m.acls[key]

	// Check user-specific permissions
	for _, entry := range entries {
		if entry.SubjectType == "user" && entry.SubjectID == subjectID {
			if contains(entry.Permissions, permission) || contains(entry.Permissions, "*") {
				return true, nil
			}
		}
	}

	// Check role-based permissions
	if identity != nil {
		for _, role := range identity.Roles {
			for _, entry := range entries {
				if entry.SubjectType == "role" && entry.SubjectID == role {
					if contains(entry.Permissions, permission) || contains(entry.Permissions, "*") {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

// GetPermissions gets all permissions for a subject on a resource
func (m *Manager) GetPermissions(ctx context.Context, tenantID, appID, resourceType, resourceID, subjectID string, identity *subject.IdentityContext) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)
	entries := m.acls[key]

	permSet := make(map[string]bool)

	// Get user-specific permissions
	for _, entry := range entries {
		if entry.SubjectType == "user" && entry.SubjectID == subjectID {
			for _, perm := range entry.Permissions {
				permSet[perm] = true
			}
		}
	}

	// Get role-based permissions
	if identity != nil {
		for _, role := range identity.Roles {
			for _, entry := range entries {
				if entry.SubjectType == "role" && entry.SubjectID == role {
					for _, perm := range entry.Permissions {
						permSet[perm] = true
					}
				}
			}
		}
	}

	// Convert to slice
	permissions := make([]string, 0, len(permSet))
	for perm := range permSet {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// GetSubjects gets all subjects with permissions on a resource
func (m *Manager) GetSubjects(ctx context.Context, tenantID, appID, resourceType, resourceID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)
	entries := m.acls[key]

	subjects := make([]string, 0, len(entries))
	for _, entry := range entries {
		subjects = append(subjects, fmt.Sprintf("%s:%s", entry.SubjectType, entry.SubjectID))
	}

	return subjects, nil
}

// GetACL gets the full ACL for a resource
func (m *Manager) GetACL(ctx context.Context, tenantID, appID, resourceType, resourceID string) ([]*ACLEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)
	entries := m.acls[key]

	// Return a copy
	result := make([]*ACLEntry, len(entries))
	for i, entry := range entries {
		result[i] = &ACLEntry{
			SubjectID:   entry.SubjectID,
			SubjectType: entry.SubjectType,
			Permissions: append([]string{}, entry.Permissions...),
		}
	}

	return result, nil
}

// SetACL sets the full ACL for a resource (replaces existing)
func (m *Manager) SetACL(ctx context.Context, tenantID, appID, resourceType, resourceID string, entries []*ACLEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)
	m.acls[key] = entries

	return nil
}

// DeleteACL deletes the entire ACL for a resource
func (m *Manager) DeleteACL(ctx context.Context, tenantID, appID, resourceType, resourceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.resourceKey(tenantID, appID, resourceType, resourceID)
	delete(m.acls, key)

	return nil
}

// CopyACL copies ACL from one resource to another
func (m *Manager) CopyACL(ctx context.Context, tenantID, appID, srcType, srcID, dstType, dstID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	srcKey := m.resourceKey(tenantID, appID, srcType, srcID)
	dstKey := m.resourceKey(tenantID, appID, dstType, dstID)

	srcEntries := m.acls[srcKey]

	// Deep copy entries
	dstEntries := make([]*ACLEntry, len(srcEntries))
	for i, entry := range srcEntries {
		dstEntries[i] = &ACLEntry{
			SubjectID:   entry.SubjectID,
			SubjectType: entry.SubjectType,
			Permissions: append([]string{}, entry.Permissions...),
		}
	}

	m.acls[dstKey] = dstEntries

	return nil
}

// Evaluate evaluates an authorization request using ACL
func (m *Manager) Evaluate(ctx context.Context, request *authz.AuthorizationRequest) (*authz.AuthorizationDecision, error) {
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

	allowed, err := m.checkPermission(
		ctx,
		tenantID,
		appID,
		request.Resource.Type,
		request.Resource.ID,
		request.Subject.Subject.ID,
		string(request.Action),
		request.Subject,
	)

	if err != nil {
		return nil, err
	}

	reason := "access denied by ACL"
	if allowed {
		reason = "access granted by ACL"
	}

	return &authz.AuthorizationDecision{
		Allowed: allowed,
		Reason:  reason,
		Metadata: map[string]any{
			"resource": fmt.Sprintf("%s:%s", request.Resource.Type, request.Resource.ID),
			"action":   request.Action,
		},
	}, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
