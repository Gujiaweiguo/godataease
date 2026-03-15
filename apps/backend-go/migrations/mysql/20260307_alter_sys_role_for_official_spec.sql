-- Migration: Alter sys_role table to align with official DataEase specification
-- Description: Add role level (system/org), organization association, builtin flag, and readonly flag
-- Reference: https://dataease.io/docs/v2/xpack/user_management_user/#2

-- Add new columns to sys_role table
ALTER TABLE sys_role
ADD COLUMN org_id BIGINT NULL COMMENT '组织ID，系统级角色为NULL' AFTER role_code,
ADD COLUMN role_type ENUM('system', 'org') NOT NULL DEFAULT 'org' COMMENT '角色类型：system-系统级，org-组织级' AFTER org_id,
ADD COLUMN is_builtin TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否内置角色：0-否，1-是' AFTER role_type,
ADD COLUMN readonly TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否只读：0-否，1-是（不可编辑删除）' AFTER is_builtin;

-- Add index for org_id
CREATE INDEX idx_org_id ON sys_role(org_id);

-- Add index for role_type
CREATE INDEX idx_role_type ON sys_role(role_type);

-- Update existing admin role to system level
UPDATE sys_role 
SET 
    role_type = 'system',
    is_builtin = 1,
    readonly = 1,
    org_id = NULL
WHERE role_code = 'admin';

-- Update existing user role to org level builtin
UPDATE sys_role 
SET 
    role_type = 'org',
    is_builtin = 1,
    readonly = 0,
    org_id = 1
WHERE role_code = 'user';

-- Insert org_admin role if not exists
INSERT INTO sys_role (role_name, role_code, role_desc, org_id, role_type, is_builtin, readonly, parent_id, level, data_scope, status, create_by, create_time)
SELECT '组织管理员', 'org_admin', '组织管理员，管理当前组织', 1, 'org', 1, 0, NULL, 2, 'dept_and_child', 1, 'system', NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_role WHERE role_code = 'org_admin');

-- Add constraint: system roles must have NULL org_id
ALTER TABLE sys_role
ADD CONSTRAINT chk_system_role_org 
CHECK (NOT (role_type = 'system' AND org_id IS NOT NULL));

-- Add constraint: org roles must have non-NULL org_id
ALTER TABLE sys_role
ADD CONSTRAINT chk_org_role_org
CHECK (NOT (role_type = 'org' AND org_id IS NULL));
