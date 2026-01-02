-- ============================================================
-- Lokstra Auth - PostgreSQL Database Schema
-- ============================================================
-- This script creates the necessary tables for the lokstra-auth
-- core entities: Tenants, Apps, Branches, Users, User-App Access, and App Keys
-- ============================================================

CREATE SCHEMA IF NOT EXISTS lokstra_core;

SET SEARCH_PATH TO lokstra_core;

-- Drop tables if they exist (for clean re-creation)
-- DROP TABLE IF EXISTS user_apps CASCADE;
-- DROP TABLE IF EXISTS app_keys CASCADE;
-- DROP TABLE IF EXISTS branches CASCADE;
-- DROP TABLE IF EXISTS users CASCADE;
-- DROP TABLE IF EXISTS apps CASCADE;
-- DROP TABLE IF EXISTS tenants CASCADE;

-- ============================================================
-- Tenants Table
-- ============================================================
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    owner_id VARCHAR(255) NOT NULL,
    db_dsn TEXT NOT NULL,
    db_schema VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    config JSONB,
    settings JSONB,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    -- Indexes
    CONSTRAINT unique_tenant_name UNIQUE (name)
);

CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_domain ON tenants(domain) WHERE domain IS NOT NULL;
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_tenants_owner ON tenants(owner_id);

COMMENT ON TABLE tenants IS 'Multi-tenant organizations/companies';
COMMENT ON COLUMN tenants.id IS 'Unique tenant identifier (e.g., acme-corp)';
COMMENT ON COLUMN tenants.owner_id IS 'User ID of the tenant owner (billing and legal owner) - exactly 1 per tenant';
COMMENT ON COLUMN tenants.db_dsn IS 'Database connection string for multi-database tenancy';
COMMENT ON COLUMN tenants.db_schema IS 'Database schema name for schema-based tenancy';
COMMENT ON COLUMN tenants.status IS 'Tenant status: active, suspended, deleted';

-- ============================================================
-- Apps Table
-- ============================================================
CREATE TABLE IF NOT EXISTS apps (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    config JSONB,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Composite primary key
    PRIMARY KEY (tenant_id, id),
    
    -- Unique constraint on name per tenant
    CONSTRAINT unique_app_name_per_tenant UNIQUE (tenant_id, name)
);

CREATE INDEX idx_apps_tenant_id ON apps(tenant_id);
CREATE INDEX idx_apps_type ON apps(type);
CREATE INDEX idx_apps_status ON apps(status);

COMMENT ON TABLE apps IS 'Applications within a tenant (web, mobile, API, service)';
COMMENT ON COLUMN apps.type IS 'App type: web, mobile, api, desktop, service';
COMMENT ON COLUMN apps.status IS 'App status: active, disabled';

-- ============================================================
-- Branches Table
-- ============================================================
CREATE TABLE IF NOT EXISTS branches (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    address JSONB,
    contact JSONB,
    settings JSONB,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Composite primary key
    PRIMARY KEY (tenant_id, app_id, id),
    
    -- Foreign key to apps
    CONSTRAINT fk_branches_app FOREIGN KEY (tenant_id, app_id) 
        REFERENCES apps(tenant_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_branches_tenant_app ON branches(tenant_id, app_id);
CREATE INDEX idx_branches_type ON branches(type);
CREATE INDEX idx_branches_status ON branches(status);

COMMENT ON TABLE branches IS 'Branch/location within a tenant app (stores, offices, warehouses)';
COMMENT ON COLUMN branches.type IS 'Branch type: headquarters, regional, store, warehouse, franchise, office, other';
COMMENT ON COLUMN branches.status IS 'Branch status: active, disabled';

-- ============================================================
-- Users Table
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    password_hash TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    is_tenant_owner BOOLEAN NOT NULL DEFAULT FALSE,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    last_failed_login_at TIMESTAMP,
    locked_at TIMESTAMP,
    locked_until TIMESTAMP,
    lockout_count INT NOT NULL DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    -- Composite primary key
    PRIMARY KEY (tenant_id, id),
    
    -- Unique constraints per tenant
    CONSTRAINT unique_username_per_tenant UNIQUE (tenant_id, username),
    CONSTRAINT unique_email_per_tenant UNIQUE (tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_username ON users(tenant_id, username);
CREATE INDEX idx_users_email ON users(tenant_id, email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_users_tenant_owner ON users(tenant_id, is_tenant_owner) WHERE is_tenant_owner = TRUE;
CREATE INDEX idx_users_locked ON users(tenant_id, locked_until) WHERE locked_until IS NOT NULL;
CREATE INDEX idx_users_failed_login ON users(tenant_id, failed_login_attempts, last_failed_login_at);

COMMENT ON TABLE users IS 'Users within a tenant';
COMMENT ON COLUMN users.password_hash IS 'Optional bcrypt password hash for basic auth (NULL if user uses OAuth/passkey only)';
COMMENT ON COLUMN users.status IS 'User status: active, suspended, deleted, locked';
COMMENT ON COLUMN users.is_tenant_owner IS 'TRUE if this user is the owner of their tenant (only 1 per tenant)';
COMMENT ON COLUMN users.failed_login_attempts IS 'Counter for consecutive failed login attempts';
COMMENT ON COLUMN users.last_failed_login_at IS 'Timestamp of the most recent failed login attempt';
COMMENT ON COLUMN users.locked_at IS 'Timestamp when account was locked due to failed attempts';
COMMENT ON COLUMN users.locked_until IS 'Timestamp when account will be automatically unlocked';
COMMENT ON COLUMN users.lockout_count IS 'Total number of times account has been locked';

-- ============================================================
-- User Identities Table (Linked Authentication Providers)
-- ============================================================
CREATE TABLE IF NOT EXISTS user_identities (
    id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    username VARCHAR(255),
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Composite primary key
    PRIMARY KEY (tenant_id, user_id, id),
    
    -- Foreign key to users
    CONSTRAINT fk_user_identities_user FOREIGN KEY (tenant_id, user_id) 
        REFERENCES users(tenant_id, id) ON DELETE CASCADE,
    
    -- Unique constraint: one provider identity can only be linked to one user per tenant
    CONSTRAINT unique_provider_identity_per_tenant UNIQUE (tenant_id, provider, provider_id)
);

CREATE INDEX idx_user_identities_user ON user_identities(tenant_id, user_id);
CREATE INDEX idx_user_identities_provider ON user_identities(tenant_id, provider, provider_id);

COMMENT ON TABLE user_identities IS 'Linked authentication provider identities (OAuth2, SAML, passkey, etc.)';
COMMENT ON COLUMN user_identities.provider IS 'Provider type: local, google, github, microsoft, facebook, apple, passkey, magic_link, otp, saml, oidc, ldap';
COMMENT ON COLUMN user_identities.provider_id IS 'Unique ID from provider (e.g., Google sub, GitHub user ID)';
COMMENT ON COLUMN user_identities.verified IS 'Email verification status from provider';

-- ============================================================
-- Credential Providers Table
-- ============================================================
-- Manages OAuth2/SAML/Email provider configurations
-- Supports multiple providers per tenant+app (e.g., 2 Google OAuth2 configs)
-- ============================================================
CREATE TABLE IF NOT EXISTS credential_providers (
    id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255),  -- NULL = tenant-level default, non-NULL = app-specific
    type VARCHAR(50) NOT NULL,  -- oauth2_google, oauth2_github, saml, email, etc.
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',  -- active, disabled
    config JSONB NOT NULL,  -- Provider-specific configuration (ClientID, ClientSecret, etc.)
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Primary key
    PRIMARY KEY (tenant_id, id),
    
    -- Foreign keys
    CONSTRAINT fk_credential_providers_tenant FOREIGN KEY (tenant_id)
        REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_credential_providers_app FOREIGN KEY (tenant_id, app_id)
        REFERENCES apps(tenant_id, id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_credential_providers_type ON credential_providers(tenant_id, type);
CREATE INDEX idx_credential_providers_app_id ON credential_providers(tenant_id, app_id) WHERE app_id IS NOT NULL;
CREATE INDEX idx_credential_providers_status ON credential_providers(tenant_id, status);

COMMENT ON TABLE credential_providers IS 'OAuth2/SAML/Email provider configurations (supports multiple configs per provider type)';
COMMENT ON COLUMN credential_providers.type IS 'Provider type: oauth2_google, oauth2_github, saml, email, etc.';
COMMENT ON COLUMN credential_providers.app_id IS 'NULL for tenant-level default, specific app_id for app-level override';
COMMENT ON COLUMN credential_providers.config IS 'JSONB config: {client_id, client_secret, scopes, ...}';

-- Example data:
-- Tenant "acme-corp" with 2 Google OAuth2 configs:
-- INSERT INTO credential_providers (id, tenant_id, app_id, type, name, config)
-- VALUES 
--   ('google-web', 'acme-corp', 'web-portal', 'oauth2_google', 'Google OAuth (Web)', '{"client_id": "xxx-web.apps.googleusercontent.com", ...}'),
--   ('google-mobile', 'acme-corp', 'mobile-app', 'oauth2_google', 'Google OAuth (Mobile)', '{"client_id": "yyy-mobile.apps.googleusercontent.com", ...}');

-- ============================================================
-- User-App Access Table (Access Control)
-- ============================================================
CREATE TABLE IF NOT EXISTS user_apps (
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP,
    
    -- Composite primary key
    PRIMARY KEY (tenant_id, app_id, user_id),
    
    -- Foreign keys
    CONSTRAINT fk_user_apps_user FOREIGN KEY (tenant_id, user_id) 
        REFERENCES users(tenant_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_user_apps_app FOREIGN KEY (tenant_id, app_id) 
        REFERENCES apps(tenant_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_user_apps_user ON user_apps(tenant_id, user_id);
CREATE INDEX idx_user_apps_app ON user_apps(tenant_id, app_id);
CREATE INDEX idx_user_apps_status ON user_apps(status);

COMMENT ON TABLE user_apps IS 'User access control to apps (not authorization - see authz)';
COMMENT ON COLUMN user_apps.status IS 'Access status: active, revoked';

-- ============================================================
-- App Keys Table (API Keys for Application Authentication)
-- ============================================================
CREATE TABLE IF NOT EXISTS app_keys (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    key_id VARCHAR(255) NOT NULL,
    prefix VARCHAR(50) NOT NULL,
    secret_hash TEXT NOT NULL,
    key_type VARCHAR(50) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    user_id VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    scopes JSONB,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    last_used TIMESTAMP,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at TIMESTAMP,
    
    -- Foreign key to apps
    CONSTRAINT fk_app_keys_app FOREIGN KEY (tenant_id, app_id) 
        REFERENCES apps(tenant_id, id) ON DELETE CASCADE,
    
    -- Unique constraint on key_id
    CONSTRAINT unique_key_id UNIQUE (key_id)
);

CREATE INDEX idx_app_keys_tenant_app ON app_keys(tenant_id, app_id);
CREATE INDEX idx_app_keys_key_id ON app_keys(key_id);
CREATE INDEX idx_app_keys_prefix ON app_keys(prefix);
CREATE INDEX idx_app_keys_user_id ON app_keys(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_app_keys_revoked ON app_keys(revoked);
CREATE INDEX idx_app_keys_environment ON app_keys(environment);

COMMENT ON TABLE app_keys IS 'API keys for application authentication';
COMMENT ON COLUMN app_keys.key_id IS 'Public key identifier (visible part of the API key)';
COMMENT ON COLUMN app_keys.prefix IS 'Key prefix for quick identification (e.g., sk_live_)';
COMMENT ON COLUMN app_keys.secret_hash IS 'SHA3-256 hashed secret (never store plain text)';
COMMENT ON COLUMN app_keys.key_type IS 'Key type: secret, public';
COMMENT ON COLUMN app_keys.environment IS 'Environment: live, test';

-- ============================================================
-- Functions and Triggers
-- ============================================================

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_apps_updated_at
    BEFORE UPDATE ON apps
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_branches_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- RBAC Tables (Roles, Permissions, Assignments)
-- ============================================================

-- Roles Table (tenant+app scoped)
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_role_name_per_tenant_app UNIQUE (tenant_id, app_id, name),
    CONSTRAINT fk_roles_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_roles_app FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_roles_tenant_app ON roles(tenant_id, app_id);
CREATE INDEX idx_roles_status ON roles(status);

COMMENT ON TABLE roles IS 'Roles for RBAC (tenant+app scoped)';
COMMENT ON COLUMN roles.status IS 'Role status: active, inactive';

-- User-Role assignments (tenant+app scoped)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    role_id VARCHAR(255) NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP,
    
    PRIMARY KEY (user_id, tenant_id, app_id, role_id),
    CONSTRAINT fk_user_roles_user FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id, tenant_id, app_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_user_roles_active ON user_roles(user_id, tenant_id, app_id) WHERE revoked_at IS NULL;

COMMENT ON TABLE user_roles IS 'User-to-role assignments';

-- Permissions Table (tenant+app scoped)
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    resource VARCHAR(255),
    action VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_permission_name_per_tenant_app UNIQUE (tenant_id, app_id, name),
    CONSTRAINT fk_permissions_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_permissions_app FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_permissions_tenant_app ON permissions(tenant_id, app_id);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
CREATE INDEX idx_permissions_status ON permissions(status);

COMMENT ON TABLE permissions IS 'Permissions for fine-grained access control';
COMMENT ON COLUMN permissions.resource IS 'Resource type (e.g., users, documents)';
COMMENT ON COLUMN permissions.action IS 'Action (e.g., read, write, delete)';
COMMENT ON COLUMN permissions.status IS 'Permission status: active, inactive';

-- Permission Compositions Table (for compound/composite permissions)
CREATE TABLE IF NOT EXISTS permission_compositions (
    parent_permission_id VARCHAR(255) NOT NULL,
    child_permission_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT TRUE,
    priority INT NOT NULL DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (parent_permission_id, child_permission_id),
    CONSTRAINT fk_perm_comp_parent FOREIGN KEY (parent_permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    CONSTRAINT fk_perm_comp_child FOREIGN KEY (child_permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    CONSTRAINT fk_perm_comp_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_perm_comp_app FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE,
    CONSTRAINT no_self_reference CHECK (parent_permission_id != child_permission_id)
);

CREATE INDEX idx_perm_comp_parent ON permission_compositions(parent_permission_id);
CREATE INDEX idx_perm_comp_child ON permission_compositions(child_permission_id);
CREATE INDEX idx_perm_comp_tenant_app ON permission_compositions(tenant_id, app_id);
CREATE INDEX idx_perm_comp_priority ON permission_compositions(parent_permission_id, priority);

COMMENT ON TABLE permission_compositions IS 'Compound permission definitions - maps parent permissions to their constituent child permissions';
COMMENT ON COLUMN permission_compositions.parent_permission_id IS 'The compound/UI permission (e.g., "ui:user_form")';
COMMENT ON COLUMN permission_compositions.child_permission_id IS 'The included permission (e.g., "users:read", "users:create")';
COMMENT ON COLUMN permission_compositions.is_required IS 'Whether this child permission is required or optional';
COMMENT ON COLUMN permission_compositions.priority IS 'Evaluation order (lower number = higher priority)';

-- Role-Permission assignments (tenant+app scoped)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    permission_id VARCHAR(255) NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP,
    
    PRIMARY KEY (role_id, tenant_id, app_id, permission_id),
    CONSTRAINT fk_role_perms_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_perms_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);
CREATE INDEX idx_role_permissions_active ON role_permissions(role_id, tenant_id, app_id) WHERE revoked_at IS NULL;

COMMENT ON TABLE role_permissions IS 'Role-to-permission assignments';

-- User-Permission assignments (direct, tenant+app scoped)
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,
    permission_id VARCHAR(255) NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP,
    
    PRIMARY KEY (user_id, tenant_id, app_id, permission_id),
    CONSTRAINT fk_user_perms_user FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_user_perms_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_permissions_user ON user_permissions(user_id, tenant_id, app_id);
CREATE INDEX idx_user_permissions_permission ON user_permissions(permission_id);
CREATE INDEX idx_user_permissions_active ON user_permissions(user_id, tenant_id, app_id) WHERE revoked_at IS NULL;

COMMENT ON TABLE user_permissions IS 'Direct user-to-permission assignments';

-- Triggers for RBAC tables
CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_permissions_updated_at
    BEFORE UPDATE ON permissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Authorization Policies (Policy-Based Access Control)
-- ============================================================

CREATE TABLE IF NOT EXISTS policies (
    -- Primary identifiers
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    app_id VARCHAR(255) NOT NULL,

    -- Policy details
    name VARCHAR(255) NOT NULL,
    description TEXT,
    effect VARCHAR(20) NOT NULL CHECK (effect IN ('allow', 'deny')),
    
    -- Policy rules (stored as JSONB arrays)
    subjects JSONB NOT NULL DEFAULT '[]'::jsonb,
    resources JSONB NOT NULL DEFAULT '[]'::jsonb,
    actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    conditions JSONB,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    
    -- Metadata
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT unique_policy_name_per_tenant_app UNIQUE (tenant_id, app_id, name),
    CONSTRAINT fk_policies_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_policies_app FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE
);

-- Create indexes for query performance
CREATE INDEX idx_policies_tenant_app ON policies(tenant_id, app_id);
CREATE INDEX idx_policies_status ON policies(status);
CREATE INDEX idx_policies_effect ON policies(effect);
CREATE INDEX idx_policies_created_at ON policies(created_at DESC);

-- GIN indexes for JSONB array searches
CREATE INDEX idx_policies_subjects ON policies USING GIN (subjects);
CREATE INDEX idx_policies_resources ON policies USING GIN (resources);
CREATE INDEX idx_policies_actions ON policies USING GIN (actions);

-- Comments
COMMENT ON TABLE policies IS 'Authorization policies for multi-tenant access control';
COMMENT ON COLUMN policies.effect IS 'Policy effect: allow or deny';
COMMENT ON COLUMN policies.subjects IS 'Subject patterns (user IDs, role IDs, wildcards) - JSONB array';
COMMENT ON COLUMN policies.resources IS 'Resource patterns (resource:id, wildcards) - JSONB array';
COMMENT ON COLUMN policies.actions IS 'Allowed/denied actions (read, write, delete, etc.) - JSONB array';
COMMENT ON COLUMN policies.conditions IS 'ABAC conditions for context-aware authorization';

CREATE TRIGGER update_policies_updated_at
    BEFORE UPDATE ON policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- Tenant Ownership and History
-- ============================================================

CREATE TABLE IF NOT EXISTS tenant_ownership_history (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    previous_owner VARCHAR(255) NOT NULL,
    new_owner VARCHAR(255) NOT NULL,
    transferred_by VARCHAR(255) NOT NULL,
    reason TEXT,
    transferred_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Foreign keys
    CONSTRAINT fk_ownership_tenant 
        FOREIGN KEY (tenant_id) 
        REFERENCES tenants(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_ownership_previous_owner
        FOREIGN KEY (previous_owner)
        REFERENCES users(id)
        ON DELETE RESTRICT,  -- Don't allow deleting users with ownership history
    
    CONSTRAINT fk_ownership_new_owner
        FOREIGN KEY (new_owner)
        REFERENCES users(id)
        ON DELETE RESTRICT,
    
    CONSTRAINT fk_ownership_transferred_by
        FOREIGN KEY (transferred_by)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_ownership_tenant ON tenant_ownership_history(tenant_id, transferred_at DESC);
CREATE INDEX idx_ownership_user ON tenant_ownership_history(new_owner);
CREATE INDEX idx_ownership_date ON tenant_ownership_history(transferred_at DESC);

COMMENT ON TABLE tenant_ownership_history IS 'Audit trail of tenant ownership transfers';
COMMENT ON COLUMN tenant_ownership_history.transferred_by IS 'User who initiated transfer - must be current owner or platform admin';

-- ============================================================
-- Helper Functions for Tenant Ownership
-- ============================================================

-- Function to transfer ownership
CREATE OR REPLACE FUNCTION transfer_tenant_ownership(
    p_tenant_id VARCHAR(255),
    p_new_owner_id VARCHAR(255),
    p_transferred_by VARCHAR(255),
    p_reason TEXT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_old_owner_id VARCHAR(255);
    v_new_owner_tenant VARCHAR(255);
BEGIN
    -- Get current owner
    SELECT owner_id INTO v_old_owner_id
    FROM tenants
    WHERE id = p_tenant_id;
    
    IF v_old_owner_id IS NULL THEN
        RAISE EXCEPTION 'Tenant % not found', p_tenant_id;
    END IF;
    
    -- Validate new owner is in the same tenant
    SELECT tenant_id INTO v_new_owner_tenant
    FROM users
    WHERE id = p_new_owner_id;
    
    IF v_new_owner_tenant IS NULL THEN
        RAISE EXCEPTION 'New owner user % not found', p_new_owner_id;
    END IF;
    
    IF v_new_owner_tenant != p_tenant_id THEN
        RAISE EXCEPTION 'New owner % is not in tenant %', p_new_owner_id, p_tenant_id;
    END IF;
    
    -- Update tenant owner
    UPDATE tenants
    SET owner_id = p_new_owner_id,
        updated_at = NOW()
    WHERE id = p_tenant_id;
    
    -- Update old owner flag
    UPDATE users
    SET is_tenant_owner = FALSE,
        updated_at = NOW()
    WHERE id = v_old_owner_id;
    
    -- Update new owner flag
    UPDATE users
    SET is_tenant_owner = TRUE,
        updated_at = NOW()
    WHERE id = p_new_owner_id;
    
    -- Create ownership history record
    INSERT INTO tenant_ownership_history (
        id,
        tenant_id,
        previous_owner,
        new_owner,
        transferred_by,
        reason,
        transferred_at
    ) VALUES (
        'owner_transfer_' || p_tenant_id || '_' || EXTRACT(EPOCH FROM NOW())::BIGINT,
        p_tenant_id,
        v_old_owner_id,
        p_new_owner_id,
        p_transferred_by,
        COALESCE(p_reason, 'Ownership transferred'),
        NOW()
    );
END;
$$ LANGUAGE plpgsql;

-- Function to check if user is tenant owner
CREATE OR REPLACE FUNCTION is_tenant_owner(p_user_id VARCHAR(255), p_tenant_id VARCHAR(255))
RETURNS BOOLEAN AS $$
DECLARE
    v_is_owner BOOLEAN;
BEGIN
    SELECT owner_id = p_user_id INTO v_is_owner
    FROM tenants
    WHERE id = p_tenant_id;
    
    RETURN COALESCE(v_is_owner, FALSE);
END;
$$ LANGUAGE plpgsql;

-- Function to get tenant owner
CREATE OR REPLACE FUNCTION get_tenant_owner(p_tenant_id VARCHAR(100))
RETURNS TABLE (
    user_id VARCHAR(100),
    username VARCHAR(100),
    email VARCHAR(255),
    full_name VARCHAR(255)
) AS $$
BEGIN
    RETURN QUERY
    SELECT u.id, u.username, u.email, u.full_name
    FROM users u
    JOIN tenants t ON t.owner_id = u.id
    WHERE t.id = p_tenant_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Audit Logs (Centralized Logging for All Transactions)
-- ============================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255),
    app_id VARCHAR(255),
    user_id VARCHAR(255),
    session_id VARCHAR(255),
    
    -- Action details
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    
    -- HTTP context
    method VARCHAR(10),
    path TEXT,
    status_code INT,
    
    -- Request/Response
    request_body JSONB,
    response_body JSONB,
    
    -- Meta information
    ip_address INET,
    user_agent TEXT,
    source VARCHAR(50),
    
    -- Result
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    
    -- Additional context
    metadata JSONB,
    
    -- Timestamp
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Foreign keys (nullable for system-level actions)
    CONSTRAINT fk_audit_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_audit_app FOREIGN KEY (tenant_id, app_id) REFERENCES apps(tenant_id, id) ON DELETE CASCADE
);

-- Indexes for efficient querying
CREATE INDEX idx_audit_logs_tenant_app ON audit_logs(tenant_id, app_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_session ON audit_logs(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_audit_logs_success ON audit_logs(success, created_at DESC);
CREATE INDEX idx_audit_logs_source ON audit_logs(source);

-- GIN indexes for JSONB searches
CREATE INDEX idx_audit_logs_metadata ON audit_logs USING GIN (metadata);
CREATE INDEX idx_audit_logs_request ON audit_logs USING GIN (request_body);

-- Partitioning preparation comment (for future scaling)
COMMENT ON TABLE audit_logs IS 'Centralized audit log for all system actions. Consider partitioning by created_at for large datasets.';
COMMENT ON COLUMN audit_logs.action IS 'Action performed (e.g., login, create, update, delete, authorize, etc.)';
COMMENT ON COLUMN audit_logs.resource_type IS 'Type of resource (e.g., user, tenant, app, role, permission, policy, etc.)';
COMMENT ON COLUMN audit_logs.resource_id IS 'ID of the affected resource';
COMMENT ON COLUMN audit_logs.source IS 'Source of the action (e.g., api, web, mobile, system, cron, etc.)';
COMMENT ON COLUMN audit_logs.session_id IS 'Session/request ID for tracing';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional context-specific data (changes, filters, etc.)';

-- Function to automatically clean old audit logs (optional)
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(days_to_keep INT DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM audit_logs
    WHERE created_at < NOW() - (days_to_keep || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_audit_logs IS 'Delete audit logs older than specified days (default: 90 days)';


-- Trigger function to validate owner is in same tenant
CREATE OR REPLACE FUNCTION validate_tenant_owner()
RETURNS TRIGGER AS $$
DECLARE
    owner_tenant_id VARCHAR(100);
BEGIN
    -- Get the tenant_id of the owner user
    SELECT tenant_id INTO owner_tenant_id
    FROM users
    WHERE id = NEW.owner_id;
    
    -- Validate owner exists and is in the same tenant
    IF owner_tenant_id IS NULL THEN
        RAISE EXCEPTION 'Owner user % does not exist', NEW.owner_id;
    END IF;
    
    IF owner_tenant_id != NEW.id THEN
        RAISE EXCEPTION 'Owner user % is not in tenant % (belongs to %)', 
            NEW.owner_id, NEW.id, owner_tenant_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on INSERT and UPDATE
CREATE TRIGGER trg_validate_tenant_owner
    BEFORE INSERT OR UPDATE OF owner_id ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION validate_tenant_owner();

-- ============================================================
-- Sample Data (Optional - for development/testing)
-- ============================================================

-- Uncomment the following lines to insert sample data:

/*
-- Sample tenant
INSERT INTO tenants (id, name, domain, db_dsn, db_schema, status, created_at, updated_at)
VALUES (
    'demo-tenant',
    'Demo Tenant Inc.',
    'demo.example.com',
    'host=localhost port=5432 user=postgres password=postgres dbname=lokstra_core sslmode=disable',
    'demo_tenant',
    'active',
    NOW(),
    NOW()
);

-- Sample app
INSERT INTO apps (id, tenant_id, name, type, status, created_at, updated_at)
VALUES (
    'web-portal',
    'demo-tenant',
    'Web Portal',
    'web',
    'active',
    NOW(),
    NOW()
);

-- Sample branch
INSERT INTO branches (id, tenant_id, app_id, name, type, status, created_at, updated_at)
VALUES (
    'hq-branch',
    'demo-tenant',
    'web-portal',
    'Headquarters',
    'headquarters',
    'active',
    NOW(),
    NOW()
);

-- Sample user
INSERT INTO users (id, tenant_id, username, email, full_name, status, created_at, updated_at)
VALUES (
    'user-001',
    'demo-tenant',
    'admin',
    'admin@demo.example.com',
    'Admin User',
    'active',
    NOW(),
    NOW()
);

-- Grant user access to app
INSERT INTO user_apps (tenant_id, app_id, user_id, status, granted_at)
VALUES (
    'demo-tenant',
    'web-portal',
    'user-001',
    'active',
    NOW()
);
*/

-- ============================================================
-- End of Schema
-- ============================================================

-- Show table sizes
SELECT 
    schemaname as schema,
    tablename as table,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
