package cached

import (
	"context"
	"fmt"
	"sync"
	"time"

	subject "github.com/primadi/lokstra-auth/03_subject"
)

// Resolver wraps a subject resolver with caching
type Resolver struct {
	baseResolver subject.SubjectResolver
	cache        subject.IdentityCache
	ttl          time.Duration
}

// NewResolver creates a new cached subject resolver
func NewResolver(baseResolver subject.SubjectResolver, cache subject.IdentityCache, ttl time.Duration) *Resolver {
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
func (r *Resolver) Resolve(ctx context.Context, claims map[string]any) (*subject.Subject, error) {
	// Generate cache key from subject ID
	subID, ok := claims["sub"].(string)
	if !ok || subID == "" {
		// Can't cache without subject ID, resolve directly
		return r.baseResolver.Resolve(ctx, claims)
	}

	cacheKey := fmt.Sprintf("subject:%s", subID)

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
	identity := &subject.IdentityContext{
		Subject: sub,
	}
	_ = r.cache.Set(ctx, cacheKey, identity, int64(r.ttl.Seconds()))

	return sub, nil
}

// ContextBuilder wraps an identity context builder with caching
type ContextBuilder struct {
	baseBuilder subject.IdentityContextBuilder
	cache       subject.IdentityCache
	ttl         time.Duration
}

// NewContextBuilder creates a new cached identity context builder
func NewContextBuilder(baseBuilder subject.IdentityContextBuilder, cache subject.IdentityCache, ttl time.Duration) *ContextBuilder {
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
func (b *ContextBuilder) Build(ctx context.Context, sub *subject.Subject) (*subject.IdentityContext, error) {
	cacheKey := fmt.Sprintf("identity:%s", sub.ID)

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
func (b *ContextBuilder) Invalidate(ctx context.Context, subjectID string) error {
	cacheKey := fmt.Sprintf("identity:%s", subjectID)
	return b.cache.Delete(ctx, cacheKey)
}

// InMemoryCache is an in-memory implementation of IdentityCache
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	identity  *subject.IdentityContext
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
func (c *InMemoryCache) Set(ctx context.Context, key string, identity *subject.IdentityContext, ttl int64) error {
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
func (c *InMemoryCache) Get(ctx context.Context, key string) (*subject.IdentityContext, error) {
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
