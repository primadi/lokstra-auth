package revocation_list

import (
	"context"
	"sync"
	"time"

	"github.com/primadi/lokstra-auth/token"
)

// InMemoryRevocationList is an in-memory implementation of TokenRevocationList
// @Service "in-memory-revocation-list"
type InMemoryRevocationList struct {
	mu         sync.RWMutex
	revoked    map[string]time.Time // tokenID -> expiresAt
	ctrCleanUp int
}

var _ token.TokenRevocationList = (*InMemoryRevocationList)(nil)

const MAXCTR = 20 // after 20 operations, do clean-up

// NewInMemoryRevocationList creates a new in-memory revocation list
func (r *InMemoryRevocationList) Init() {
	r.revoked = make(map[string]time.Time)
}

func (r *InMemoryRevocationList) Add(ctx context.Context, tokenID string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revoked[tokenID] = expiresAt
	r.autoCleanUp(MAXCTR)
	return nil
}

func (r *InMemoryRevocationList) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, revoked := r.revoked[tokenID]
	return revoked, nil
}

func (r *InMemoryRevocationList) Remove(ctx context.Context, tokenID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.revoked, tokenID)
	r.autoCleanUp(MAXCTR)
	return nil
}

func (r *InMemoryRevocationList) Cleanup(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ctrCleanUp = MAXCTR // force clean-up
	r.autoCleanUp(MAXCTR)
	return nil
}

func (r *InMemoryRevocationList) autoCleanUp(maxCtr int) {
	r.ctrCleanUp++

	if r.ctrCleanUp > maxCtr {
		now := time.Now()
		for tokenID, expiresAt := range r.revoked {
			if now.After(expiresAt) {
				delete(r.revoked, tokenID)
			}
		}
		r.ctrCleanUp = 0
	}
}
