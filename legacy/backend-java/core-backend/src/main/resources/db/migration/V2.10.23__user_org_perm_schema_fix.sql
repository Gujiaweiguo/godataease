-- Align user/org/permission schema with entities and mappers
-- Version: V2.10.23__user_org_perm_schema_fix

SET @schema := DATABASE();

-- sys_perm: add missing columns
-- 检查表是否存在，不存在则跳过迁移

SET @sys_perm_exists = (SELECT COUNT(*) FROM information_schema.TABLES WHERE table_schema = @schema AND table_name = 'sys_perm');

-- 只有当 sys_perm 表存在时才执行迁移
SET @should_migrate = IF(@sys_perm_exists > 0, 1, 0);

-- add perm_key column
SET @perm_key_exists = IF(@should_migrate = 1,
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'perm_key'),
    0
);

SET @add_perm_key_sql = IF(@perm_key_exists = 0 AND @should_migrate = 1,
    'ALTER TABLE sys_perm ADD COLUMN perm_key VARCHAR(64) COMMENT ''权限标识'' AFTER perm_name',
    'SELECT 1'
);

PREPARE add_perm_key_stmt FROM @add_perm_key_sql;
EXECUTE add_perm_key_stmt;
DEALLOCATE PREPARE add_perm_key_stmt;

-- add perm_desc column
SET @perm_desc_exists = IF(@should_migrate = 1,
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'perm_desc'),
    0
);

SET @add_perm_desc_sql = IF(@perm_desc_exists = 0 AND @should_migrate = 1,
    'ALTER TABLE sys_perm ADD COLUMN perm_desc TEXT COMMENT ''权限描述'' AFTER perm_type',
    'SELECT 1'
);

PREPARE add_perm_desc_stmt FROM @add_perm_desc_sql;
EXECUTE add_perm_desc_stmt;
DEALLOCATE PREPARE add_perm_desc_stmt;

-- add status column
SET @status_exists = IF(@should_migrate = 1,
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'status'),
    0
);

SET @add_status_sql = IF(@status_exists = 0 AND @should_migrate = 1,
    'ALTER TABLE sys_perm ADD COLUMN status TINYINT(1) NOT NULL DEFAULT 1 COMMENT ''状态：0-禁用，1-启用'' AFTER perm_desc',
    'SELECT 1'
);

PREPARE add_status_stmt FROM @add_status_sql;
EXECUTE add_status_stmt;
DEALLOCATE PREPARE add_status_stmt;

-- add update_by column
SET @update_by_exists = IF(@should_migrate = 1,
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'update_by'),
    0
);

SET @add_update_by_sql = IF(@update_by_exists = 0 AND @should_migrate = 1,
    'ALTER TABLE sys_perm ADD COLUMN update_by VARCHAR(50) COMMENT ''更新人'' AFTER status',
    'SELECT 1'
);

PREPARE add_update_by_stmt FROM @add_update_by_sql;
EXECUTE add_update_by_stmt;
DEALLOCATE PREPARE add_update_by_stmt;

-- add update_time column
SET @update_time_exists = IF(@should_migrate = 1,
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'update_time'),
    0
);

SET @add_update_time_sql = IF(@update_time_exists = 0 AND @should_migrate = 1,
    'ALTER TABLE sys_perm ADD COLUMN update_time BIGINT COMMENT ''更新时间'' AFTER update_by',
    'SELECT 1'
);

PREPARE add_update_time_stmt FROM @add_update_time_sql;
EXECUTE add_update_time_stmt;
DEALLOCATE PREPARE add_update_time_stmt;

-- add del_flag column
SET @del_flag_exists = IF(@should_migrate = 1,
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'del_flag'),
    0
);

SET @add_del_flag_sql = IF(@del_flag_exists = 0 AND @should_migrate = 1,
    'ALTER TABLE sys_perm ADD COLUMN del_flag TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''删除标记：0-未删除，1-已删除'' AFTER update_time',
    'SELECT 1'
);

PREPARE add_del_flag_stmt FROM @add_del_flag_sql;
EXECUTE add_del_flag_stmt;
DEALLOCATE PREPARE add_del_flag_stmt;

-- Backfill defaults
SET @update_perm_key_sql = IF(@should_migrate = 1,
    'UPDATE sys_perm SET perm_key = perm_code WHERE perm_key IS NULL AND perm_code IS NOT NULL',
    'SELECT 1'
);
SET @update_perm_key_exists = IF(@should_migrate = 1,
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'perm_key'),
    0
);

SET @execute_update_perm_key = IF(@update_perm_key_exists = 1 AND @should_migrate = 1,
    @update_perm_key_sql,
    'SELECT 1'
);
PREPARE update_perm_key_stmt FROM @execute_update_perm_key;
EXECUTE update_perm_key_stmt;
DEALLOCATE PREPARE update_perm_key_stmt;

SET @update_update_time_sql = IF(@should_migrate = 1,
    'UPDATE sys_perm SET update_time = create_time WHERE update_time IS NULL AND create_time IS NOT NULL',
    'SELECT 1'
);
SET @update_update_time_exists = IF(@should_migrate = 1,
    (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE table_schema = @schema AND table_name = 'sys_perm' AND column_name = 'update_time'),
    0
);

SET @execute_update_update_time = IF(@update_update_time_exists = 1 AND @should_migrate = 1,
    @update_update_time_sql,
    'SELECT 1'
);
PREPARE update_update_time_stmt FROM @execute_update_update_time;
EXECUTE update_update_time_stmt;
DEALLOCATE PREPARE update_update_time_stmt;

SELECT CONCAT('Migration completed successfully - added columns if sys_perm table exists') AS result;
