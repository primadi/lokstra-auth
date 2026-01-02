package enriched

import (
	"context"
	"fmt"

	identity "github.com/primadi/lokstra-auth/identity"
)

// ContextBuilder builds identity contexts with enrichment support
type ContextBuilder struct {
	baseBuilder identity.IdentityContextBuilder
	enrichers   []identity.DataEnricher
}

// NewContextBuilder creates a new enriched identity context builder
func NewContextBuilder(baseBuilder identity.IdentityContextBuilder, enrichers ...identity.DataEnricher) *ContextBuilder {
	return &ContextBuilder{
		baseBuilder: baseBuilder,
		enrichers:   enrichers,
	}
}

// Build creates an enriched IdentityContext from a subject
func (b *ContextBuilder) Build(ctx context.Context, sub *identity.Subject) (*identity.IdentityContext, error) {
	// Build base identity context
	identity, err := b.baseBuilder.Build(ctx, sub)
	if err != nil {
		return nil, fmt.Errorf("failed to build base identity: %w", err)
	}

	// Apply enrichers
	for i, enricher := range b.enrichers {
		if err := enricher.Enrich(ctx, identity); err != nil {
			return nil, fmt.Errorf("enricher %d failed: %w", i, err)
		}
	}

	return identity, nil
}

// AddEnricher adds a new enricher to the builder
func (b *ContextBuilder) AddEnricher(enricher identity.DataEnricher) {
	b.enrichers = append(b.enrichers, enricher)
}

// AttributeEnricher enriches identity with subject attributes
type AttributeEnricher struct {
	// AttributeMapping maps attribute keys to identity metadata keys
	AttributeMapping map[string]string
}

// NewAttributeEnricher creates a new attribute enricher
func NewAttributeEnricher() *AttributeEnricher {
	return &AttributeEnricher{
		AttributeMapping: make(map[string]string),
	}
}

// Enrich adds subject attributes to identity metadata
func (e *AttributeEnricher) Enrich(ctx context.Context, identity *identity.IdentityContext) error {
	if identity.Subject == nil || identity.Subject.Attributes == nil {
		return nil
	}

	if identity.Metadata == nil {
		identity.Metadata = make(map[string]any)
	}

	// Copy attributes to metadata with optional mapping
	for key, value := range identity.Subject.Attributes {
		targetKey := key
		if mapped, ok := e.AttributeMapping[key]; ok {
			targetKey = mapped
		}
		identity.Metadata[targetKey] = value
	}

	return nil
}

// RoleBasedEnricher enriches identity based on roles
type RoleBasedEnricher struct {
	// RoleAttributes maps roles to additional attributes
	RoleAttributes map[string]map[string]any
}

// NewRoleBasedEnricher creates a new role-based enricher
func NewRoleBasedEnricher(roleAttributes map[string]map[string]any) *RoleBasedEnricher {
	return &RoleBasedEnricher{
		RoleAttributes: roleAttributes,
	}
}

// Enrich adds role-based attributes to identity
func (e *RoleBasedEnricher) Enrich(ctx context.Context, identity *identity.IdentityContext) error {
	if identity.Metadata == nil {
		identity.Metadata = make(map[string]any)
	}

	// Add attributes for each role
	for _, role := range identity.Roles {
		if attrs, ok := e.RoleAttributes[role]; ok {
			for key, value := range attrs {
				// Don't overwrite existing metadata
				if _, exists := identity.Metadata[key]; !exists {
					identity.Metadata[key] = value
				}
			}
		}
	}

	return nil
}

// ProfileEnricher enriches identity with profile data
type ProfileEnricher struct {
	profileProvider identity.ProfileProvider
}

// NewProfileEnricher creates a new profile enricher
func NewProfileEnricher(profileProvider identity.ProfileProvider) *ProfileEnricher {
	return &ProfileEnricher{
		profileProvider: profileProvider,
	}
}

// Enrich adds profile data to identity
func (e *ProfileEnricher) Enrich(ctx context.Context, identity *identity.IdentityContext) error {
	if e.profileProvider == nil {
		return nil
	}

	// Use tenant ID from identity context (already set by builder)
	profile, err := e.profileProvider.GetProfile(ctx, identity.TenantID, identity.Subject)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	identity.Profile = profile
	return nil
}

// CustomEnricher allows custom enrichment logic
type CustomEnricher struct {
	enrichFunc func(context.Context, *identity.IdentityContext) error
}

// NewCustomEnricher creates a new custom enricher
func NewCustomEnricher(fn func(context.Context, *identity.IdentityContext) error) *CustomEnricher {
	return &CustomEnricher{
		enrichFunc: fn,
	}
}

// Enrich executes custom enrichment logic
func (e *CustomEnricher) Enrich(ctx context.Context, identity *identity.IdentityContext) error {
	if e.enrichFunc == nil {
		return nil
	}
	return e.enrichFunc(ctx, identity)
}

// SessionEnricher enriches identity with session information
type SessionEnricher struct{}

// NewSessionEnricher creates a new session enricher
func NewSessionEnricher() *SessionEnricher {
	return &SessionEnricher{}
}

// Enrich adds session information to identity
func (e *SessionEnricher) Enrich(ctx context.Context, ic *identity.IdentityContext) error {
	// Extract session info from context if available
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		if ic.Session == nil {
			ic.Session = &identity.SessionInfo{}
		}
		ic.Session.ID = sessionID
	}

	if ipAddress, ok := ctx.Value("ip_address").(string); ok {
		if ic.Session == nil {
			ic.Session = &identity.SessionInfo{}
		}
		ic.Session.IPAddress = ipAddress
	}

	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		if ic.Session == nil {
			ic.Session = &identity.SessionInfo{}
		}
		ic.Session.UserAgent = userAgent
	}

	return nil
}

// ChainEnricher chains multiple enrichers
type ChainEnricher struct {
	enrichers []identity.DataEnricher
}

// NewChainEnricher creates a new chain enricher
func NewChainEnricher(enrichers ...identity.DataEnricher) *ChainEnricher {
	return &ChainEnricher{
		enrichers: enrichers,
	}
}

// Enrich applies all enrichers in sequence
func (e *ChainEnricher) Enrich(ctx context.Context, identity *identity.IdentityContext) error {
	for i, enricher := range e.enrichers {
		if err := enricher.Enrich(ctx, identity); err != nil {
			return fmt.Errorf("enricher %d failed: %w", i, err)
		}
	}
	return nil
}

// Add adds a new enricher to the chain
func (e *ChainEnricher) Add(enricher identity.DataEnricher) {
	e.enrichers = append(e.enrichers, enricher)
}
