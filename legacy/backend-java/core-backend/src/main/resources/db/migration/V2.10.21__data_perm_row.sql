-- Row-Level Permissions Table for MVP
-- Version: V2.10.21__data_perm_row

CREATE TABLE IF NOT EXISTS `data_perm_row` (
  `id`               BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `dataset_id`       BIGINT      NOT NULL COMMENT '数据集ID',
  `auth_target_type`  VARCHAR(20) NOT NULL COMMENT '授权对象类型：user-用户，role-角色',
  `auth_target_id`    BIGINT      NOT NULL COMMENT '授权对象ID（用户ID或角色ID）',
  `expression_tree`   TEXT        NOT NULL COMMENT '权限表达式树：JSON格式 { logic: "or"|"and", items: [...] }',
  `status`           TINYINT(1)  NOT NULL DEFAULT 1 COMMENT '状态：0-禁用，1-启用',
  `create_by`        VARCHAR(50)     NULL COMMENT '创建人',
  `create_time`      BIGINT      NOT NULL COMMENT '创建时间',
  `update_by`        VARCHAR(50)     NULL COMMENT '更新人',
  `update_time`      BIGINT      NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_dataset_id` (`dataset_id`),
  KEY `idx_auth_target` (`auth_target_type`, `auth_target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据集行级权限表';
