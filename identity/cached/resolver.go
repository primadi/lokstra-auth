package cached

import (
	"context"
	"fmt"
	"sync"
	"time"

	identity "github.com/primadi/lokstra-auth/identity"
)

// Resolver wraps an identity resolver with caching
type Resolver struct {
	baseResolver identity.IdentityResolver
	cache        identity.IdentityCache
	ttl          time.Duration
}

// NewResolver creates a new cached identity resolver
func NewResolver(baseResolver identity.IdentityResolver, cache identity.IdentityCache, ttl time.Duration) *Resolver {
	if cache == nil {
		cache = NewInMemoryCache()
	}
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	return &Resolver{
		baseResolver: baseResolver,
		cache:        cache,
		ttl:          ttl,
	}
}

// Resolve creates a Subject from claims with caching
func (r *Resolver) Resolve(ctx context.Context, claims map[string]any) (*identity.Subject, error) {
	// Generate cache key from subject ID + tenant ID for proper isolation
	subID, ok := claims["sub"].(string)
	if !ok || subID == "" {
		// Can't cache without subject ID, resolve directly
		return r.baseResolver.Resolve(ctx, claims)
	}

	// Extract tenant_id for cache key scoping
	tenantID, ok := claims["tenant_id"].(string)
	if !ok || tenantID == "" {
		// Can't cache without tenant ID, resolve directly
		return r.baseResolver.Resolve(ctx, claims)
	}

	// Use composite cache key for tenant isolation
	cacheKey := fmt.Sprintf("subject:%s:%s", tenantID, subID)

	// Try to get from cache
	if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != nil && cached.Subject != nil {
		return cached.Subject, nil
	}

	// Cache miss, resolve from base resolver
	sub, err := r.baseResolver.Resolve(ctx, claims)
	if err != nil {
		return nil, err
	}

	// Cache the result
	identity := &identity.IdentityContext{
		Subject:  sub,
		TenantID: tenantID,
	}
	_ = r.cache.Set(ctx, cacheKey, identity, int64(r.ttl.Seconds()))

	return sub, nil
}

// ContextBuilder wraps an identity context builder with caching
type ContextBuilder struct {
	baseBuilder identity.IdentityContextBuilder
	cache       identity.IdentityCache
	ttl         time.Duration
}

// NewContextBuilder creates a new cached identity context builder
func NewContextBuilder(baseBuilder identity.IdentityContextBuilder, cache identity.IdentityCache, ttl time.Duration) *ContextBuilder {
	if cache == nil {
		cache = NewInMemoryCache()
	}
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	return &ContextBuilder{
		baseBuilder: baseBuilder,
		cache:       cache,
		ttl:         ttl,
	}
}

// Build creates an IdentityContext with caching
func (b *ContextBuilder) Build(ctx context.Context, sub *identity.Subject) (*identity.IdentityContext, error) {
	// Extract app_id from subject attributes for cache key scoping
	appID, ok := sub.Attributes["app_id"].(string)
	if !ok || appID == "" {
		// Can't cache without app ID, build directly
		return b.baseBuilder.Build(ctx, sub)
	}

	// Use composite cache key for tenant+app isolation
	cacheKey := fmt.Sprintf("identity:%s:%s:%s", sub.TenantID, appID, sub.ID)

	// Try to get from cache
	if cached, err := b.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Cache miss, build from base builder
	identity, err := b.baseBuilder.Build(ctx, sub)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = b.cache.Set(ctx, cacheKey, identity, int64(b.ttl.Seconds()))

	return identity, nil
}

// Invalidate invalidates cached identity for a subject
func (b *ContextBuilder) Invalidate(ctx context.Context, tenantID, appID, subjectID string) error {
	// Use composite cache key matching Build() method
	cacheKey := fmt.Sprintf("identity:%s:%s:%s", tenantID, appID, subjectID)
	return b.cache.Delete(ctx, cacheKey)
}

// InMemoryCache is an in-memory implementation of IdentityCache
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	identity  *identity.IdentityContext
	expiresAt time.Time
}

// NewInMemoryCache creates a new in-memory identity cache
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		items: make(map[string]*cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set caches an identity context
func (c *InMemoryCache) Set(ctx context.Context, key string, identity *identity.IdentityContext, ttl int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Now().Add(time.Duration(ttl) * time.Second)
	c.items[key] = &cacheItem{
		identity:  identity,
		expiresAt: expiresAt,
	}

	return nil
}

// Get retrieves a cached identity context
func (c *InMemoryCache) Get(ctx context.Context, key string) (*identity.IdentityContext, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil, fmt.Errorf("cache miss")
	}

	// Check expiration
	if time.Now().After(item.expiresAt) {
		return nil, fmt.Errorf("cache expired")
	}

	return item.identity, nil
}

// Delete removes a cached identity context
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Clear clears all cached identity contexts
func (c *InMemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	return nil
}

// cleanup removes expired items periodically
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
