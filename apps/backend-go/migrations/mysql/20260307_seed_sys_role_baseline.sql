-- Migration: Seed default system roles (admin and user)
-- Description: Initialize default roles for system initialization
-- - admin: System administrator with full permissions (role_code='admin')
-- - user: Normal user with basic permissions (role_code='user')

-- Insert admin role if not exists
INSERT INTO sys_role (role_name, role_code, role_desc, parent_id, level, data_scope, status, create_by, create_time)
SELECT '系统管理员', 'admin', '系统管理员，拥有所有权限', NULL, 1, 'all', 1, 'system', NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_role WHERE role_code = 'admin');

-- Insert user role if not exists
INSERT INTO sys_role (role_name, role_code, role_desc, parent_id, level, data_scope, status, create_by, create_time)
SELECT '普通用户', 'user', '普通用户，拥有基本权限', NULL, 1, 'self', 1, 'system', NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_role WHERE role_code = 'user');
