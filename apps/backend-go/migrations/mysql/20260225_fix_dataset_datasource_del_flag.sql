-- Fix del_flag column missing in core_dataset_group and core_datasource tables
-- This migration adds the del_flag column required by Go backend repository queries
-- Run with: docker exec -i mysql8 mysql -u root -pAdmin168 dataease_dev < 20260225_fix_dataset_datasource_del_flag.sql

-- Add del_flag column to core_dataset_group if not exists
SET @dbname = DATABASE();
SET @tablename = 'core_dataset_group';
SET @columnname = 'del_flag';
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ', @columnname, ' INT DEFAULT 0')
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- Add del_flag column to core_datasource if not exists
SET @tablename = 'core_datasource';
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = @columnname) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ', @columnname, ' INT DEFAULT 0')
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;
