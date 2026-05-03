-- User authentication baseline.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` VARCHAR(64) NOT NULL COMMENT '用户名',
  `email` VARCHAR(191) NOT NULL COMMENT '邮箱',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `display_name` VARCHAR(64) NULL COMMENT '显示名称',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '状态：active/disabled',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`),
  UNIQUE KEY `uk_users_email` (`email`),
  KEY `idx_users_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户端账号';

CREATE TABLE IF NOT EXISTS `user_sessions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `session_id` VARCHAR(64) NOT NULL COMMENT '会话唯一标识，对应用户端JWT jti',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '状态：active/revoked/expired',
  `issued_at` DATETIME(3) NOT NULL COMMENT '签发时间',
  `expires_at` DATETIME(3) NOT NULL COMMENT '过期时间',
  `revoked_at` DATETIME(3) NULL COMMENT '吊销时间',
  `revoke_reason` VARCHAR(64) NULL COMMENT '吊销原因',
  `last_seen_at` DATETIME(3) NULL COMMENT '最近访问时间',
  `last_seen_ip` VARCHAR(64) NULL COMMENT '最近访问IP',
  `user_agent` VARCHAR(500) NULL COMMENT '浏览器User-Agent',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_sessions_session_id` (`session_id`),
  KEY `idx_user_sessions_user_status` (`user_id`, `status`),
  KEY `idx_user_sessions_expires_at` (`expires_at`),
  CONSTRAINT `fk_user_sessions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户端登录会话';
