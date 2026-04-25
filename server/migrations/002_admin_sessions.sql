-- 管理端登录会话，用于 JWT 吊销、刷新轮换和会话自检。
-- Target: MariaDB 11.4.9 / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `admin_sessions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '管理端会话ID',
  `session_id` VARCHAR(64) NOT NULL COMMENT 'JWT jti，会话唯一标识',
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '会话状态：active/revoked/expired',
  `issued_at` DATETIME(3) NOT NULL COMMENT '签发时间',
  `expires_at` DATETIME(3) NOT NULL COMMENT '过期时间',
  `last_seen_at` DATETIME(3) NULL COMMENT '最后访问时间',
  `last_seen_ip` VARCHAR(64) NULL COMMENT '最后访问IP',
  `user_agent` VARCHAR(500) NULL COMMENT '登录或最近访问User-Agent',
  `revoked_at` DATETIME(3) NULL COMMENT '吊销时间',
  `revoke_reason` VARCHAR(64) NULL COMMENT '吊销原因：logout/refresh/admin_disabled/expired',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_sessions_session_id` (`session_id`),
  KEY `idx_admin_sessions_admin_status` (`admin_id`, `status`),
  KEY `idx_admin_sessions_status_expires` (`status`, `expires_at`),
  KEY `idx_admin_sessions_expires_at` (`expires_at`),
  CONSTRAINT `fk_admin_sessions_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理端登录会话';
