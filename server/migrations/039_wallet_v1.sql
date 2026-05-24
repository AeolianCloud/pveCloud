-- Wallet v1 schema, permissions and system config seeds.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.
-- This migration re-opens the wallet domain after the legacy cleanup in 009.
-- It is intentionally idempotent and only creates new wallet facts plus stable
-- seeds; it does not restore old wallet data that may have existed before 009.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @payment_provider_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'payment_transactions'
    AND COLUMN_NAME = 'provider'
);
SET @modify_payment_provider_for_wallet_sql := IF(
  @payment_provider_column_exists = 1,
  'ALTER TABLE `payment_transactions` MODIFY COLUMN `provider` VARCHAR(32) NOT NULL COMMENT ''支付供应商：alipay/wechat/wallet''',
  'SELECT 1'
);
PREPARE modify_payment_provider_for_wallet_stmt FROM @modify_payment_provider_for_wallet_sql;
EXECUTE modify_payment_provider_for_wallet_stmt;
DEALLOCATE PREPARE modify_payment_provider_for_wallet_stmt;

SET @payment_method_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'payment_transactions'
    AND COLUMN_NAME = 'method'
);
SET @modify_payment_method_for_wallet_sql := IF(
  @payment_method_column_exists = 1,
  'ALTER TABLE `payment_transactions` MODIFY COLUMN `method` VARCHAR(32) NOT NULL COMMENT ''支付方式：alipay_page/alipay_wap/wechat_native/wechat_h5/wallet_balance''',
  'SELECT 1'
);
PREPARE modify_payment_method_for_wallet_stmt FROM @modify_payment_method_for_wallet_sql;
EXECUTE modify_payment_method_for_wallet_stmt;
DEALLOCATE PREPARE modify_payment_method_for_wallet_stmt;

SET @refund_provider_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'refund_transactions'
    AND COLUMN_NAME = 'provider'
);
SET @modify_refund_provider_for_wallet_sql := IF(
  @refund_provider_column_exists = 1,
  'ALTER TABLE `refund_transactions` MODIFY COLUMN `provider` VARCHAR(32) NOT NULL COMMENT ''退款来源支付供应商：alipay/wechat/wallet''',
  'SELECT 1'
);
PREPARE modify_refund_provider_for_wallet_stmt FROM @modify_refund_provider_for_wallet_sql;
EXECUTE modify_refund_provider_for_wallet_stmt;
DEALLOCATE PREPARE modify_refund_provider_for_wallet_stmt;

CREATE TABLE IF NOT EXISTS `wallet_accounts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '钱包账户ID',
  `wallet_no` VARCHAR(64) NOT NULL COMMENT '对外钱包编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种，v1固定CNY',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '钱包状态：active/disabled',
  `available_balance_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '可用余额，单位分',
  `total_recharged_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计充值入账金额，单位分',
  `total_spent_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计余额支付扣款金额，单位分',
  `total_refunded_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计退回钱包金额，单位分',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wallet_accounts_wallet_no` (`wallet_no`),
  UNIQUE KEY `uk_wallet_accounts_user_currency` (`user_id`, `currency`),
  KEY `idx_wallet_accounts_status` (`status`, `created_at`),
  CONSTRAINT `fk_wallet_accounts_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户钱包账户';

CREATE TABLE IF NOT EXISTS `wallet_ledger_entries` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '钱包流水ID',
  `entry_no` VARCHAR(64) NOT NULL COMMENT '对外流水编号',
  `wallet_id` BIGINT UNSIGNED NOT NULL COMMENT '钱包账户ID',
  `wallet_no` VARCHAR(64) NOT NULL COMMENT '钱包编号快照',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `direction` VARCHAR(16) NOT NULL COMMENT '流水方向：credit/debit',
  `entry_type` VARCHAR(32) NOT NULL COMMENT '流水类型：recharge/payment/refund',
  `amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '变动金额，单位分',
  `balance_before_cents` BIGINT UNSIGNED NOT NULL COMMENT '变动前余额，单位分',
  `balance_after_cents` BIGINT UNSIGNED NOT NULL COMMENT '变动后余额，单位分',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种，v1固定CNY',
  `related_type` VARCHAR(32) NOT NULL COMMENT '关联对象类型：recharge/payment/refund/order',
  `related_no` VARCHAR(64) NOT NULL COMMENT '关联对象编号',
  `idempotency_key` VARCHAR(160) NOT NULL COMMENT '钱包流水幂等键',
  `summary` JSON NULL COMMENT '业务摘要，不保存渠道完整响应或敏感原文',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wallet_ledger_entries_entry_no` (`entry_no`),
  UNIQUE KEY `uk_wallet_ledger_entries_idempotency` (`wallet_id`, `idempotency_key`),
  KEY `idx_wallet_ledger_entries_wallet_created` (`wallet_id`, `created_at`),
  KEY `idx_wallet_ledger_entries_user_created` (`user_id`, `created_at`),
  KEY `idx_wallet_ledger_entries_related` (`related_type`, `related_no`),
  CONSTRAINT `fk_wallet_ledger_entries_wallet` FOREIGN KEY (`wallet_id`) REFERENCES `wallet_accounts` (`id`),
  CONSTRAINT `fk_wallet_ledger_entries_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包余额变动流水';

CREATE TABLE IF NOT EXISTS `wallet_recharges` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '钱包充值ID',
  `recharge_no` VARCHAR(64) NOT NULL COMMENT '对外充值编号',
  `wallet_id` BIGINT UNSIGNED NOT NULL COMMENT '钱包账户ID',
  `wallet_no` VARCHAR(64) NOT NULL COMMENT '钱包编号快照',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `provider` VARCHAR(32) NOT NULL COMMENT '充值支付供应商：alipay/wechat',
  `method` VARCHAR(32) NOT NULL COMMENT '充值支付方式：alipay_page/alipay_wap/wechat_native/wechat_h5',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '充值状态：pending/paid/closed/failed',
  `client_token` VARCHAR(128) NOT NULL COMMENT '用户端充值幂等键',
  `amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '充值金额，单位分',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种，v1固定CNY',
  `upstream_trade_no` VARCHAR(128) NULL COMMENT '渠道交易号',
  `upstream_prepay_id` VARCHAR(128) NULL COMMENT '渠道预下单或会话标识',
  `qr_code_url` VARCHAR(1000) NULL COMMENT '扫码支付二维码内容或URL',
  `redirect_url` VARCHAR(1000) NULL COMMENT '网页或H5支付跳转URL',
  `callback_summary` JSON NULL COMMENT '回调摘要，不保存完整payload或签名串',
  `query_summary` JSON NULL COMMENT '主动查询摘要，不保存完整上游响应',
  `last_error_code` VARCHAR(64) NULL COMMENT '最近错误码',
  `last_error_message` VARCHAR(500) NULL COMMENT '最近错误说明',
  `expires_at` DATETIME(3) NOT NULL COMMENT '充值支付过期时间',
  `paid_at` DATETIME(3) NULL COMMENT '充值入账时间',
  `closed_at` DATETIME(3) NULL COMMENT '关闭时间',
  `failed_at` DATETIME(3) NULL COMMENT '失败时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wallet_recharges_recharge_no` (`recharge_no`),
  UNIQUE KEY `uk_wallet_recharges_idempotency` (`wallet_id`, `provider`, `method`, `client_token`),
  UNIQUE KEY `uk_wallet_recharges_upstream_trade` (`provider`, `upstream_trade_no`),
  KEY `idx_wallet_recharges_wallet_created` (`wallet_id`, `created_at`),
  KEY `idx_wallet_recharges_user_status` (`user_id`, `status`, `created_at`),
  KEY `idx_wallet_recharges_provider_status` (`provider`, `status`, `created_at`),
  CONSTRAINT `fk_wallet_recharges_wallet` FOREIGN KEY (`wallet_id`) REFERENCES `wallet_accounts` (`id`),
  CONSTRAINT `fk_wallet_recharges_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包充值记录';

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('wallet.enabled', 'false', 'bool', '钱包设置', 0, '是否开放用户端钱包、充值和余额支付入口'),
  ('wallet.recharge_min_cents', '100', 'int', '钱包设置', 0, '单笔钱包充值最小金额，单位分'),
  ('wallet.recharge_max_cents', '500000', 'int', '钱包设置', 0, '单笔钱包充值最大金额，单位分')
ON DUPLICATE KEY UPDATE
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.wallets', '钱包管理', 'menu', NULL, '/wallets', 'Wallet', 75, 1, '菜单', '显示钱包管理菜单和页面入口'),
  ('wallet:view', '查看钱包', 'action', 'page.wallets', NULL, NULL, 100, 0, '钱包管理', '查看用户钱包、余额、充值和流水')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `type` = VALUES(`type`),
  `parent_code` = VALUES(`parent_code`),
  `path` = VALUES(`path`),
  `icon` = VALUES(`icon`),
  `sort_order` = VALUES(`sort_order`),
  `visible_in_menu` = VALUES(`visible_in_menu`),
  `group_name` = VALUES(`group_name`),
  `description` = VALUES(`description`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT `admin_roles`.`id`, `admin_permissions`.`id`
FROM `admin_roles`
JOIN `admin_permissions`
WHERE `admin_roles`.`code` = 'super_admin'
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
