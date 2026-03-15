INSERT INTO sys_role_perm (role_id, perm_id)
SELECT r.role_id, p.perm_id
FROM sys_role r
JOIN sys_perm p ON p.del_flag = 0
WHERE (r.role_id = 1 OR LOWER(r.role_code) = 'admin')
  AND p.perm_key IN (
    'datasource:view',
    'datasource:manage',
    'dataset:view',
    'dataset:manage',
    'dashboard:view',
    'dashboard:edit',
    'screen:view',
    'screen:edit'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM sys_role_perm rp
    WHERE rp.role_id = r.role_id
      AND rp.perm_id = p.perm_id
  );
