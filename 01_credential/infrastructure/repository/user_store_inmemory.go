package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/primadi/lokstra-auth/01_credential/domain/basic"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

// InMemoryUserStore provides in-memory user storage for basic authentication
type InMemoryUserStore struct {
	mu    sync.RWMutex
	users map[string]*basic.User // key: "tenantID:username"
}

var _ UserProvider = (*InMemoryUserStore)(nil)

// NewInMemoryUserStore creates a new in-memory user store
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*basic.User),
	}
}

// GetUserByUsername retrieves user by username within a tenant
func (s *InMemoryUserStore) GetUserByUsername(ctx context.Context, tenantID, username string) (*basic.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := tenantID + ":" + username
	user, exists := s.users[key]
	if !exists {
		return nil, basic.ErrUserNotFound
	}

	return user, nil
}

// CreateUser creates a new user
func (s *InMemoryUserStore) CreateUser(ctx context.Context, user *basic.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := user.TenantID + ":" + user.Username
	if _, exists := s.users[key]; exists {
		return ErrUserAlreadyExists
	}

	s.users[key] = user
	return nil
}

// UpdateUser updates existing user
func (s *InMemoryUserStore) UpdateUser(ctx context.Context, user *basic.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := user.TenantID + ":" + user.Username
	if _, exists := s.users[key]; !exists {
		return basic.ErrUserNotFound
	}

	s.users[key] = user
	return nil
}

// GetUserByID retrieves user by ID
func (s *InMemoryUserStore) GetUserByID(ctx context.Context, tenantID, userID string) (*basic.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.TenantID == tenantID && user.ID == userID {
			return user, nil
		}
	}

	return nil, basic.ErrUserNotFound
}
