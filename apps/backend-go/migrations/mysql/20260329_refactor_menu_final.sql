-- Migration: Refactor navigation menu to six first-level groups
-- Description: 重组导航菜单结构，拆分系统管理为组织权限+系统设置，升级工具箱为一级菜单，移除帮助链接入口
-- Date: 2026-03-29
--
-- Prerequisites: Database must have the following IDs:
--   system(71), user(72), role(73), org(74), permission(75),
--   audit(76), audit-dashboard(77), audit-settings(78),
--   menu(79), export-center(80), help(81-85), mine-sys-setting(89), toolbox(30)

-- ============================================================
-- 1. 创建「组织权限」一级分组菜单 (id=200)
-- ============================================================
INSERT INTO core_menu (id, pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (200, 0, 1, 'commons.org_permission', NULL, 4, 'peoples', '/org-permission', 0, 1, 1, 'group')
ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    menu_sort = VALUES(menu_sort),
    icon = VALUES(icon),
    hidden = VALUES(hidden);

-- ============================================================
-- 2. 移动用户管理/组织管理/角色管理/权限管理到组织权限(200)下
-- ============================================================
UPDATE core_menu SET pid = 200, menu_sort = 1 WHERE id = 72;  -- 用户管理
UPDATE core_menu SET pid = 200, menu_sort = 2 WHERE id = 74;  -- 组织管理
UPDATE core_menu SET pid = 200, menu_sort = 3, hidden = 0 WHERE id = 73;  -- 角色管理（取消隐藏）
UPDATE core_menu SET pid = 200, menu_sort = 4 WHERE id = 75;  -- 权限管理

-- ============================================================
-- 3. 系统设置(71)：重新排序剩余子菜单
--    保留: menu(79), audit(76), audit-dashboard(77), audit-settings(78)
-- ============================================================
UPDATE core_menu SET menu_sort = 1 WHERE id = 79;  -- 菜单管理
UPDATE core_menu SET menu_sort = 2 WHERE id = 76;  -- 审计
UPDATE core_menu SET menu_sort = 3 WHERE id = 77;  -- 审计仪表板
UPDATE core_menu SET menu_sort = 4 WHERE id = 78;  -- 审计设置

-- ============================================================
-- 4. 工具箱(30)升级为一级菜单（取消隐藏，调整排序）
-- ============================================================
UPDATE core_menu SET hidden = 0, menu_sort = 6, in_layout = 1 WHERE id = 30;

-- ============================================================
-- 5. 数据导出中心(80)从 data(4) 移到工具箱(30)下
-- ============================================================
UPDATE core_menu SET pid = 30 WHERE id = 80;

-- ============================================================
-- 6. 删除帮助文档分组(81)及子项(82-85)
-- ============================================================
DELETE FROM sys_role_menu WHERE menu_id IN (81, 82, 83, 84, 85);
DELETE FROM core_menu WHERE id IN (82, 83, 84, 85);
DELETE FROM core_menu WHERE id = 81;

-- ============================================================
-- 7. 删除个人菜单中的系统设置快捷入口(89)
-- ============================================================
DELETE FROM sys_role_menu WHERE menu_id = 89;
DELETE FROM core_menu WHERE id = 89;

-- ============================================================
-- 8. 为管理员角色授权新的「组织权限」菜单
-- ============================================================
INSERT INTO sys_role_menu (role_id, menu_id, created_at, updated_at)
SELECT 1, 200, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM sys_role_menu
    WHERE role_id = 1 AND menu_id = 200
);
