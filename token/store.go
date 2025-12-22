package token

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExists   = errors.New("token already exists")
)

// InMemoryTokenStore is an in-memory implementation of TokenStore
type InMemoryTokenStore struct {
	mu      sync.RWMutex
	tokens  map[string]map[string]*Token // subject -> tokenID -> Token
	revoked map[string]bool              // tokenID -> revoked
}

// NewInMemoryTokenStore creates a new in-memory token store
func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		tokens:  make(map[string]map[string]*Token),
		revoked: make(map[string]bool),
	}
}

// Store saves a token
func (s *InMemoryTokenStore) Store(ctx context.Context, subject string, token *Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tokens[subject]; !ok {
		s.tokens[subject] = make(map[string]*Token)
	}

	// Use token value as ID if not in metadata
	tokenID := token.Value
	if id, ok := token.Metadata["token_id"].(string); ok {
		tokenID = id
	}

	s.tokens[subject][tokenID] = token
	return nil
}

// Get retrieves a token
func (s *InMemoryTokenStore) Get(ctx context.Context, subject string, tokenID string) (*Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subjectTokens, ok := s.tokens[subject]
	if !ok {
		return nil, ErrTokenNotFound
	}

	token, ok := subjectTokens[tokenID]
	if !ok {
		return nil, ErrTokenNotFound
	}

	// Check if expired
	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	return token, nil
}

// Delete removes a token
func (s *InMemoryTokenStore) Delete(ctx context.Context, subject string, tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subjectTokens, ok := s.tokens[subject]
	if !ok {
		return ErrTokenNotFound
	}

	delete(subjectTokens, tokenID)

	// Clean up empty subject map
	if len(subjectTokens) == 0 {
		delete(s.tokens, subject)
	}

	return nil
}

// List returns all tokens for a subject
func (s *InMemoryTokenStore) List(ctx context.Context, subject string) ([]*Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subjectTokens, ok := s.tokens[subject]
	if !ok {
		return []*Token{}, nil
	}

	tokens := make([]*Token, 0, len(subjectTokens))
	now := time.Now()

	for _, token := range subjectTokens {
		// Skip expired tokens
		if now.After(token.ExpiresAt) {
			continue
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Revoke invalidates a token
func (s *InMemoryTokenStore) Revoke(ctx context.Context, tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.revoked[tokenID] = true
	return nil
}

// IsRevoked checks if a token is revoked
func (s *InMemoryTokenStore) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.revoked[tokenID], nil
}

// Cleanup removes expired tokens
func (s *InMemoryTokenStore) Cleanup(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for subject, subjectTokens := range s.tokens {
		for tokenID, token := range subjectTokens {
			if now.After(token.ExpiresAt) {
				delete(subjectTokens, tokenID)
				delete(s.revoked, tokenID)
			}
		}

		// Clean up empty subject maps
		if len(subjectTokens) == 0 {
			delete(s.tokens, subject)
		}
	}

	return nil
}
