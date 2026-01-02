-- ============================================================
-- Platform Bootstrap - Create Platform Admin
-- ============================================================
-- This script creates the initial platform admin user
-- Run this ONCE during initial deployment
-- ============================================================

-- Create "platform" tenant for platform-level operations
INSERT INTO tenants (
    id, 
    name, 
    domain, 
    db_dsn, 
    db_schema, 
    status, 
    settings,
    metadata,
    created_at, 
    updated_at
) VALUES (
    'platform',
    'Platform Admin',
    'platform.internal',
    'postgres://postgres:adm1n@localhost:5432/lokstra_db?sslmode=disable',
    'platform',
    'active',
    '{"is_platform": true, "max_users": 10, "max_apps": 1}'::jsonb,
    '{"type": "platform", "description": "Internal platform administration tenant"}'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Create "platform-admin-app" for platform operations
INSERT INTO apps (
    id,
    tenant_id,
    name,
    type,
    status,
    config,
    created_at,
    updated_at
) VALUES (
    'platform-admin-app',
    'platform',
    'Platform Administration',
    'admin',
    'active',
    '{"is_platform_app": true}'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, id) DO NOTHING;

-- Create platform admin user
-- Username: platform-admin
-- Password: PlatformAdmin@2025! (SHA3-256 hash below)
-- CHANGE THIS PASSWORD IMMEDIATELY AFTER FIRST LOGIN!
INSERT INTO users (
    id,
    tenant_id,
    username,
    email,
    full_name,
    password_hash,
    status,
    metadata,
    created_at,
    updated_at
) VALUES (
    'platform-admin-user',
    'platform',
    'platform-admin',
    'admin@platform.internal',
    'Platform Administrator',
    -- Password: PlatformAdmin@2025!
    -- Generate using: echo -n 'PlatformAdmin@2025!' | sha3sum -a 256
    -- Or use your password hashing function
    'change-me-use-proper-hash',
    'active',
    '{"is_platform_admin": true, "force_password_change": true}'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, id) DO NOTHING;

-- Grant platform admin access to platform-admin-app
INSERT INTO user_apps (
    tenant_id,
    app_id,
    user_id,
    status,
    granted_at
) VALUES (
    'platform',
    'platform-admin-app',
    'platform-admin-user',
    'active',
    NOW()
) ON CONFLICT (tenant_id, app_id, user_id) DO NOTHING;

-- ============================================================
-- Platform Permissions & Roles (for RBAC)
-- ============================================================

-- Platform-level permissions (app-scoped)
INSERT INTO permissions (id, tenant_id, app_id, name, resource, action, description, created_at, updated_at) VALUES
('perm_tenant_create', 'platform', 'platform-admin-app', 'tenant:create', 'tenant', 'create', 'Create new tenants', NOW(), NOW()),
('perm_tenant_read', 'platform', 'platform-admin-app', 'tenant:read', 'tenant', 'read', 'View tenant details', NOW(), NOW()),
('perm_tenant_update', 'platform', 'platform-admin-app', 'tenant:update', 'tenant', 'update', 'Modify tenant settings', NOW(), NOW()),
('perm_tenant_delete', 'platform', 'platform-admin-app', 'tenant:delete', 'tenant', 'delete', 'Remove tenants', NOW(), NOW()),
('perm_tenant_list', 'platform', 'platform-admin-app', 'tenant:list', 'tenant', 'list', 'List all tenants', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Platform Admin Role
INSERT INTO roles (id, tenant_id, app_id, name, description, created_at, updated_at) VALUES
('role_platform_admin', 'platform', 'platform-admin-app', 'Platform Administrator', 'Full platform access', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Assign all tenant permissions to platform-admin role
INSERT INTO role_permissions (role_id, tenant_id, app_id, permission_id, granted_at)
SELECT 'role_platform_admin', 'platform', 'platform-admin-app', id, NOW()
FROM permissions
WHERE tenant_id = 'platform' AND app_id = 'platform-admin-app' AND resource = 'tenant'
ON CONFLICT (role_id, tenant_id, app_id, permission_id) DO NOTHING;

-- Assign platform-admin role to platform-admin user
INSERT INTO user_roles (user_id, tenant_id, app_id, role_id, granted_at) VALUES
('platform-admin-user', 'platform', 'platform-admin-app', 'role_platform_admin', NOW())
ON CONFLICT (user_id, tenant_id, app_id, role_id) DO NOTHING;

-- ============================================================
-- Verification Queries
-- ============================================================

-- Verify platform tenant created
SELECT 'Platform tenant created:' as message, id, name, status FROM tenants WHERE id = 'platform';

-- Verify platform admin user created
SELECT 'Platform admin user created:' as message, id, username, email, status FROM users WHERE tenant_id = 'platform';

-- Verify platform admin has platform-admin role
SELECT 'Platform admin roles:' as message, r.name 
FROM user_roles ur 
JOIN roles r ON ur.role_id = r.id
WHERE ur.tenant_id = 'platform' AND ur.app_id = 'platform-admin-app' AND ur.user_id = 'platform-admin-user';

-- Display login credentials (REMOVE IN PRODUCTION)
SELECT '============================================================' as separator;
SELECT 'PLATFORM ADMIN CREDENTIALS (CHANGE IMMEDIATELY!)' as notice;
SELECT '============================================================' as separator;
SELECT 'Username: platform-admin' as username;
SELECT 'Password: PlatformAdmin@2025!' as password;
SELECT 'Tenant: platform' as tenant;
SELECT 'App: platform-admin-app' as app;
SELECT '============================================================' as separator;
SELECT 'IMPORTANT: Change password after first login!' as warning;
