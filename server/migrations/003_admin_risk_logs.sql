-- 管理端高危操作日志，用于和普通审计日志分开查看、告警和复核。
-- Target: MariaDB 11.4.9 / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `admin_risk_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '高危操作日志ID',
  `audit_log_id` BIGINT UNSIGNED NULL COMMENT '关联普通审计日志ID',
  `admin_id` BIGINT UNSIGNED NULL COMMENT '管理员ID',
  `risk_level` VARCHAR(32) NOT NULL COMMENT '风险等级：medium/high/critical',
  `action` VARCHAR(96) NOT NULL COMMENT '高危操作动作',
  `object_type` VARCHAR(64) NOT NULL COMMENT '操作对象类型',
  `object_id` VARCHAR(64) NULL COMMENT '操作对象ID',
  `risk_reason` VARCHAR(255) NOT NULL COMMENT '判定为高危的原因',
  `before_data` JSON NULL COMMENT '操作前数据，需脱敏',
  `after_data` JSON NULL COMMENT '操作后数据，需脱敏',
  `ip` VARCHAR(64) NULL COMMENT '操作IP',
  `user_agent` VARCHAR(500) NULL COMMENT '浏览器User-Agent',
  `remark` VARCHAR(500) NULL COMMENT '操作备注',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_risk_logs_audit_log` (`audit_log_id`),
  KEY `idx_admin_risk_logs_admin_created` (`admin_id`, `created_at`),
  KEY `idx_admin_risk_logs_level_created` (`risk_level`, `created_at`),
  KEY `idx_admin_risk_logs_object` (`object_type`, `object_id`),
  KEY `idx_admin_risk_logs_action_created` (`action`, `created_at`),
  CONSTRAINT `fk_admin_risk_logs_audit_log` FOREIGN KEY (`audit_log_id`) REFERENCES `admin_audit_logs` (`id`),
  CONSTRAINT `fk_admin_risk_logs_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台高危操作日志';

INSERT INTO `admin_permissions` (`code`, `name`, `group_name`, `description`) VALUES
  ('audit:sensitive_view', '查看审计敏感详情', 'audit', '查看审计日志和高危操作日志中的已脱敏详情字段')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `group_name` = VALUES(`group_name`),
  `description` = VALUES(`description`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT `admin_roles`.`id`, `admin_permissions`.`id`
FROM `admin_roles`
JOIN `admin_permissions`
WHERE `admin_roles`.`code` = 'super_admin'
  AND `admin_permissions`.`code` = 'audit:sensitive_view'
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
