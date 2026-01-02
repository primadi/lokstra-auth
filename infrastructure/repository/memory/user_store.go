package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/primadi/lokstra-auth/core/domain"
	"github.com/primadi/lokstra-auth/infrastructure/repository"
)

// InMemoryUserStore is an in-memory implementation of UserStore
type InMemoryUserStore struct {
	// users: map[tenantID:userID] -> User
	users map[string]*domain.User
	mu    sync.RWMutex
}

var _ repository.UserStore = (*InMemoryUserStore)(nil)

// NewUserStore creates a new in-memory user store
func NewUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*domain.User),
	}
}

// makeKey creates a composite key for tenant+user
func (s *InMemoryUserStore) makeKey(tenantID, userID string) string {
	return tenantID + ":" + userID
}

// Create creates a new user
func (s *InMemoryUserStore) Create(ctx context.Context, user *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(user.TenantID, user.ID)
	if _, exists := s.users[key]; exists {
		return fmt.Errorf("user already exists: %s in tenant %s", user.ID, user.TenantID)
	}

	s.users[key] = user
	return nil
}

// Get retrieves a user by ID within a tenant
func (s *InMemoryUserStore) Get(ctx context.Context, tenantID, userID string) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, userID)
	user, exists := s.users[key]
	if !exists {
		return nil, fmt.Errorf("user not found: %s in tenant %s", userID, tenantID)
	}

	return user, nil
}

// Update updates an existing user
func (s *InMemoryUserStore) Update(ctx context.Context, user *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(user.TenantID, user.ID)
	if _, exists := s.users[key]; !exists {
		return fmt.Errorf("user not found: %s in tenant %s", user.ID, user.TenantID)
	}

	s.users[key] = user
	return nil
}

// Delete deletes a user
func (s *InMemoryUserStore) Delete(ctx context.Context, tenantID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, userID)
	delete(s.users, key)
	return nil
}

// List lists all users for a tenant
func (s *InMemoryUserStore) List(ctx context.Context, tenantID string) ([]*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*domain.User, 0)
	for _, user := range s.users {
		if user.TenantID == tenantID {
			users = append(users, user)
		}
	}

	return users, nil
}

// GetByUsername retrieves a user by username within a tenant
func (s *InMemoryUserStore) GetByUsername(ctx context.Context, tenantID, username string) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.TenantID == tenantID && user.Username == username {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found with username: %s in tenant %s", username, tenantID)
}

// GetByEmail retrieves a user by email within a tenant
func (s *InMemoryUserStore) GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.TenantID == tenantID && user.Email == email {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found with email: %s in tenant %s", email, tenantID)
}

// Exists checks if a user exists within a tenant
func (s *InMemoryUserStore) Exists(ctx context.Context, tenantID, userID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.makeKey(tenantID, userID)
	_, exists := s.users[key]
	return exists, nil
}

// ListByApp lists users assigned to an app
// Note: This is a placeholder. Implement using UserAppStore for production
func (s *InMemoryUserStore) ListByApp(ctx context.Context, tenantID, appID string) ([]*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// TODO: Query UserAppStore to get user IDs for this app, then fetch users
	// For now, return empty list
	return []*domain.User{}, nil
}

// SetPassword sets or updates user password hash (for basic auth)
func (s *InMemoryUserStore) SetPassword(ctx context.Context, tenantID, userID, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, userID)
	user, exists := s.users[key]
	if !exists {
		return fmt.Errorf("user not found: %s in tenant %s", userID, tenantID)
	}

	user.PasswordHash = &passwordHash
	return nil
}

// RemovePassword removes user password hash (disables basic auth)
func (s *InMemoryUserStore) RemovePassword(ctx context.Context, tenantID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.makeKey(tenantID, userID)
	user, exists := s.users[key]
	if !exists {
		return fmt.Errorf("user not found: %s in tenant %s", userID, tenantID)
	}

	user.PasswordHash = nil
	return nil
}
