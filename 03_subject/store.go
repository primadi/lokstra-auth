package subject

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InMemoryIdentityStore is an in-memory implementation of IdentityStore
type InMemoryIdentityStore struct {
	mu         sync.RWMutex
	identities map[string]*IdentityContext
	expiresAt  map[string]time.Time
}

// NewInMemoryIdentityStore creates a new in-memory identity store
func NewInMemoryIdentityStore() *InMemoryIdentityStore {
	store := &InMemoryIdentityStore{
		identities: make(map[string]*IdentityContext),
		expiresAt:  make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go store.cleanup()

	return store
}

// Store saves an identity context
func (s *InMemoryIdentityStore) Store(ctx context.Context, sessionID string, identity *IdentityContext) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.identities[sessionID] = identity

	// Set expiration if session has ExpiresAt
	if identity.Session != nil && identity.Session.ExpiresAt > 0 {
		s.expiresAt[sessionID] = time.Unix(identity.Session.ExpiresAt, 0)
	} else {
		// Default expiration: 24 hours
		s.expiresAt[sessionID] = time.Now().Add(24 * time.Hour)
	}

	return nil
}

// Get retrieves an identity context
func (s *InMemoryIdentityStore) Get(ctx context.Context, sessionID string) (*IdentityContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	identity, ok := s.identities[sessionID]
	if !ok {
		return nil, fmt.Errorf("identity not found for session: %s", sessionID)
	}

	// Check expiration
	if expiresAt, ok := s.expiresAt[sessionID]; ok {
		if time.Now().After(expiresAt) {
			return nil, fmt.Errorf("session expired: %s", sessionID)
		}
	}

	return identity, nil
}

// Delete removes an identity context
func (s *InMemoryIdentityStore) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.identities, sessionID)
	delete(s.expiresAt, sessionID)
	return nil
}

// Update updates an identity context
func (s *InMemoryIdentityStore) Update(ctx context.Context, sessionID string, identity *IdentityContext) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.identities[sessionID]; !ok {
		return fmt.Errorf("identity not found for session: %s", sessionID)
	}

	s.identities[sessionID] = identity
	return nil
}

// Cleanup removes expired sessions
func (s *InMemoryIdentityStore) Cleanup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for sessionID, expiresAt := range s.expiresAt {
		if now.After(expiresAt) {
			delete(s.identities, sessionID)
			delete(s.expiresAt, sessionID)
		}
	}

	return nil
}

// cleanup runs periodic cleanup of expired sessions
func (s *InMemoryIdentityStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		_ = s.Cleanup(context.Background())
	}
}

// ListSessions lists all active session IDs
func (s *InMemoryIdentityStore) ListSessions(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]string, 0, len(s.identities))
	now := time.Now()

	for sessionID, expiresAt := range s.expiresAt {
		if now.Before(expiresAt) {
			sessions = append(sessions, sessionID)
		}
	}

	return sessions, nil
}

// ListBySubject lists all sessions for a subject
func (s *InMemoryIdentityStore) ListBySubject(ctx context.Context, subjectID string) ([]*IdentityContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	identities := make([]*IdentityContext, 0)
	now := time.Now()

	for sessionID, identity := range s.identities {
		// Check expiration
		if expiresAt, ok := s.expiresAt[sessionID]; ok {
			if now.After(expiresAt) {
				continue
			}
		}

		// Check if subject matches
		if identity.Subject != nil && identity.Subject.ID == subjectID {
			identities = append(identities, identity)
		}
	}

	return identities, nil
}

// DeleteBySubject deletes all sessions for a subject
func (s *InMemoryIdentityStore) DeleteBySubject(ctx context.Context, subjectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionsToDelete := make([]string, 0)

	for sessionID, identity := range s.identities {
		if identity.Subject != nil && identity.Subject.ID == subjectID {
			sessionsToDelete = append(sessionsToDelete, sessionID)
		}
	}

	for _, sessionID := range sessionsToDelete {
		delete(s.identities, sessionID)
		delete(s.expiresAt, sessionID)
	}

	return nil
}
