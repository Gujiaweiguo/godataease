-- Migration: Seed admin menu authorization
-- Description: 为 bootstrap 管理员角色授权所有 core_menu 记录
-- Date: 2026-02-22
-- Bootstrap admin identification: role_code='admin' first, fallback to role_id=1

-- Insert all menu permissions for admin role
-- Using INSERT IGNORE to safely handle re-runs
INSERT INTO sys_role_menu (role_id, menu_id, created_at, updated_at)
SELECT
    COALESCE(
        (SELECT role_id FROM sys_role WHERE role_code = 'admin' LIMIT 1),
        1
    ) AS role_id,
    m.id AS menu_id,
    NOW() AS created_at,
    NOW() AS updated_at
FROM core_menu m
WHERE m.id NOT IN (
    SELECT rm.menu_id FROM sys_role_menu rm
    WHERE rm.role_id = COALESCE(
        (SELECT role_id FROM sys_role WHERE role_code = 'admin' LIMIT 1),
        1
    )
);
