-- Migration: Refactor system management menu structure
-- Description: 重组系统管理菜单结构，-- Date: 2026-03-20

-- 1. 创建「可视化」分组菜单
INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (100, 0, 1, 'commons.visualization', NULL, 2, 'chart', '/visualization', 0, 1, 0, 'group')
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    menu_sort = VALUES(menu_sort),
    icon = VALUES(icon),
    hidden = VALUES(hidden);

-- 2. 移动仪表板到可视化下
UPDATE core_menu SET pid = 100, menu_sort = 1 WHERE id = 2 AND path = '/panel';

-- 3. 移动数据大屏到可视化下
UPDATE core_menu SET pid = 100, menu_sort = 2 WHERE id = 3 AND path = '/screen';

-- 4. 移动模板管理到可视化下
UPDATE core_menu SET pid = 100, menu_sort = 3 WHERE id = 31 AND path = '/template-setting';

-- 5. 隐藏工具箱菜单（保留但不再独立显示）
UPDATE core_menu SET hidden = 1 WHERE id = 30 AND path = '/toolbox';

-- 6. 调整菜单排序
-- 工作台排序为 1
UPDATE core_menu SET menu_sort = 1 WHERE id = 1 AND path = '/workbranch';
-- 可视化排序为 2（已设置）
-- 数据管理排序为 3
UPDATE core_menu SET menu_sort = 3 WHERE id = 4 AND path = '/data';
-- 系统管理排序为 4（需要先创建或确认 system 分组）

-- 7. 创建「系统管理」分组菜单（如果不存在）
INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (101, 0, 1, 'commons.system_setting', NULL, 4, 'setting', '/system', 0, 1, 0, 'group')
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    menu_sort = VALUES(menu_sort),
    icon = VALUES(icon),
    hidden = VALUES(hidden);

-- 8. 移动用户管理到系统管理下
-- 注意：用户管理可能已存在于其他位置，INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (102, 101, 2, 'system.user_management', 'system/user/index', 1, 'user', '/system/user', 0, 1, 1, 'route')
ON DUPLICATE KEY UPDATE 
    pid = VALUES(pid),
    menu_sort = VALUES(menu_sort);

-- 9. 移动组织管理到系统管理下
INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (103, 101, 2, 'system.org_management', 'system/org/index', 2, 'org', '/system/org', 0, 1, 1, 'route')
ON DUPLICATE KEY UPDATE 
    pid = VALUES(pid),
    menu_sort = VALUES(menu_sort);

-- 10. 移动角色管理到系统管理下（将被用户管理合并，INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (104, 101, 2, 'system.role_management', 'system/role/index', 3, 'role', '/system/role', 0, 1, 1, 'route')
ON DUPLICATE KEY UPDATE 
    pid = VALUES(pid),
    menu_sort = VALUES(menu_sort);

-- 11. 创建「权限配置」菜单
INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (105, 101, 2, 'system.permission_config', 'system/permission/index', 4, 'lock', '/system/permission', 0, 1, 1, 'route')
ON DUPLICATE KEY UPDATE 
    pid = VALUES(pid),
    menu_sort = VALUES(menu_sort);

-- 12. 创建「菜单配置」菜单
INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (106, 101, 2, 'system.menu_config', 'system/menu/index', 5, 'menu', '/system/menu', 0, 1, 1, 'route')
ON DUPLICATE KEY UPDATE 
    pid = VALUES(pid),
    menu_sort = VALUES(menu_sort);

-- 13. 移动系统参数到系统管理下
UPDATE core_menu SET pid = 101, menu_sort = 6 WHERE id = 16 AND path = '/parameter';

-- 14. 创建「字体管理」菜单
INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (107, 101, 2, 'system.font_management', 'system/font/index', 7, 'font', '/system/font', 0, 1, 1, 'route')
ON DUPLICATE KEY UPDATE 
    pid = VALUES(pid),
    menu_sort = VALUES(menu_sort);

-- 15. 隐藏原有的 sys-setting 分组（已整合到系统管理）
UPDATE core_menu SET hidden = 1 WHERE id = 15 AND path = '/sys-setting';

-- 16. 删除重复的菜单权限配置菜单（如果存在）
DELETE FROM core_menu WHERE path = '/system/menu-permission';

-- 17. 为新菜单授权给管理员角色
INSERT INTO sys_role_menu (role_id, menu_id, created_at, updated_at)
SELECT 
    1 AS role_id,
    m.id AS menu_id,
    NOW() AS created_at,
    NOW() AS updated_at
FROM core_menu m
WHERE m.id IN (100, 101, 102, 103, 104, 105, 106, 107)
AND NOT EXISTS (
    SELECT 1 FROM sys_role_menu rm 
    WHERE rm.role_id = 1 AND rm.menu_id = m.id
);

-- 18. 同步菜单权限给管理员（如果已存在的菜单移动了位置）
INSERT INTO sys_role_menu (role_id, menu_id, created_at, updated_at)
SELECT 
    1 AS role_id,
    m.id AS menu_id,
    NOW() AS created_at,
    NOW() AS updated_at
FROM core_menu m
WHERE m.id IN (2, 3, 31)  -- panel, screen, template-setting
AND NOT EXISTS (
    SELECT 1 FROM sys_role_menu rm 
    WHERE rm.role_id = 1 AND rm.menu_id = m.id
);
