-- Audit Logs Tables
-- 审计日志表 - 用于记录系统所有操作日志

-- 审计日志主表
CREATE TABLE IF NOT EXISTS `de_audit_log` (
    `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT(20) DEFAULT NULL COMMENT '操作用户ID',
    `username` VARCHAR(100) DEFAULT NULL COMMENT '操作用户名',
    `action_type` VARCHAR(50) NOT NULL COMMENT '操作类型: USER_ACTION, PERMISSION_CHANGE, DATA_ACCESS, SYSTEM_CONFIG',
    `action_name` VARCHAR(100) NOT NULL COMMENT '操作名称: CREATE_USER, DELETE_ROLE, EXPORT_DATA, etc.',
    `resource_type` VARCHAR(50) DEFAULT NULL COMMENT '资源类型: USER, ORGANIZATION, ROLE, PERMISSION, DATASET, DASHBOARD',
    `resource_id` BIGINT(20) DEFAULT NULL COMMENT '操作的资源ID',
    `resource_name` VARCHAR(200) DEFAULT NULL COMMENT '资源名称',
    `operation` VARCHAR(20) NOT NULL COMMENT '操作类型: CREATE, UPDATE, DELETE, EXPORT, LOGIN, LOGOUT',
    `status` VARCHAR(20) NOT NULL DEFAULT 'SUCCESS' COMMENT '操作状态: SUCCESS, FAILED',
    `failure_reason` VARCHAR(500) DEFAULT NULL COMMENT '失败原因',
    `ip_address` VARCHAR(50) DEFAULT NULL COMMENT '操作IP地址',
    `user_agent` VARCHAR(500) DEFAULT NULL COMMENT '用户代理',
    `before_value` TEXT DEFAULT NULL COMMENT '操作前的值（JSON格式）',
    `after_value` TEXT DEFAULT NULL COMMENT '操作后的值（JSON格式）',
    `organization_id` BIGINT(20) DEFAULT NULL COMMENT '操作时所在的组织ID',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_action_type` (`action_type`),
    INDEX `idx_resource_type` (`resource_type`),
    INDEX `idx_organization_id` (`organization_id`),
    INDEX `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志主表';

-- 审计日志详情表（用于存储大对象或详细信息）
CREATE TABLE IF NOT EXISTS `de_audit_log_detail` (
    `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `audit_log_id` BIGINT(20) NOT NULL COMMENT '审计日志主表ID',
    `detail_type` VARCHAR(50) DEFAULT NULL COMMENT '详情类型: BEFORE_AFTER, PERMISSION_CHANGE, EXPORT_RECORD',
    `detail_key` VARCHAR(100) DEFAULT NULL COMMENT '详情键',
    `detail_value` TEXT DEFAULT NULL COMMENT '详情值',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_audit_log_id` (`audit_log_id`),
    FOREIGN KEY (`audit_log_id`) REFERENCES `de_audit_log` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志详情表';

-- 登录失败记录表
CREATE TABLE IF NOT EXISTS `de_login_failure` (
    `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `username` VARCHAR(100) NOT NULL COMMENT '尝试登录的用户名',
    `ip_address` VARCHAR(50) DEFAULT NULL COMMENT '登录IP地址',
    `failure_reason` VARCHAR(200) DEFAULT NULL COMMENT '失败原因',
    `user_agent` VARCHAR(500) DEFAULT NULL COMMENT '用户代理',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_username` (`username`),
    INDEX `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录失败记录表';
