package passkey

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/go-webauthn/webauthn/webauthn"
)

// CredentialStore manages passkey credentials
type CredentialStore interface {
	// StoreCredential stores a passkey credential for a user
	StoreCredential(ctx context.Context, userID string, credential *webauthn.Credential) error

	// UpdateCredential updates an existing credential
	UpdateCredential(ctx context.Context, userID string, credential *webauthn.Credential) error

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, userID string) (*User, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *User) error

	// DeleteCredential removes a credential
	DeleteCredential(ctx context.Context, userID string, credentialID []byte) error

	// ListCredentials lists all credentials for a user
	ListCredentials(ctx context.Context, userID string) ([]webauthn.Credential, error)
}

// InMemoryCredentialStore implements in-memory storage for passkey credentials
type InMemoryCredentialStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

// NewInMemoryCredentialStore creates a new in-memory credential store
func NewInMemoryCredentialStore() *InMemoryCredentialStore {
	return &InMemoryCredentialStore{
		users: make(map[string]*User),
	}
}

// StoreCredential stores a passkey credential
func (s *InMemoryCredentialStore) StoreCredential(ctx context.Context, userID string, credential *webauthn.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Check if credential already exists
	for i, cred := range user.Credentials {
		if string(cred.ID) == string(credential.ID) {
			// Update existing
			user.Credentials[i] = *credential
			return nil
		}
	}

	// Add new credential
	user.Credentials = append(user.Credentials, *credential)
	return nil
}

// UpdateCredential updates an existing credential
func (s *InMemoryCredentialStore) UpdateCredential(ctx context.Context, userID string, credential *webauthn.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Find and update credential
	for i, cred := range user.Credentials {
		if string(cred.ID) == string(credential.ID) {
			user.Credentials[i] = *credential
			return nil
		}
	}

	return fmt.Errorf("credential not found")
}

// GetUser retrieves a user by ID
func (s *InMemoryCredentialStore) GetUser(ctx context.Context, userID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	return user, nil
}

// CreateUser creates a new user
func (s *InMemoryCredentialStore) CreateUser(ctx context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userID := base64.StdEncoding.EncodeToString(user.ID)
	if _, exists := s.users[userID]; exists {
		return fmt.Errorf("user already exists: %s", userID)
	}

	s.users[userID] = user
	return nil
}

// DeleteCredential removes a credential
func (s *InMemoryCredentialStore) DeleteCredential(ctx context.Context, userID string, credentialID []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Find and remove credential
	for i, cred := range user.Credentials {
		if string(cred.ID) == string(credentialID) {
			user.Credentials = append(user.Credentials[:i], user.Credentials[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("credential not found")
}

// ListCredentials lists all credentials for a user
func (s *InMemoryCredentialStore) ListCredentials(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	return user.Credentials, nil
}
