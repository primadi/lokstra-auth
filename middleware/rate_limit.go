package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra/core/request"
)

// RateLimiter is the interface for rate limiting implementations
type RateLimiter interface {
	// Allow checks if the request is allowed under rate limit
	Allow(ctx context.Context, key string) (bool, error)
	// Remaining returns how many requests are left in current window
	Remaining(ctx context.Context, key string) (int, error)
}

// RateLimitMiddleware limits requests per tenant/user/IP
type RateLimitMiddleware struct {
	limiter RateLimiter
	keyFunc func(c *request.Context) string
	limit   int
	window  time.Duration
}

// RateLimitConfig holds configuration for rate limit middleware
type RateLimitConfig struct {
	// Limiter is the rate limit implementation (Redis, in-memory, etc.)
	Limiter RateLimiter

	// KeyFunc generates the rate limit key from request context
	KeyFunc func(c *request.Context) string

	// Limit is the maximum requests allowed per window
	Limit int

	// Window is the time window for rate limiting
	Window time.Duration
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(config RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: config.Limiter,
		keyFunc: config.KeyFunc,
		limit:   config.Limit,
		window:  config.Window,
	}
}

// Handler returns the middleware handler function
func (m *RateLimitMiddleware) Handler() func(c *request.Context) error {
	return func(c *request.Context) error {
		// Generate rate limit key
		key := m.keyFunc(c)
		if key == "" {
			// If no key, skip rate limiting
			return c.Next()
		}

		// Check if allowed
		allowed, err := m.limiter.Allow(c, key)
		if err != nil {
			// On error, log but allow request (fail open)
			// In production, you might want to fail closed
			return c.Next()
		}

		if !allowed {
			// Get remaining count for headers
			remaining, _ := m.limiter.Remaining(c, key)

			c.Resp.WithStatus(429)
			c.W.Header().Add("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
			c.W.Header().Add("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.W.Header().Add("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(m.window).Unix()))
			c.W.Header().Add("Retry-After", fmt.Sprintf("%d", int(m.window.Seconds())))

			return c.Resp.Json(map[string]any{
				"error":   "Too Many Requests",
				"message": fmt.Sprintf("rate limit exceeded: %d requests per %v", m.limit, m.window),
			})
		}

		// Add rate limit headers
		remaining, _ := m.limiter.Remaining(c, key)
		c.W.Header().Add("X-RateLimit-Limit", fmt.Sprintf("%d", m.limit))
		c.W.Header().Add("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		return c.Next()
	}
}

// Rate limit key functions for common scenarios

// PerTenantKey generates rate limit key per tenant
func PerTenantKey(c *request.Context) string {
	tenantID, ok := GetTenantID(c)
	if !ok {
		// Fallback to identity
		identity, ok := GetIdentity(c)
		if ok {
			return fmt.Sprintf("tenant:%s", identity.TenantID)
		}
		return ""
	}
	return fmt.Sprintf("tenant:%s", tenantID)
}

// PerUserKey generates rate limit key per user
func PerUserKey(c *request.Context) string {
	identity, ok := GetIdentity(c)
	if !ok {
		return ""
	}
	return fmt.Sprintf("tenant:%s:user:%s", identity.TenantID, identity.Subject.ID)
}

// PerIPKey generates rate limit key per IP address
func PerIPKey(c *request.Context) string {
	ip := c.R.RemoteAddr
	if ip == "" {
		ip = c.R.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = c.R.Header.Get("X-Real-IP")
	}
	return fmt.Sprintf("ip:%s", ip)
}

// PerTenantAndIPKey generates rate limit key per tenant and IP
func PerTenantAndIPKey(c *request.Context) string {
	tenantID, _ := GetTenantID(c)
	ip := c.R.RemoteAddr
	if ip == "" {
		ip = c.R.Header.Get("X-Forwarded-For")
	}
	return fmt.Sprintf("tenant:%s:ip:%s", tenantID, ip)
}

// InMemoryRateLimiter is a simple in-memory rate limiter (for development/testing)
// For production, use Redis or similar distributed solution
type InMemoryRateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(limit int, window time.Duration) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (r *InMemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()
	cutoff := now.Add(-r.window)

	// Clean old requests
	timestamps := r.requests[key]
	var valid []time.Time
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}

	// Check limit
	if len(valid) >= r.limit {
		r.requests[key] = valid
		return false, nil
	}

	// Add current request
	valid = append(valid, now)
	r.requests[key] = valid

	return true, nil
}

func (r *InMemoryRateLimiter) Remaining(ctx context.Context, key string) (int, error) {
	now := time.Now()
	cutoff := now.Add(-r.window)

	timestamps := r.requests[key]
	count := 0
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			count++
		}
	}

	remaining := r.limit - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}
