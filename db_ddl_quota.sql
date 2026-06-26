-- 配额控制与限流功能数据库建表脚本
-- 版本: v0.2.2
-- 创建时间: 2026-06-23

-- 1. 创建 entity_types 表（Entity类型定义表）
CREATE TABLE IF NOT EXISTS `entity_types` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `type_name` VARCHAR(32) NOT NULL COMMENT '类型标识',
  `description` VARCHAR(256) DEFAULT '' COMMENT '类型描述',
  `level` INT NOT NULL DEFAULT 1 COMMENT '层级级别（1-5，数字越小级别越高）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY `uk_type_name` (`type_name`),
  INDEX `idx_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Entity类型定义表';

-- 2. 创建 entities 表（Entity实体表）
CREATE TABLE IF NOT EXISTS `entities` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `entity_id` VARCHAR(64) NOT NULL COMMENT 'Entity唯一标识（业务ID）',
  `name` VARCHAR(128) NOT NULL COMMENT 'Entity名称',
  `type` VARCHAR(32) NOT NULL COMMENT 'Entity类型（关联entity_types.type_name）',
  `parent_id` VARCHAR(64) DEFAULT NULL COMMENT '父Entity ID',
  `allow_models` TEXT COMMENT '允许访问的模型白名单（JSON数组）',
  `block_models` TEXT COMMENT '禁止访问的模型黑名单（JSON数组）',
  `quota_plan_id` BIGINT DEFAULT NULL COMMENT '配额计划ID',
  `rate_limit_policy_id` BIGINT DEFAULT NULL COMMENT '限流策略ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY `uk_entity_id` (`entity_id`),
  UNIQUE KEY `uk_name` (`name`),
  INDEX `idx_parent_id` (`parent_id`),
  INDEX `idx_type` (`type`),
  INDEX `idx_quota_plan_id` (`quota_plan_id`),
  INDEX `idx_rate_limit_policy_id` (`rate_limit_policy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Entity实体表';

-- 3. 创建 quota_plans 表（配额计划表）
CREATE TABLE IF NOT EXISTS `quota_plans` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `unlimited` TINYINT(1) DEFAULT 1 COMMENT '是否无限配额：0-有限，1-无限',
  `pass_when_no_enough_quota` TINYINT(1) DEFAULT 0 COMMENT '配额不足时是否放行：0-拒绝，1-放行',
  `quota` BIGINT DEFAULT 0 COMMENT '配额总量',
  `unit` VARCHAR(32) DEFAULT 'total_token' COMMENT '配额单位',
  `reset_period` VARCHAR(16) DEFAULT 'never' COMMENT '配额重置周期：never/weekly/monthly，重置均基于日历周期（如自然周/自然月）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX `idx_unlimited` (`unlimited`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配额计划表';

-- 4. 创建 quota_balances 表（配额余额表）
CREATE TABLE IF NOT EXISTS `quota_balances` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `quota_plan_id` BIGINT NOT NULL COMMENT '配额计划ID',
  `used` BIGINT DEFAULT 0 COMMENT '已使用量',
  `remaining` BIGINT DEFAULT 0 COMMENT '剩余量',
  `last_reset_at` DATETIME DEFAULT NULL COMMENT '上次重置时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY `uk_quota_plan_id` (`quota_plan_id`),
  INDEX `idx_remaining` (`remaining`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配额余额表';

-- 5. 创建 rate_limit_policies 表（限流策略表）
CREATE TABLE IF NOT EXISTS `rate_limit_policies` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `enabled` TINYINT(1) DEFAULT 0 COMMENT '是否启用：0-禁用，1-启用',
  `max_concurrency` INT DEFAULT -1 COMMENT '最大并发数（-1表示不限制）',
  `tpm_configs` TEXT COMMENT 'TPM限流配置（JSON数组）',
  `rpm_configs` TEXT COMMENT 'RPM限流配置（JSON数组）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='限流策略表';

-- 6. 修改 api_keys 表，添加新字段
ALTER TABLE `api_keys` 
ADD COLUMN `description` VARCHAR(512) DEFAULT '' COMMENT 'API-Key描述',
ADD COLUMN `unlimited_quota` TINYINT(1) DEFAULT 0 COMMENT '是否无限配额：0-有限，1-无限',
ADD COLUMN `subnet` TEXT COMMENT '允许的客户端子网（JSON数组）',
ADD COLUMN `entity_id` VARCHAR(64) DEFAULT NULL COMMENT '挂载的Entity ID',
ADD COLUMN `quota_plan_id` BIGINT DEFAULT NULL COMMENT '配额计划ID',
ADD COLUMN `rate_limit_policy_id` BIGINT DEFAULT NULL COMMENT '限流策略ID';

-- 7. 为 api_keys 表添加索引
ALTER TABLE `api_keys`
ADD INDEX `idx_entity_id` (`entity_id`),
ADD INDEX `idx_quota_plan_id` (`quota_plan_id`),
ADD INDEX `idx_rate_limit_policy_id` (`rate_limit_policy_id`);

-- 8. 创建初始数据（可选）
-- 插入默认的Entity类型
INSERT INTO `entity_types` (`type_name`, `description`, `level`) VALUES
('organization', '组织', 1),
('department', '部门', 2),
('team', '团队', 3),
('user', '用户', 4)
ON DUPLICATE KEY UPDATE `description` = VALUES(`description`);

-- 插入默认的配额计划（无限配额）
INSERT INTO `quota_plans` (`unlimited`, `pass_when_no_enough_quota`, `quota`, `unit`, `reset_period`) VALUES
(1, 0, 0, 'total_token', 'never')
ON DUPLICATE KEY UPDATE `unlimited` = VALUES(`unlimited`);

-- 插入默认的限流策略（不限制）
INSERT INTO `rate_limit_policies` (`enabled`, `max_concurrency`, `tpm_configs`, `rpm_configs`) VALUES
(0, -1, '[]', '[]')
ON DUPLICATE KEY UPDATE `enabled` = VALUES(`enabled`);
