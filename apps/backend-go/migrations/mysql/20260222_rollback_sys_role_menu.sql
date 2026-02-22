-- Migration: Rollback sys_role_menu table
-- Description: 回滚角色菜单关联表，支持备份恢复
-- Date: 2026-02-22

SET @backup_enabled = IFNULL((SELECT VALUE FROM system_parameters WHERE param_key = 'migration.backup_enabled'), 'true');

SET @table_exists = (
    SELECT COUNT(*)
    FROM information_schema.tables
    WHERE table_schema = DATABASE()
    AND table_name = 'sys_role_menu'
);

IF @table_exists > 0 AND @backup_enabled = 'true' THEN
    CREATE TABLE IF NOT EXISTS sys_role_menu_backup_20260222 AS
    SELECT * FROM sys_role_menu;
END IF;

DROP TABLE IF EXISTS sys_role_menu;
