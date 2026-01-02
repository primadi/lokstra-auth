-- ============================================================
-- Tenant Creation with Auto-Admin
-- ============================================================
-- This function automatically creates tenant admin when tenant is created
-- ============================================================

CREATE OR REPLACE FUNCTION create_tenant_with_admin(
    p_tenant_id VARCHAR(255),
    p_tenant_name VARCHAR(255),
    p_tenant_domain VARCHAR(255),
    p_db_dsn TEXT,
    p_db_schema VARCHAR(255),
    p_admin_username VARCHAR(255),
    p_admin_email VARCHAR(255),
    p_admin_password_hash TEXT,
    p_admin_full_name VARCHAR(255) DEFAULT NULL
) RETURNS TABLE (
    tenant_id VARCHAR(255),
    tenant_name VARCHAR(255),
    admin_user_id VARCHAR(255),
    admin_username VARCHAR(255),
    default_app_id VARCHAR(255)
) AS $$
DECLARE
    v_user_id VARCHAR(255);
    v_app_id VARCHAR(255);
BEGIN
    -- 1. Create tenant
    INSERT INTO tenants (
        id, name, domain, db_dsn, db_schema, 
        status, created_at, updated_at
    ) VALUES (
        p_tenant_id, p_tenant_name, p_tenant_domain, p_db_dsn, p_db_schema,
        'active', NOW(), NOW()
    );

    -- 2. Create default app for tenant
    v_app_id := p_tenant_id || '-admin';
    INSERT INTO apps (
        id, tenant_id, name, type, status, created_at, updated_at
    ) VALUES (
        v_app_id, p_tenant_id, p_tenant_name || ' Admin', 'admin', 'active', NOW(), NOW()
    );

    -- 3. Create tenant admin user
    v_user_id := p_tenant_id || '-admin-user';
    INSERT INTO users (
        id, tenant_id, username, email, full_name, password_hash,
        status, metadata, created_at, updated_at
    ) VALUES (
        v_user_id, p_tenant_id, p_admin_username, p_admin_email, 
        COALESCE(p_admin_full_name, 'Tenant Administrator'),
        p_admin_password_hash, 'active',
        '{"is_tenant_owner": true, "created_via": "auto_admin"}'::jsonb,
        NOW(), NOW()
    );

    -- 4. Grant user access to admin app
    INSERT INTO user_apps (tenant_id, app_id, user_id, status, granted_at)
    VALUES (p_tenant_id, v_app_id, v_user_id, 'active', NOW());

    -- 5. Create tenant-admin role if not exists
    INSERT INTO roles (tenant_id, id, name, description, created_at)
    VALUES (p_tenant_id, 'tenant-admin', 'Tenant Administrator', 'Full tenant access', NOW())
    ON CONFLICT (tenant_id, id) DO NOTHING;

    -- 6. Assign tenant-admin role to user
    INSERT INTO user_roles (tenant_id, user_id, role_id, granted_at)
    VALUES (p_tenant_id, v_user_id, 'tenant-admin', NOW());

    -- Return created entities
    RETURN QUERY
    SELECT 
        p_tenant_id, 
        p_tenant_name, 
        v_user_id, 
        p_admin_username, 
        v_app_id;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Usage Example:
-- ============================================================
/*
SELECT * FROM create_tenant_with_admin(
    'acme-corp',                    -- tenant_id
    'Acme Corporation',             -- tenant_name
    'acme-corp.com',                -- domain
    'postgres://...',               -- db_dsn
    'acme_corp',                    -- db_schema
    'admin',                        -- admin_username
    'admin@acme-corp.com',          -- admin_email
    'hashed-password-here',         -- admin_password_hash
    'John Doe'                      -- admin_full_name (optional)
);
*/
