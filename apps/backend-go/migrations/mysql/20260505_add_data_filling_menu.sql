-- Migration: Add DataFilling menu entry under data group
-- Description: 为数据准备分组新增数据填报菜单，并授权给管理员角色
-- Date: 2026-05-05

-- ============================================================
-- 1. 创建 DataFilling 菜单
-- ============================================================
INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (300, 4, 2, 'data-filling', 'data-filling/manage/index', 5, 'form', 'data-filling', 0, 1, 1, 'route')
ON DUPLICATE KEY UPDATE
    pid = VALUES(pid),
    name = VALUES(name),
    component = VALUES(component),
    menu_sort = VALUES(menu_sort),
    icon = VALUES(icon),
    path = VALUES(path),
    hidden = VALUES(hidden),
    in_layout = VALUES(in_layout),
    auth = VALUES(auth),
    menu_type = VALUES(menu_type);

-- ============================================================
-- 2. 为管理员角色授权 DataFilling 菜单
-- ============================================================
INSERT INTO sys_role_menu (role_id, menu_id, created_at, updated_at)
SELECT 1, 300, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_role_menu WHERE role_id = 1 AND menu_id = 300
);
