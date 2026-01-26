-- Align user/org/permission schema with entities and mappers
-- Version: V2.10.23__user_org_perm_schema_fix

SET @schema := DATABASE();

-- sys_perm: add missing columns
SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'perm_key') = 0,
    'ALTER TABLE sys_perm ADD COLUMN perm_key VARCHAR(64) NULL COMMENT ''权限标识'' AFTER perm_name',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'perm_desc') = 0,
    'ALTER TABLE sys_perm ADD COLUMN perm_desc TEXT NULL COMMENT ''权限描述'' AFTER perm_type',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'status') = 0,
    'ALTER TABLE sys_perm ADD COLUMN status TINYINT(1) NOT NULL DEFAULT 1 COMMENT ''状态：0-禁用，1-启用'' AFTER perm_desc',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'update_by') = 0,
    'ALTER TABLE sys_perm ADD COLUMN update_by VARCHAR(50) NULL COMMENT ''更新人'' AFTER create_time',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'update_time') = 0,
    'ALTER TABLE sys_perm ADD COLUMN update_time BIGINT NULL COMMENT ''更新时间'' AFTER update_by',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'del_flag') = 0,
    'ALTER TABLE sys_perm ADD COLUMN del_flag TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''删除标记：0-未删除，1-已删除'' AFTER update_time',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- sys_user_perm: add status and del_flag
SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_user_perm' AND column_name = 'status') = 0,
    'ALTER TABLE sys_user_perm ADD COLUMN status TINYINT(1) NOT NULL DEFAULT 1 COMMENT ''状态：0-禁用，1-启用'' AFTER perm_id',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_user_perm' AND column_name = 'del_flag') = 0,
    'ALTER TABLE sys_user_perm ADD COLUMN del_flag TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''删除标记：0-未删除，1-已删除'' AFTER status',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- data_perm_row: add MVP columns if missing
SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'data_perm_row' AND column_name = 'auth_target_type') = 0,
    'ALTER TABLE data_perm_row ADD COLUMN auth_target_type VARCHAR(20) NULL COMMENT ''授权对象类型：user-用户，role-角色'' AFTER dataset_id',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'data_perm_row' AND column_name = 'auth_target_id') = 0,
    'ALTER TABLE data_perm_row ADD COLUMN auth_target_id BIGINT NULL COMMENT ''授权对象ID（用户ID或角色ID）'' AFTER auth_target_type',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'data_perm_row' AND column_name = 'expression_tree') = 0,
    'ALTER TABLE data_perm_row ADD COLUMN expression_tree TEXT NULL COMMENT ''权限表达式树：JSON格式 { logic: "or"|"and", items: [...] }'' AFTER auth_target_id',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'data_perm_row' AND column_name = 'status') = 0,
    'ALTER TABLE data_perm_row ADD COLUMN status TINYINT(1) NOT NULL DEFAULT 1 COMMENT ''状态：0-禁用，1-启用'' AFTER expression_tree',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
    (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE table_schema = @schema AND table_name = 'data_perm_row' AND index_name = 'idx_auth_target') = 0,
    'CREATE INDEX idx_auth_target ON data_perm_row (auth_target_type, auth_target_id)',
    'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Backfill defaults
UPDATE sys_perm SET perm_key = perm_code WHERE perm_key IS NULL AND perm_code IS NOT NULL;
UPDATE sys_perm SET update_time = create_time WHERE update_time IS NULL AND create_time IS NOT NULL;
UPDATE sys_user_perm SET status = 1 WHERE status IS NULL;
UPDATE sys_user_perm SET del_flag = 0 WHERE del_flag IS NULL;
UPDATE data_perm_row SET status = 1 WHERE status IS NULL;
