-- Fix permission data encoding and grant admin all permissions
-- Run with: docker exec -i mysql8 mysql -u root -pAdmin168 dataease_dev < fix_perm_data.sql

SET NAMES utf8mb4;

-- Clear existing permission data
DELETE FROM sys_role_perm WHERE role_id = 1;
DELETE FROM sys_perm;

-- Reset auto increment
ALTER TABLE sys_perm AUTO_INCREMENT = 1;

-- Insert correct permission data with proper UTF-8 encoding
INSERT INTO sys_perm (perm_id, perm_name, perm_key, perm_type, perm_desc, status, create_by, create_time, del_flag) VALUES
(1, '数据准备', 'data:prepare', 'menu', '数据准备模块访问权限', 1, 'system', NOW(), 0),
(2, '菜单查看', 'menu:view', 'menu', '菜单查看权限', 1, 'system', NOW(), 0),
(3, '仪表板查看', 'dashboard:view', 'data', '仪表板查看权限', 1, 'system', NOW(), 0),
(4, '仪表板编辑', 'dashboard:edit', 'data', '仪表板编辑权限', 1, 'system', NOW(), 0),
(5, '数据大屏查看', 'screen:view', 'data', '数据大屏查看权限', 1, 'system', NOW(), 0),
(6, '数据大屏编辑', 'screen:edit', 'data', '数据大屏编辑权限', 1, 'system', NOW(), 0),
(7, '数据源查看', 'datasource:view', 'data', '数据源查看权限', 1, 'system', NOW(), 0),
(8, '数据源管理', 'datasource:manage', 'data', '数据源管理权限', 1, 'system', NOW(), 0),
(9, '数据集查看', 'dataset:view', 'data', '数据集查看权限', 1, 'system', NOW(), 0),
(10, '数据集管理', 'dataset:manage', 'data', '数据集管理权限', 1, 'system', NOW(), 0),
(11, '用户管理', 'user:manage', 'system', '用户管理权限', 1, 'system', NOW(), 0),
(12, '角色管理', 'role:manage', 'system', '角色管理权限', 1, 'system', NOW(), 0),
(13, '组织管理', 'org:manage', 'system', '组织管理权限', 1, 'system', NOW(), 0),
(14, '系统参数', 'system:param', 'system', '系统参数管理权限', 1, 'system', NOW(), 0),
(15, '审计日志', 'audit:view', 'system', '审计日志查看权限', 1, 'system', NOW(), 0);

-- Grant admin role (role_id=1) all permissions
INSERT INTO sys_role_perm (role_id, perm_id)
SELECT 1, perm_id FROM sys_perm WHERE del_flag = 0;
