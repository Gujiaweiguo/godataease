-- Migration: Extend menu system for dynamic menus
-- Description: 扩展菜单系统支持动态菜单（外部链接、事件触发等）
-- Date: 2026-03-08

-- 1. 添加新字段到 core_menu 表
ALTER TABLE core_menu 
ADD COLUMN menu_type VARCHAR(20) DEFAULT 'route' COMMENT '菜单类型: route=路由菜单, event=事件触发, external=外部链接, group=分组' AFTER auth,
ADD COLUMN action_config JSON COMMENT '动作配置: {"event":"xxx"} 或 {"url":"https://..."}' AFTER menu_type;

-- 2. 更新现有菜单的 menu_type
UPDATE core_menu SET menu_type = 'route' WHERE menu_type IS NULL OR menu_type = '';

-- 3. 添加新的动态菜单项

-- 3.1 数据导出中心 (事件触发型)
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type, action_config)
VALUES (0, 2, 'data_export.export_center', NULL, 85, 'download', '/export-center', 0, 0, 0, 'event', '{"event": "data-export-center"}');

-- 3.2 工具箱 (已存在，更新 hidden=0 使其显示在顶部菜单)
UPDATE core_menu SET hidden = 0, menu_sort = 88 WHERE id = 30;

-- 3.3 帮助文档分组 (外部链接分组)
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (0, 1, 'api_pagination.help_documentation', NULL, 92, 'document', '/help', 0, 0, 0, 'group');

SET @help_pid = LAST_INSERT_ID();

-- 3.3.1 帮助文档子项
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type, action_config)
VALUES 
(@help_pid, 2, 'api_pagination.help_documentation', NULL, 1, 'document', '/help/doc', 0, 0, 0, 'external', '{"url": "https://dataease.io/docs/v2/", "dynamicUrl": "helpUrl"}'),
(@help_pid, 2, 'api_pagination.product_forum', NULL, 2, 'forum', '/help/forum', 0, 0, 0, 'external', '{"url": "https://bbs.fit2cloud.com/c/de/6"}'),
(@help_pid, 2, 'api_pagination.technical_blog', NULL, 3, 'blog', '/help/blog', 0, 0, 0, 'external', '{"url": "https://blog.fit2cloud.com/categories/dataease"}'),
(@help_pid, 2, 'api_pagination.enterprise_edition_trial', NULL, 4, 'enterprise', '/help/trial', 0, 0, 0, 'external', '{"url": "https://jinshuju.net/f/TK5TTd"}');

-- 3.4 我的菜单分组 (用户相关功能)
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (0, 1, 'commons.mine', NULL, 99, 'user', '/mine', 0, 0, 0, 'group');

SET @mine_pid = LAST_INSERT_ID();

-- 3.4.1 关于 (事件触发)
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type, action_config)
VALUES (@mine_pid, 2, 'common.about', NULL, 1, 'info', '/mine/about', 0, 0, 0, 'event', '{"event": "open-about-dialog"}');

-- 3.4.2 修改密码 (路由)
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (@mine_pid, 2, 'user.change_password', 'system/modify-pwd/index', 2, 'lock', '/mine/modify-pwd', 0, 0, 0, 'route');

-- 3.4.3 系统设置 (路由，仅管理员)
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (@mine_pid, 2, 'commons.system_setting', NULL, 3, 'setting', '/mine/sys-setting', 0, 0, 1, 'route');

-- 3.4.4 语言选择 (事件触发)
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type, action_config)
VALUES (@mine_pid, 2, 'commons.language', NULL, 4, 'language', '/mine/language', 0, 0, 0, 'event', '{"event": "open-language-selector"}');

-- 3.4.5 退出登录 (事件触发)
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type, action_config)
VALUES (@mine_pid, 2, 'common.exit_system', NULL, 5, 'logout', '/mine/logout', 0, 0, 0, 'event', '{"event": "user-logout"}');

-- 4. 为新菜单授权给管理员角色
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
