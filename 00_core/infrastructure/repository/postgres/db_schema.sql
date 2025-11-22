-- ============================================================
-- Lokstra Auth - PostgreSQL Database Schema
-- ============================================================
-- This script creates the necessary tables for the lokstra-auth
-- core entities: Tenants, Apps, Branches, Users, User-App Access, and App Keys
-- ============================================================

CREATE SCHEMA IF NOT EXISTS lokstra_auth;

SET SEARCH_PATH TO lokstra_auth;

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

COMMENT ON TABLE tenants IS 'Multi-tenant organizations/companies';
COMMENT ON COLUMN tenants.id IS 'Unique tenant identifier (e.g., acme-corp)';
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
    code VARCHAR(100) NOT NULL,
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
        REFERENCES apps(tenant_id, id) ON DELETE CASCADE,
    
    -- Unique constraint on code per app
    CONSTRAINT unique_branch_code_per_app UNIQUE (tenant_id, app_id, code)
);

CREATE INDEX idx_branches_tenant_app ON branches(tenant_id, app_id);
CREATE INDEX idx_branches_type ON branches(type);
CREATE INDEX idx_branches_status ON branches(status);
CREATE INDEX idx_branches_code ON branches(code);

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
    status VARCHAR(50) NOT NULL DEFAULT 'active',
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

COMMENT ON TABLE users IS 'Users within a tenant';
COMMENT ON COLUMN users.status IS 'User status: active, suspended, deleted';

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

COMMENT ON TABLE user_apps IS 'User access control to apps (not authorization - see 04_authz)';
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
    'host=localhost port=5432 user=postgres password=postgres dbname=lokstra_auth sslmode=disable',
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
INSERT INTO branches (id, tenant_id, app_id, code, name, type, status, created_at, updated_at)
VALUES (
    'hq-branch',
    'demo-tenant',
    'web-portal',
    'HQ-001',
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
