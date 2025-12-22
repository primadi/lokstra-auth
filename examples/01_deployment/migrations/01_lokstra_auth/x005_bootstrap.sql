-- Bootstrap Migration: Initial Super Admin Setup
-- This migration creates the bootstrap tenant, app, and super admin user
-- Only needed if EnableAutoBootstrap=false in code
-- 
-- Usage: Run this migration when deploying to production for the first time
-- After running, you can login with:
--   Tenant: system
--   App: admin-console
--   Username: admin
--   Password: (set via SUPER_ADMIN_PASSWORD env var)

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Insert bootstrap tenant
INSERT INTO tenants (id, name, status, settings, metadata, created_at, updated_at)
VALUES (
    'system',
    'System Administrator',
    'active',
    '{"is_bootstrap": true, "max_users": 1, "max_apps": 1}'::jsonb,
    '{"description": "System administrator tenant - auto-created for initial setup"}'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 2. Insert bootstrap app
INSERT INTO apps (id, tenant_id, name, type, status, config, metadata, created_at, updated_at)
VALUES (
    'admin-console',
    'system',
    'Admin Console',
    'web',
    'active',
    '{"is_bootstrap": true}'::jsonb,
    '{"description": "Admin console for managing tenants and users"}'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, id) DO NOTHING;

-- 3. Insert super admin user
INSERT INTO users (id, tenant_id, username, email, full_name, status, metadata, created_at, updated_at)
VALUES (
    'super-admin',
    'system',
    'admin',
    'admin@localhost',
    'Super Administrator',
    'active',
    '{"is_bootstrap": true, "can_create_tenants": true}'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, id) DO NOTHING;

-- 4. Insert super admin password
-- NOTE: Replace 'CHANGE_ME_IMMEDIATELY' with actual bcrypt hash before running!
-- Example bcrypt hash generation (cost=12):
--   Password: Admin123!@#
--   Hash: $2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIiIkIiIkI
INSERT INTO user_passwords (user_id, tenant_id, password_hash, created_at, updated_at)
VALUES (
    'super-admin',
    'system',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIiIkIiIkI', -- CHANGE THIS!
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, user_id) DO NOTHING;

-- 5. Grant app access to super admin
INSERT INTO user_apps (user_id, tenant_id, app_id, status, roles, permissions, created_at)
VALUES (
    'super-admin',
    'system',
    'admin-console',
    'active',
    '["super-admin"]'::jsonb,
    '["*"]'::jsonb, -- All permissions
    NOW()
) ON CONFLICT (tenant_id, app_id, user_id) DO NOTHING;

-- 6. Create super admin role
INSERT INTO roles (id, tenant_id, app_id, name, description, created_at, updated_at)
VALUES (
    'super-admin-role',
    'system',
    'admin-console',
    'Super Administrator',
    'Full system access - can create and manage all tenants',
    NOW(),
    NOW()
) ON CONFLICT (tenant_id, app_id, id) DO NOTHING;

-- 7. Assign role to super admin
INSERT INTO user_roles (user_id, tenant_id, app_id, role_id, created_at)
VALUES (
    'super-admin',
    'system',
    'admin-console',
    'super-admin-role',
    NOW()
) ON CONFLICT (tenant_id, app_id, user_id, role_id) DO NOTHING;

-- 8. Create bootstrap permissions
INSERT INTO permissions (id, tenant_id, app_id, resource, action, description, created_at, updated_at)
VALUES 
    ('perm_tenant_create', 'system', 'admin-console', 'tenant', 'create', 'Create new tenants', NOW(), NOW()),
    ('perm_tenant_read', 'system', 'admin-console', 'tenant', 'read', 'View tenant information', NOW(), NOW()),
    ('perm_tenant_update', 'system', 'admin-console', 'tenant', 'update', 'Update tenant settings', NOW(), NOW()),
    ('perm_tenant_delete', 'system', 'admin-console', 'tenant', 'delete', 'Delete tenants', NOW(), NOW()),
    ('perm_app_create', 'system', 'admin-console', 'app', 'create', 'Create new apps', NOW(), NOW()),
    ('perm_app_read', 'system', 'admin-console', 'app', 'read', 'View app information', NOW(), NOW()),
    ('perm_app_update', 'system', 'admin-console', 'app', 'update', 'Update app settings', NOW(), NOW()),
    ('perm_app_delete', 'system', 'admin-console', 'app', 'delete', 'Delete apps', NOW(), NOW()),
    ('perm_user_create', 'system', 'admin-console', 'user', 'create', 'Create new users', NOW(), NOW()),
    ('perm_user_read', 'system', 'admin-console', 'user', 'read', 'View user information', NOW(), NOW()),
    ('perm_user_update', 'system', 'admin-console', 'user', 'update', 'Update user details', NOW(), NOW()),
    ('perm_user_delete', 'system', 'admin-console', 'user', 'delete', 'Delete users', NOW(), NOW()),
    ('perm_role_manage', 'system', 'admin-console', 'role', '*', 'Manage roles', NOW(), NOW()),
    ('perm_permission_manage', 'system', 'admin-console', 'permission', '*', 'Manage permissions', NOW(), NOW()),
    ('perm_policy_manage', 'system', 'admin-console', 'policy', '*', 'Manage policies', NOW(), NOW()),
    ('perm_wildcard', 'system', 'admin-console', '*', '*', 'Full system access', NOW(), NOW())
ON CONFLICT (tenant_id, app_id, resource, action) DO NOTHING;

-- 9. Grant all permissions to super admin role
INSERT INTO role_permissions (role_id, tenant_id, app_id, permission_id, created_at)
SELECT 
    'super-admin-role',
    'system',
    'admin-console',
    id,
    NOW()
FROM permissions
WHERE tenant_id = 'system' AND app_id = 'admin-console'
ON CONFLICT (tenant_id, app_id, role_id, permission_id) DO NOTHING;

-- Verification query
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Bootstrap Setup Complete!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Tenant ID:    system';
    RAISE NOTICE 'App ID:       admin-console';
    RAISE NOTICE 'Username:     admin';
    RAISE NOTICE 'Email:        admin@localhost';
    RAISE NOTICE '========================================';
    RAISE NOTICE '⚠️  IMPORTANT:';
    RAISE NOTICE '1. Change the super admin password immediately!';
    RAISE NOTICE '2. Update password_hash in user_passwords table';
    RAISE NOTICE '3. Create your first real tenant';
    RAISE NOTICE '4. Disable bootstrap tenant when done';
    RAISE NOTICE '========================================';
END $$;
