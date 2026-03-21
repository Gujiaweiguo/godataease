-- Migration: Add menu_location field for dynamic menu positioning
-- Description: 添加 menu_location 字段支持菜单位置配置
-- Date: 2026-03-19

ALTER TABLE core_menu 
ADD COLUMN menu_location VARCHAR(20) DEFAULT 'sidebar' COMMENT '菜单位置: sidebar=侧边栏, user_menu=用户菜单, help_menu=帮助菜单' AFTER auth;

UPDATE core_menu SET menu_location = 'sidebar' WHERE menu_location IS NULL OR menu_location = '';
