INSERT INTO sys_perm (perm_name, perm_key, perm_type, perm_desc, status, create_by, del_flag)
SELECT '菜单查看', 'menu:view', 'menu', '菜单查看权限', 1, 'system', 0
WHERE NOT EXISTS (
    SELECT 1 FROM sys_perm WHERE perm_key = 'menu:view' AND del_flag = 0
);

INSERT INTO sys_perm (perm_name, perm_key, perm_type, perm_desc, status, create_by, del_flag)
SELECT '仪表板查看', 'dashboard:view', 'data', '仪表板查看权限', 1, 'system', 0
WHERE NOT EXISTS (
    SELECT 1 FROM sys_perm WHERE perm_key = 'dashboard:view' AND del_flag = 0
);

INSERT INTO sys_perm (perm_name, perm_key, perm_type, perm_desc, status, create_by, del_flag)
SELECT '仪表板编辑', 'dashboard:edit', 'data', '仪表板编辑权限', 1, 'system', 0
WHERE NOT EXISTS (
    SELECT 1 FROM sys_perm WHERE perm_key = 'dashboard:edit' AND del_flag = 0
);

INSERT INTO sys_perm (perm_name, perm_key, perm_type, perm_desc, status, create_by, del_flag)
SELECT '数据大屏查看', 'screen:view', 'data', '数据大屏查看权限', 1, 'system', 0
WHERE NOT EXISTS (
    SELECT 1 FROM sys_perm WHERE perm_key = 'screen:view' AND del_flag = 0
);

INSERT INTO sys_perm (perm_name, perm_key, perm_type, perm_desc, status, create_by, del_flag)
SELECT '数据大屏编辑', 'screen:edit', 'data', '数据大屏编辑权限', 1, 'system', 0
WHERE NOT EXISTS (
    SELECT 1 FROM sys_perm WHERE perm_key = 'screen:edit' AND del_flag = 0
);
