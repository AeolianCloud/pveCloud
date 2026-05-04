-- User account self-service: registration support, password reset tokens and profile editing.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `user_password_reset_tokens` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '密码重置Token ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `token_hash` CHAR(64) NOT NULL COMMENT '密码重置Token哈希，禁止保存Token明文',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '状态：active/used/revoked/expired',
  `expires_at` DATETIME(3) NOT NULL COMMENT '过期时间',
  `used_at` DATETIME(3) NULL COMMENT '使用时间',
  `requested_ip` VARCHAR(64) NULL COMMENT '申请IP',
  `user_agent` VARCHAR(500) NULL COMMENT '申请浏览器User-Agent',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_password_reset_tokens_token_hash` (`token_hash`),
  KEY `idx_user_password_reset_tokens_user_status` (`user_id`, `status`),
  KEY `idx_user_password_reset_tokens_expires_at` (`expires_at`),
  CONSTRAINT `fk_user_password_reset_tokens_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户端密码重置Token';
