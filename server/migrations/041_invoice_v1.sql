-- Invoice v1 schema and permission seeds.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.
--
-- This migration re-opens the invoice domain as a manual electronic normal
-- invoice workflow. It creates invoice facts and permission seeds only; it
-- does not backfill old invoice data removed by historical cleanup.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `invoice_applications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '发票申请ID',
  `invoice_no` VARCHAR(64) NOT NULL COMMENT '对外发票申请编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `client_token` VARCHAR(128) NOT NULL COMMENT '用户端创建申请幂等键',
  `invoice_type` VARCHAR(32) NOT NULL DEFAULT 'electronic_normal' COMMENT '发票类型：electronic_normal',
  `title_type` VARCHAR(32) NOT NULL COMMENT '抬头类型：personal/company',
  `title` VARCHAR(100) NOT NULL COMMENT '发票抬头',
  `tax_no` VARCHAR(64) NULL COMMENT '纳税人识别号，企业抬头必填',
  `email` VARCHAR(128) NULL COMMENT '接收邮箱，v1仅保存不自动发送',
  `amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '申请开票金额，单位分',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种，v1固定CNY',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '申请状态：pending/processing/issued/rejected/cancelled',
  `remark` VARCHAR(500) NULL COMMENT '用户备注',
  `admin_note` VARCHAR(1000) NULL COMMENT '后台备注，不返回用户端',
  `reject_reason` VARCHAR(500) NULL COMMENT '驳回原因',
  `cancel_reason` VARCHAR(500) NULL COMMENT '用户取消原因',
  `invoice_code` VARCHAR(64) NULL COMMENT '发票代码，可选',
  `invoice_number` VARCHAR(128) NULL COMMENT '发票号码',
  `invoice_file_id` BIGINT UNSIGNED NULL COMMENT '发票PDF文件ID',
  `accepted_by_admin_id` BIGINT UNSIGNED NULL COMMENT '受理管理员ID',
  `rejected_by_admin_id` BIGINT UNSIGNED NULL COMMENT '驳回管理员ID',
  `issued_by_admin_id` BIGINT UNSIGNED NULL COMMENT '开票登记管理员ID',
  `accepted_at` DATETIME(3) NULL COMMENT '受理时间',
  `rejected_at` DATETIME(3) NULL COMMENT '驳回时间',
  `cancelled_at` DATETIME(3) NULL COMMENT '取消时间',
  `issued_at` DATETIME(3) NULL COMMENT '开票时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_invoice_applications_invoice_no` (`invoice_no`),
  UNIQUE KEY `uk_invoice_applications_idempotency` (`user_id`, `client_token`),
  KEY `idx_invoice_applications_user_status` (`user_id`, `status`, `created_at`),
  KEY `idx_invoice_applications_status_created` (`status`, `created_at`),
  KEY `idx_invoice_applications_title` (`title`),
  KEY `idx_invoice_applications_file` (`invoice_file_id`),
  CONSTRAINT `fk_invoice_applications_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_invoice_applications_file` FOREIGN KEY (`invoice_file_id`) REFERENCES `file_attachments` (`id`),
  CONSTRAINT `fk_invoice_applications_accepted_admin` FOREIGN KEY (`accepted_by_admin_id`) REFERENCES `admin_users` (`id`),
  CONSTRAINT `fk_invoice_applications_rejected_admin` FOREIGN KEY (`rejected_by_admin_id`) REFERENCES `admin_users` (`id`),
  CONSTRAINT `fk_invoice_applications_issued_admin` FOREIGN KEY (`issued_by_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='发票申请';

CREATE TABLE IF NOT EXISTS `invoice_application_orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '发票申请订单明细ID',
  `invoice_id` BIGINT UNSIGNED NOT NULL COMMENT '发票申请ID',
  `invoice_no` VARCHAR(64) NOT NULL COMMENT '发票申请编号快照',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `order_no` VARCHAR(64) NOT NULL COMMENT '订单编号快照',
  `order_type` VARCHAR(32) NOT NULL COMMENT '订单类型快照：purchase/renewal',
  `order_amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '订单开票金额，单位分',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种，v1固定CNY',
  `payment_status` VARCHAR(32) NOT NULL COMMENT '订单支付状态快照',
  `paid_at` DATETIME(3) NULL COMMENT '订单支付或人工确认时间快照',
  `product_name` VARCHAR(128) NULL COMMENT '产品名称快照',
  `plan_name` VARCHAR(128) NULL COMMENT '套餐名称快照',
  `status_snapshot` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '申请状态快照，用于订单占用唯一约束',
  `active_order_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `status_snapshot` IN ('pending', 'processing', 'issued') THEN `order_id`
      ELSE NULL
    END
  ) STORED COMMENT '有效发票申请订单占用投影，仅pending/processing/issued参与唯一约束',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_invoice_application_orders_invoice_order` (`invoice_id`, `order_id`),
  UNIQUE KEY `uk_invoice_application_orders_active_order` (`active_order_id`),
  KEY `idx_invoice_application_orders_order` (`order_id`, `created_at`),
  KEY `idx_invoice_application_orders_order_no` (`order_no`),
  KEY `idx_invoice_application_orders_user_created` (`user_id`, `created_at`),
  CONSTRAINT `fk_invoice_application_orders_invoice` FOREIGN KEY (`invoice_id`) REFERENCES `invoice_applications` (`id`),
  CONSTRAINT `fk_invoice_application_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_invoice_application_orders_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='发票申请关联订单';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.invoices', '发票运营', 'menu', NULL, '/invoices', 'Document', 78, 1, '菜单', '显示发票运营菜单和页面入口'),
  ('invoice:*', '发票全权限', 'action', 'page.invoices', NULL, NULL, 100, 0, '发票运营', '发票查看、受理、开票、驳回和后台备注全部能力'),
  ('invoice:view', '查看发票', 'action', 'page.invoices', NULL, NULL, 110, 0, '发票运营', '查看发票申请、订单明细和PDF'),
  ('invoice:update', '处理发票', 'action', 'page.invoices', NULL, NULL, 120, 0, '发票运营', '受理发票申请和维护后台备注'),
  ('invoice:issue', '登记开票', 'action', 'page.invoices', NULL, NULL, 130, 0, '发票运营', '登记发票号码并绑定PDF文件'),
  ('invoice:reject', '驳回发票', 'action', 'page.invoices', NULL, NULL, 140, 0, '发票运营', '驳回待处理或处理中的发票申请')
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
