-- Real payment phase 1 schema, permissions and system config seeds.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.
-- This migration is intentionally idempotent: it creates new payment facts,
-- broadens order status comments for historical rows, and seeds permissions/configs.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @orders_status_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'status'
);
SET @modify_orders_status_for_payment_sql := IF(
  @orders_status_column_exists = 1,
  'ALTER TABLE `orders` MODIFY COLUMN `status` VARCHAR(32) NOT NULL DEFAULT ''pending'' COMMENT ''订单状态：pending/provisioning/fulfilled/error/cancelled/closed''',
  'SELECT 1'
);
PREPARE modify_orders_status_for_payment_stmt FROM @modify_orders_status_for_payment_sql;
EXECUTE modify_orders_status_for_payment_stmt;
DEALLOCATE PREPARE modify_orders_status_for_payment_stmt;

SET @orders_payment_status_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'payment_status'
);
SET @modify_orders_payment_status_for_payment_sql := IF(
  @orders_payment_status_column_exists = 1,
  'ALTER TABLE `orders` MODIFY COLUMN `payment_status` VARCHAR(32) NOT NULL DEFAULT ''unpaid'' COMMENT ''支付状态：unpaid/paid/manual_confirmed/refunded''',
  'SELECT 1'
);
PREPARE modify_orders_payment_status_for_payment_stmt FROM @modify_orders_payment_status_for_payment_sql;
EXECUTE modify_orders_payment_status_for_payment_stmt;
DEALLOCATE PREPARE modify_orders_payment_status_for_payment_stmt;

CREATE TABLE IF NOT EXISTS `payment_transactions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '支付交易ID',
  `payment_no` VARCHAR(64) NOT NULL COMMENT '对外支付编号',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `order_no` VARCHAR(64) NOT NULL COMMENT '订单编号快照',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `provider` VARCHAR(32) NOT NULL COMMENT '支付供应商：alipay/wechat',
  `method` VARCHAR(32) NOT NULL COMMENT '支付方式：alipay_page/alipay_wap/wechat_native/wechat_h5',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '交易状态：pending/paid/closed/failed/refunded',
  `client_token` VARCHAR(128) NOT NULL COMMENT '用户端支付幂等键',
  `amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '支付金额，单位分',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种，一期固定CNY',
  `upstream_trade_no` VARCHAR(128) NULL COMMENT '渠道交易号',
  `upstream_prepay_id` VARCHAR(128) NULL COMMENT '渠道预下单或会话标识',
  `qr_code_url` VARCHAR(1000) NULL COMMENT '扫码支付二维码内容或URL',
  `redirect_url` VARCHAR(1000) NULL COMMENT '网页或H5支付跳转URL',
  `callback_summary` JSON NULL COMMENT '回调摘要，不保存完整payload或签名串',
  `query_summary` JSON NULL COMMENT '主动查询摘要，不保存完整上游响应',
  `last_error_code` VARCHAR(64) NULL COMMENT '最近错误码',
  `last_error_message` VARCHAR(500) NULL COMMENT '最近错误说明',
  `expires_at` DATETIME(3) NOT NULL COMMENT '支付过期时间',
  `paid_at` DATETIME(3) NULL COMMENT '支付完成时间',
  `closed_at` DATETIME(3) NULL COMMENT '关闭时间',
  `failed_at` DATETIME(3) NULL COMMENT '失败时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_transactions_payment_no` (`payment_no`),
  UNIQUE KEY `uk_payment_transactions_idempotency` (`order_id`, `provider`, `method`, `client_token`),
  UNIQUE KEY `uk_payment_transactions_upstream_trade` (`provider`, `upstream_trade_no`),
  KEY `idx_payment_transactions_order` (`order_id`, `created_at`),
  KEY `idx_payment_transactions_user_status` (`user_id`, `status`, `created_at`),
  KEY `idx_payment_transactions_provider_status` (`provider`, `status`, `created_at`),
  KEY `idx_payment_transactions_expires` (`status`, `expires_at`),
  CONSTRAINT `fk_payment_transactions_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_payment_transactions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付交易流水';

CREATE TABLE IF NOT EXISTS `refund_transactions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '退款交易ID',
  `refund_no` VARCHAR(64) NOT NULL COMMENT '对外退款编号',
  `payment_id` BIGINT UNSIGNED NOT NULL COMMENT '支付交易ID',
  `payment_no` VARCHAR(64) NOT NULL COMMENT '支付编号快照',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `order_no` VARCHAR(64) NOT NULL COMMENT '订单编号快照',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `provider` VARCHAR(32) NOT NULL COMMENT '支付供应商：alipay/wechat',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '退款状态：pending/succeeded/failed',
  `amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '退款金额，单位分，一期必须等于原支付金额',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种，一期固定CNY',
  `reason` VARCHAR(500) NOT NULL COMMENT '退款原因',
  `requested_by_admin_id` BIGINT UNSIGNED NOT NULL COMMENT '发起退款管理员ID',
  `upstream_refund_no` VARCHAR(128) NULL COMMENT '渠道退款号',
  `upstream_trade_no` VARCHAR(128) NULL COMMENT '渠道原交易号',
  `callback_summary` JSON NULL COMMENT '退款回调摘要，不保存完整payload或签名串',
  `query_summary` JSON NULL COMMENT '退款查询摘要，不保存完整上游响应',
  `last_error_code` VARCHAR(64) NULL COMMENT '最近错误码',
  `last_error_message` VARCHAR(500) NULL COMMENT '最近错误说明',
  `channel_confirmed_at` DATETIME(3) NULL COMMENT '渠道确认退款成功时间',
  `completed_at` DATETIME(3) NULL COMMENT '本地退款处理完成时间',
  `failed_at` DATETIME(3) NULL COMMENT '退款失败时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refund_transactions_refund_no` (`refund_no`),
  UNIQUE KEY `uk_refund_transactions_payment` (`payment_id`),
  UNIQUE KEY `uk_refund_transactions_upstream_refund` (`provider`, `upstream_refund_no`),
  KEY `idx_refund_transactions_order` (`order_id`, `created_at`),
  KEY `idx_refund_transactions_user_status` (`user_id`, `status`, `created_at`),
  KEY `idx_refund_transactions_provider_status` (`provider`, `status`, `created_at`),
  CONSTRAINT `fk_refund_transactions_payment` FOREIGN KEY (`payment_id`) REFERENCES `payment_transactions` (`id`),
  CONSTRAINT `fk_refund_transactions_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_refund_transactions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_refund_transactions_admin` FOREIGN KEY (`requested_by_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='退款交易流水';

CREATE TABLE IF NOT EXISTS `payment_effects` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '支付生效记录ID',
  `effect_no` VARCHAR(64) NOT NULL COMMENT '对外生效记录编号',
  `payment_id` BIGINT UNSIGNED NOT NULL COMMENT '支付交易ID',
  `payment_no` VARCHAR(64) NOT NULL COMMENT '支付编号快照',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `order_no` VARCHAR(64) NOT NULL COMMENT '订单编号快照',
  `order_type` VARCHAR(32) NOT NULL COMMENT '订单类型：purchase/renewal',
  `effect_type` VARCHAR(32) NOT NULL COMMENT '生效类型：purchase_instance/renewal_extension',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '状态：active/reverted',
  `instance_id` BIGINT UNSIGNED NULL COMMENT '关联实例ID',
  `instance_no` VARCHAR(64) NULL COMMENT '关联实例编号快照',
  `before_expires_at` DATETIME(3) NULL COMMENT '续费前到期时间',
  `after_expires_at` DATETIME(3) NULL COMMENT '续费后到期时间',
  `refund_id` BIGINT UNSIGNED NULL COMMENT '回滚来源退款ID',
  `refund_no` VARCHAR(64) NULL COMMENT '回滚来源退款编号快照',
  `applied_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '生效时间',
  `reverted_at` DATETIME(3) NULL COMMENT '回滚时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_effects_effect_no` (`effect_no`),
  UNIQUE KEY `uk_payment_effects_payment` (`payment_id`),
  KEY `idx_payment_effects_order` (`order_id`, `created_at`),
  KEY `idx_payment_effects_instance` (`instance_id`, `status`),
  KEY `idx_payment_effects_refund` (`refund_id`),
  CONSTRAINT `fk_payment_effects_payment` FOREIGN KEY (`payment_id`) REFERENCES `payment_transactions` (`id`),
  CONSTRAINT `fk_payment_effects_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_payment_effects_instance` FOREIGN KEY (`instance_id`) REFERENCES `instances` (`id`),
  CONSTRAINT `fk_payment_effects_refund` FOREIGN KEY (`refund_id`) REFERENCES `refund_transactions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付生效记录';

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('payment.enabled', 'false', 'bool', '支付设置', 0, '是否开放用户端真实支付入口'),
  ('payment.default_expire_minutes', '30', 'int', '支付设置', 0, '支付交易默认过期时间，单位分钟'),
  ('payment.callback_base_url', '', 'string', '支付设置', 0, '支付公开回调基础地址'),
  ('payment.alipay.enabled', 'false', 'bool', '支付设置', 0, '是否启用支付宝支付'),
  ('payment.alipay.app_id', '', 'string', '支付设置', 0, '支付宝应用ID'),
  ('payment.alipay.gateway_url', 'https://openapi.alipay.com/gateway.do', 'string', '支付设置', 0, '支付宝网关地址'),
  ('payment.alipay.app_private_key', '', 'string', '支付设置', 1, '支付宝应用私钥'),
  ('payment.alipay.alipay_public_key', '', 'string', '支付设置', 1, '支付宝公钥'),
  ('payment.alipay.notify_url', '', 'string', '支付设置', 0, '支付宝支付异步通知地址'),
  ('payment.alipay.return_url', '', 'string', '支付设置', 0, '支付宝支付同步返回地址'),
  ('payment.wechat.enabled', 'false', 'bool', '支付设置', 0, '是否启用微信支付'),
  ('payment.wechat.app_id', '', 'string', '支付设置', 0, '微信支付应用ID'),
  ('payment.wechat.mch_id', '', 'string', '支付设置', 0, '微信支付商户号'),
  ('payment.wechat.api_v3_key', '', 'string', '支付设置', 1, '微信支付 API v3 key'),
  ('payment.wechat.mch_private_key', '', 'string', '支付设置', 1, '微信支付商户私钥'),
  ('payment.wechat.mch_certificate_serial_no', '', 'string', '支付设置', 0, '微信支付商户证书序列号'),
  ('payment.wechat.platform_public_key_id', '', 'string', '支付设置', 0, '微信支付平台公钥ID，使用平台公钥模式时必填'),
  ('payment.wechat.platform_public_key', '', 'string', '支付设置', 1, '微信支付平台公钥或证书内容'),
  ('payment.wechat.notify_url', '', 'string', '支付设置', 0, '微信支付异步通知地址'),
  ('payment.wechat.h5_scene_info', '', 'string', '支付设置', 0, '微信 H5 支付场景信息 JSON 摘要')
ON DUPLICATE KEY UPDATE
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.payments', '支付管理', 'menu', NULL, '/payments', 'Card', 70, 1, '菜单', '显示支付管理菜单和页面入口'),
  ('payment:*', '支付全权限', 'action', 'page.payments', NULL, NULL, 100, 0, '支付管理', '支付流水、退款、同步和交付重试全部能力'),
  ('payment:view', '查看支付', 'action', 'page.payments', NULL, NULL, 110, 0, '支付管理', '查看支付和退款流水'),
  ('payment:refund', '发起退款', 'action', 'page.payments', NULL, NULL, 120, 0, '支付管理', '为已支付交易发起全额退款'),
  ('payment:sync', '同步支付', 'action', 'page.payments', NULL, NULL, 130, 0, '支付管理', '主动同步支付或退款渠道状态'),
  ('payment:retry-provision', '重试交付', 'action', 'page.payments', NULL, NULL, 140, 0, '支付管理', '重试真实支付后自动交付失败的新购订单')
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
