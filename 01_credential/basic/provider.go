package basic

import (
	"context"
	"sync"
)

// InMemoryUserProvider is a simple in-memory implementation of UserProvider for testing
type InMemoryUserProvider struct {
	users map[string]*User
	mu    sync.RWMutex
}

// NewInMemoryUserProvider creates a new in-memory user provider
func NewInMemoryUserProvider() *InMemoryUserProvider {
	return &InMemoryUserProvider{
		users: make(map[string]*User),
	}
}

// GetUserByUsername retrieves user information by username
func (p *InMemoryUserProvider) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	user, exists := p.users[username]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// AddUser adds a user to the provider
func (p *InMemoryUserProvider) AddUser(user *User) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.users[user.Username] = user
}

// RemoveUser removes a user from the provider
func (p *InMemoryUserProvider) RemoveUser(username string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.users, username)
}

// UpdateUser updates an existing user
func (p *InMemoryUserProvider) UpdateUser(user *User) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.users[user.Username]; !exists {
		return ErrUserNotFound
	}

	p.users[user.Username] = user
	return nil
}

// ListUsers returns all users (for testing/admin purposes)
func (p *InMemoryUserProvider) ListUsers() []*User {
	p.mu.RLock()
	defer p.mu.RUnlock()

	users := make([]*User, 0, len(p.users))
	for _, user := range p.users {
		users = append(users, user)
	}
	return users
}
