-- Migration: Add menu permission configuration menu
-- Description: 添加菜单权限配置菜单项到系统管理
-- Date: 2026-03-08

-- 1. 查找系统管理的菜单ID
SET @system_pid = (SELECT id FROM core_menu WHERE path = '/system' LIMIT 1);

-- 2. 添加菜单权限配置菜单
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (
    COALESCE(@system_pid, 0),
    1,
    '菜单权限配置',
    'system/menu-permission/index',
    45,
    'lock',
    '/system/menu-permission',
    0,
    1,
    1,
    'route'
);

-- 3. 获取新插入的菜单ID
SET @menu_perm_id = LAST_INSERT_ID();

-- 4. 为管理员角色授权新菜单
INSERT INTO sys_role_menu (role_id, menu_id, created_at, updated_at)
SELECT 
    COALESCE(
        (SELECT role_id FROM sys_role WHERE role_code = 'admin' LIMIT 1),
        1
    ) AS role_id,
    @menu_perm_id AS menu_id,
    NOW() AS created_at,
    NOW() AS updated_at
WHERE @menu_perm_id > 0;
