-- pveCloud initial database schema.
-- Target: MariaDB 11.4.9 / InnoDB / utf8mb4.

SET NAMES utf8mb4;

CREATE DATABASE IF NOT EXISTS `pvecloud`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` VARCHAR(64) NOT NULL COMMENT '用户名',
  `email` VARCHAR(191) NULL COMMENT '邮箱地址',
  `phone` VARCHAR(32) NULL COMMENT '手机号',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '用户状态',
  `email_verified_at` DATETIME(3) NULL COMMENT '邮箱验证时间',
  `phone_verified_at` DATETIME(3) NULL COMMENT '手机号验证时间',
  `last_login_at` DATETIME(3) NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NULL COMMENT '最后登录IP',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`),
  UNIQUE KEY `uk_users_email` (`email`),
  UNIQUE KEY `uk_users_phone` (`phone`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户账号';

CREATE TABLE IF NOT EXISTS `admin_users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '管理员ID',
  `username` VARCHAR(64) NOT NULL COMMENT '管理员用户名',
  `email` VARCHAR(191) NULL COMMENT '管理员邮箱',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `display_name` VARCHAR(64) NOT NULL COMMENT '显示名称',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '管理员状态',
  `last_login_at` DATETIME(3) NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NULL COMMENT '最后登录IP',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_users_username` (`username`),
  UNIQUE KEY `uk_admin_users_email` (`email`),
  KEY `idx_admin_users_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员账号';

CREATE TABLE IF NOT EXISTS `admin_roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `code` VARCHAR(64) NOT NULL COMMENT '角色编码',
  `name` VARCHAR(64) NOT NULL COMMENT '角色名称',
  `description` VARCHAR(255) NULL COMMENT '角色说明',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '角色状态',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_roles_code` (`code`),
  KEY `idx_admin_roles_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理端角色';

CREATE TABLE IF NOT EXISTS `admin_permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `code` VARCHAR(96) NOT NULL COMMENT '权限码',
  `name` VARCHAR(96) NOT NULL COMMENT '权限名称',
  `group_name` VARCHAR(64) NOT NULL COMMENT '权限分组',
  `description` VARCHAR(255) NULL COMMENT '权限说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_permissions_code` (`code`),
  KEY `idx_admin_permissions_group` (`group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理端权限码';

CREATE TABLE IF NOT EXISTS `admin_user_roles` (
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`admin_id`, `role_id`),
  CONSTRAINT `fk_admin_user_roles_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`),
  CONSTRAINT `fk_admin_user_roles_role` FOREIGN KEY (`role_id`) REFERENCES `admin_roles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员角色关联';

CREATE TABLE IF NOT EXISTS `admin_role_permissions` (
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`role_id`, `permission_id`),
  CONSTRAINT `fk_admin_role_permissions_role` FOREIGN KEY (`role_id`) REFERENCES `admin_roles` (`id`),
  CONSTRAINT `fk_admin_role_permissions_permission` FOREIGN KEY (`permission_id`) REFERENCES `admin_permissions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联';

CREATE TABLE IF NOT EXISTS `products` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '产品ID',
  `name` VARCHAR(96) NOT NULL COMMENT '产品名称',
  `slug` VARCHAR(96) NOT NULL COMMENT '产品标识',
  `description` TEXT NULL COMMENT '产品说明',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '产品状态',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_products_slug` (`slug`),
  KEY `idx_products_status_sort` (`status`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品系列';

CREATE TABLE IF NOT EXISTS `product_plans` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '套餐ID',
  `product_id` BIGINT UNSIGNED NOT NULL COMMENT '所属产品ID',
  `name` VARCHAR(96) NOT NULL COMMENT '套餐名称',
  `slug` VARCHAR(96) NOT NULL COMMENT '套餐标识',
  `cpu_cores` INT UNSIGNED NOT NULL COMMENT 'CPU核心数',
  `memory_mb` INT UNSIGNED NOT NULL COMMENT '内存大小MB',
  `disk_gb` INT UNSIGNED NOT NULL COMMENT '磁盘大小GB',
  `bandwidth_mbps` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '带宽Mbps',
  `traffic_gb` INT UNSIGNED NULL COMMENT '流量包GB，NULL表示不限',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '套餐状态',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_plans_slug` (`slug`),
  KEY `idx_product_plans_product_status` (`product_id`, `status`),
  CONSTRAINT `fk_product_plans_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品套餐规格';

CREATE TABLE IF NOT EXISTS `regions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '地域ID',
  `name` VARCHAR(96) NOT NULL COMMENT '地域名称',
  `code` VARCHAR(64) NOT NULL COMMENT '地域编码',
  `description` VARCHAR(255) NULL COMMENT '地域说明',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '地域状态',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_regions_code` (`code`),
  KEY `idx_regions_status_sort` (`status`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='地域';

CREATE TABLE IF NOT EXISTS `pve_nodes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'PVE节点ID',
  `region_id` BIGINT UNSIGNED NOT NULL COMMENT '所属地域ID',
  `name` VARCHAR(96) NOT NULL COMMENT '节点显示名称',
  `node_name` VARCHAR(96) NOT NULL COMMENT 'PVE节点名称',
  `endpoint` VARCHAR(255) NOT NULL COMMENT 'PVE接口地址',
  `credential_ref` VARCHAR(128) NULL COMMENT 'PVE凭据引用',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '节点状态',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `last_sync_at` DATETIME(3) NULL COMMENT '最后同步时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pve_nodes_node_name` (`node_name`),
  KEY `idx_pve_nodes_region_status` (`region_id`, `status`),
  CONSTRAINT `fk_pve_nodes_region` FOREIGN KEY (`region_id`) REFERENCES `regions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='PVE 节点';

CREATE TABLE IF NOT EXISTS `images` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '镜像ID',
  `name` VARCHAR(96) NOT NULL COMMENT '镜像名称',
  `slug` VARCHAR(96) NOT NULL COMMENT '镜像标识',
  `os_type` VARCHAR(64) NOT NULL COMMENT '操作系统类型',
  `version` VARCHAR(64) NULL COMMENT '系统版本',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '镜像状态',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_images_slug` (`slug`),
  KEY `idx_images_os_status` (`os_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统镜像';

CREATE TABLE IF NOT EXISTS `region_images` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '地域镜像ID',
  `region_id` BIGINT UNSIGNED NOT NULL COMMENT '地域ID',
  `node_id` BIGINT UNSIGNED NOT NULL COMMENT 'PVE节点ID',
  `image_id` BIGINT UNSIGNED NOT NULL COMMENT '镜像ID',
  `pve_template_id` VARCHAR(128) NOT NULL COMMENT '该节点上的PVE模板ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '地域镜像状态',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_region_images_node_image` (`node_id`, `image_id`),
  KEY `idx_region_images_region_status` (`region_id`, `status`, `sort_order`),
  KEY `idx_region_images_image_status` (`image_id`, `status`),
  CONSTRAINT `fk_region_images_region` FOREIGN KEY (`region_id`) REFERENCES `regions` (`id`),
  CONSTRAINT `fk_region_images_node` FOREIGN KEY (`node_id`) REFERENCES `pve_nodes` (`id`),
  CONSTRAINT `fk_region_images_image` FOREIGN KEY (`image_id`) REFERENCES `images` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='地域和节点可用镜像';

CREATE TABLE IF NOT EXISTS `plan_prices` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '价格ID',
  `plan_id` BIGINT UNSIGNED NOT NULL COMMENT '套餐ID',
  `region_id` BIGINT UNSIGNED NOT NULL COMMENT '地域ID',
  `billing_period` VARCHAR(32) NOT NULL COMMENT '计费周期',
  `price_cents` BIGINT UNSIGNED NOT NULL COMMENT '价格，单位分',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '价格状态',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plan_prices_plan_region_period` (`plan_id`, `region_id`, `billing_period`),
  KEY `idx_plan_prices_status` (`status`),
  CONSTRAINT `fk_plan_prices_plan` FOREIGN KEY (`plan_id`) REFERENCES `product_plans` (`id`),
  CONSTRAINT `fk_plan_prices_region` FOREIGN KEY (`region_id`) REFERENCES `regions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='套餐价格';

CREATE TABLE IF NOT EXISTS `orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `order_no` VARCHAR(64) NOT NULL COMMENT '订单号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `order_type` VARCHAR(32) NOT NULL COMMENT '订单类型，new或renew',
  `product_id` BIGINT UNSIGNED NULL COMMENT '产品ID',
  `plan_id` BIGINT UNSIGNED NULL COMMENT '套餐ID',
  `region_id` BIGINT UNSIGNED NULL COMMENT '地域ID',
  `image_id` BIGINT UNSIGNED NULL COMMENT '镜像ID',
  `instance_id` BIGINT UNSIGNED NULL COMMENT '关联实例ID，续费用',
  `billing_period` VARCHAR(32) NULL COMMENT '计费周期',
  `amount_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '订单原价，单位分',
  `discount_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '优惠金额，单位分',
  `payable_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '应付金额，单位分',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '订单状态',
  `expired_at` DATETIME(3) NULL COMMENT '支付过期时间',
  `paid_at` DATETIME(3) NULL COMMENT '支付完成时间',
  `completed_at` DATETIME(3) NULL COMMENT '订单完成时间',
  `cancelled_at` DATETIME(3) NULL COMMENT '订单取消时间',
  `remark` VARCHAR(500) NULL COMMENT '订单备注',
  `snapshot` JSON NULL COMMENT '下单时产品和价格快照',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_orders_order_no` (`order_no`),
  KEY `idx_orders_user_created` (`user_id`, `created_at`),
  KEY `idx_orders_status_expired` (`status`, `expired_at`),
  KEY `idx_orders_instance` (`instance_id`),
  CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_orders_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`),
  CONSTRAINT `fk_orders_plan` FOREIGN KEY (`plan_id`) REFERENCES `product_plans` (`id`),
  CONSTRAINT `fk_orders_region` FOREIGN KEY (`region_id`) REFERENCES `regions` (`id`),
  CONSTRAINT `fk_orders_image` FOREIGN KEY (`image_id`) REFERENCES `images` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单';

CREATE TABLE IF NOT EXISTS `payment_orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '支付单ID',
  `payment_no` VARCHAR(64) NOT NULL COMMENT '支付单号',
  `order_id` BIGINT UNSIGNED NULL COMMENT '关联订单ID，充值可为空',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `payment_scene` VARCHAR(32) NOT NULL COMMENT '支付场景，order或topup',
  `channel` VARCHAR(32) NOT NULL COMMENT '支付渠道：alipay/wechat/balance/manual',
  `amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '支付金额，单位分',
  `status` VARCHAR(32) NOT NULL DEFAULT 'created' COMMENT '支付状态',
  `third_trade_no` VARCHAR(128) NULL COMMENT '第三方交易号',
  `client_ip` VARCHAR(64) NULL COMMENT '客户端IP',
  `paid_at` DATETIME(3) NULL COMMENT '支付成功时间',
  `closed_at` DATETIME(3) NULL COMMENT '支付关闭时间',
  `refunded_at` DATETIME(3) NULL COMMENT '退款完成时间',
  `third_payload` JSON NULL COMMENT '第三方支付回包摘要',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_orders_payment_no` (`payment_no`),
  UNIQUE KEY `uk_payment_orders_channel_trade` (`channel`, `third_trade_no`),
  KEY `idx_payment_orders_user_created` (`user_id`, `created_at`),
  KEY `idx_payment_orders_order` (`order_id`),
  KEY `idx_payment_orders_status_created` (`status`, `created_at`),
  CONSTRAINT `fk_payment_orders_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_payment_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付单';

CREATE TABLE IF NOT EXISTS `payment_notify_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '回调日志ID',
  `channel` VARCHAR(32) NOT NULL COMMENT '支付渠道：alipay/wechat/balance/manual',
  `payment_no` VARCHAR(64) NULL COMMENT '本地支付单号',
  `third_trade_no` VARCHAR(128) NULL COMMENT '第三方交易号',
  `notify_id` VARCHAR(128) NULL COMMENT '第三方回调唯一ID',
  `amount_cents` BIGINT UNSIGNED NULL COMMENT '回调金额，单位分',
  `verify_status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '验签状态',
  `handled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已处理',
  `raw_payload` JSON NULL COMMENT '回调原始内容',
  `received_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '接收时间',
  `handled_at` DATETIME(3) NULL COMMENT '处理完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_notify_channel_notify` (`channel`, `notify_id`),
  KEY `idx_payment_notify_payment_no` (`payment_no`),
  KEY `idx_payment_notify_trade_no` (`third_trade_no`),
  KEY `idx_payment_notify_received` (`received_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付回调日志';

CREATE TABLE IF NOT EXISTS `wallet_accounts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '余额账户ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `balance_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '可用余额，单位分',
  `frozen_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '冻结余额，单位分',
  `total_recharged_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计充值金额，单位分',
  `total_consumed_cents` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '累计消费金额，单位分',
  `version` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wallet_accounts_user` (`user_id`),
  CONSTRAINT `fk_wallet_accounts_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户余额账户';

CREATE TABLE IF NOT EXISTS `wallet_transactions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '钱包流水ID',
  `transaction_no` VARCHAR(64) NOT NULL COMMENT '流水号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `wallet_account_id` BIGINT UNSIGNED NOT NULL COMMENT '余额账户ID',
  `related_order_id` BIGINT UNSIGNED NULL COMMENT '关联订单ID',
  `related_payment_id` BIGINT UNSIGNED NULL COMMENT '关联支付单ID',
  `operator_admin_id` BIGINT UNSIGNED NULL COMMENT '后台操作者ID',
  `direction` VARCHAR(16) NOT NULL COMMENT '资金方向，in或out',
  `tx_type` VARCHAR(32) NOT NULL COMMENT '流水类型',
  `amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '变动金额，单位分',
  `balance_after_cents` BIGINT UNSIGNED NOT NULL COMMENT '变动后余额，单位分',
  `status` VARCHAR(32) NOT NULL DEFAULT 'success' COMMENT '流水状态',
  `description` VARCHAR(255) NULL COMMENT '流水说明',
  `metadata` JSON NULL COMMENT '扩展数据',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wallet_transactions_no` (`transaction_no`),
  KEY `idx_wallet_transactions_user_created` (`user_id`, `created_at`),
  KEY `idx_wallet_transactions_account_created` (`wallet_account_id`, `created_at`),
  KEY `idx_wallet_transactions_order` (`related_order_id`),
  KEY `idx_wallet_transactions_payment` (`related_payment_id`),
  CONSTRAINT `fk_wallet_transactions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_wallet_transactions_account` FOREIGN KEY (`wallet_account_id`) REFERENCES `wallet_accounts` (`id`),
  CONSTRAINT `fk_wallet_transactions_order` FOREIGN KEY (`related_order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_wallet_transactions_payment` FOREIGN KEY (`related_payment_id`) REFERENCES `payment_orders` (`id`),
  CONSTRAINT `fk_wallet_transactions_admin_user` FOREIGN KEY (`operator_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='钱包流水';

CREATE TABLE IF NOT EXISTS `instances` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '实例ID',
  `instance_no` VARCHAR(64) NOT NULL COMMENT '实例编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `order_id` BIGINT UNSIGNED NULL COMMENT '来源订单ID',
  `plan_id` BIGINT UNSIGNED NULL COMMENT '套餐ID',
  `region_id` BIGINT UNSIGNED NULL COMMENT '地域ID',
  `image_id` BIGINT UNSIGNED NULL COMMENT '镜像ID',
  `node_id` BIGINT UNSIGNED NULL COMMENT 'PVE节点ID',
  `name` VARCHAR(96) NOT NULL COMMENT '实例名称',
  `status` VARCHAR(32) NOT NULL DEFAULT 'creating' COMMENT '实例状态',
  `vmid` BIGINT UNSIGNED NULL COMMENT 'PVE VMID',
  `provisioning_key` VARCHAR(128) NULL COMMENT '开通幂等键',
  `pve_task_upid` VARCHAR(255) NULL COMMENT 'PVE任务UPID',
  `cpu_cores` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'CPU核心数',
  `memory_mb` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '内存大小MB',
  `disk_gb` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '磁盘大小GB',
  `bandwidth_mbps` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '带宽Mbps',
  `ipv4` VARCHAR(64) NULL COMMENT 'IPv4地址',
  `ipv6` VARCHAR(128) NULL COMMENT 'IPv6地址',
  `expire_at` DATETIME(3) NULL COMMENT '到期时间',
  `last_sync_at` DATETIME(3) NULL COMMENT '最后同步时间',
  `metadata` JSON NULL COMMENT '实例扩展信息',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_instances_no` (`instance_no`),
  UNIQUE KEY `uk_instances_order` (`order_id`),
  UNIQUE KEY `uk_instances_node_vmid` (`node_id`, `vmid`),
  UNIQUE KEY `uk_instances_provisioning_key` (`provisioning_key`),
  KEY `idx_instances_user_status` (`user_id`, `status`),
  KEY `idx_instances_expire_at` (`expire_at`),
  CONSTRAINT `fk_instances_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_instances_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_instances_plan` FOREIGN KEY (`plan_id`) REFERENCES `product_plans` (`id`),
  CONSTRAINT `fk_instances_region` FOREIGN KEY (`region_id`) REFERENCES `regions` (`id`),
  CONSTRAINT `fk_instances_image` FOREIGN KEY (`image_id`) REFERENCES `images` (`id`),
  CONSTRAINT `fk_instances_node` FOREIGN KEY (`node_id`) REFERENCES `pve_nodes` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云服务器实例';

CREATE TABLE IF NOT EXISTS `async_tasks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `task_no` VARCHAR(64) NOT NULL COMMENT '任务编号',
  `task_type` VARCHAR(64) NOT NULL COMMENT '任务类型',
  `idempotency_key` VARCHAR(160) NOT NULL COMMENT '任务幂等键',
  `biz_type` VARCHAR(64) NOT NULL COMMENT '业务对象类型',
  `biz_id` BIGINT UNSIGNED NOT NULL COMMENT '业务对象ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '任务状态',
  `priority` INT NOT NULL DEFAULT 0 COMMENT '任务优先级',
  `retry_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '已重试次数',
  `max_retries` INT UNSIGNED NOT NULL DEFAULT 5 COMMENT '最大重试次数',
  `run_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '下次执行时间',
  `locked_by` VARCHAR(128) NULL COMMENT '锁定Worker标识',
  `locked_until` DATETIME(3) NULL COMMENT '锁定到期时间',
  `last_error` TEXT NULL COMMENT '最近一次错误',
  `payload` JSON NULL COMMENT '任务输入数据',
  `result` JSON NULL COMMENT '任务执行结果',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `finished_at` DATETIME(3) NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_async_tasks_task_no` (`task_no`),
  UNIQUE KEY `uk_async_tasks_idempotency` (`idempotency_key`),
  KEY `idx_async_tasks_status_run` (`status`, `run_at`, `priority`),
  KEY `idx_async_tasks_locked_until` (`locked_until`),
  KEY `idx_async_tasks_biz` (`biz_type`, `biz_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='异步任务';

CREATE TABLE IF NOT EXISTS `tickets` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单ID',
  `ticket_no` VARCHAR(64) NOT NULL COMMENT '工单编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `title` VARCHAR(160) NOT NULL COMMENT '工单标题',
  `category` VARCHAR(64) NULL COMMENT '工单分类',
  `priority` VARCHAR(32) NOT NULL DEFAULT 'normal' COMMENT '工单优先级',
  `status` VARCHAR(32) NOT NULL DEFAULT 'open' COMMENT '工单状态',
  `last_reply_by` VARCHAR(32) NULL COMMENT '最后回复方',
  `last_reply_at` DATETIME(3) NULL COMMENT '最后回复时间',
  `closed_at` DATETIME(3) NULL COMMENT '关闭时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tickets_no` (`ticket_no`),
  KEY `idx_tickets_user_created` (`user_id`, `created_at`),
  KEY `idx_tickets_status_updated` (`status`, `updated_at`),
  CONSTRAINT `fk_tickets_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单';

CREATE TABLE IF NOT EXISTS `ticket_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `ticket_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `sender_type` VARCHAR(32) NOT NULL COMMENT '发送方类型',
  `user_id` BIGINT UNSIGNED NULL COMMENT '用户ID',
  `admin_id` BIGINT UNSIGNED NULL COMMENT '管理员ID',
  `content` TEXT NOT NULL COMMENT '消息内容',
  `attachments` JSON NULL COMMENT '附件列表',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_ticket_messages_ticket_created` (`ticket_id`, `created_at`),
  CONSTRAINT `fk_ticket_messages_ticket` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`),
  CONSTRAINT `fk_ticket_messages_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_ticket_messages_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单消息';

CREATE TABLE IF NOT EXISTS `system_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `config_key` VARCHAR(128) NOT NULL COMMENT '配置键',
  `config_value` TEXT NULL COMMENT '配置值',
  `value_type` VARCHAR(32) NOT NULL DEFAULT 'string' COMMENT '值类型',
  `group_name` VARCHAR(64) NOT NULL COMMENT '配置分组',
  `is_secret` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否敏感配置',
  `description` VARCHAR(255) NULL COMMENT '配置说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_configs_key` (`config_key`),
  KEY `idx_system_configs_group` (`group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置';

CREATE TABLE IF NOT EXISTS `admin_audit_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '审计日志ID',
  `admin_id` BIGINT UNSIGNED NULL COMMENT '管理员ID',
  `action` VARCHAR(96) NOT NULL COMMENT '操作动作',
  `object_type` VARCHAR(64) NOT NULL COMMENT '操作对象类型',
  `object_id` VARCHAR(64) NULL COMMENT '操作对象ID',
  `before_data` JSON NULL COMMENT '操作前数据，需脱敏',
  `after_data` JSON NULL COMMENT '操作后数据，需脱敏',
  `ip` VARCHAR(64) NULL COMMENT '操作IP',
  `user_agent` VARCHAR(500) NULL COMMENT '浏览器User-Agent',
  `remark` VARCHAR(500) NULL COMMENT '操作备注',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_audit_logs_admin_created` (`admin_id`, `created_at`),
  KEY `idx_admin_audit_logs_object` (`object_type`, `object_id`),
  KEY `idx_admin_audit_logs_action_created` (`action`, `created_at`),
  CONSTRAINT `fk_admin_audit_logs_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台操作审计日志';

INSERT INTO `admin_permissions` (`code`, `name`, `group_name`, `description`) VALUES
  ('dashboard:view', '查看仪表盘', 'dashboard', '查看后台仪表盘统计'),
  ('user:view', '查看用户', 'user', '查看用户列表和详情'),
  ('user:update', '修改用户', 'user', '修改用户资料和状态'),
  ('product:create', '创建产品', 'product', '创建产品、套餐、地域和镜像'),
  ('product:update', '修改产品', 'product', '修改产品、套餐、地域和镜像'),
  ('order:view', '查看订单', 'order', '查看订单列表和详情'),
  ('order:cancel', '取消订单', 'order', '后台强制取消订单'),
  ('payment:view', '查看支付', 'payment', '查看支付单和钱包流水'),
  ('payment:manual_credit', '人工入账', 'payment', '后台人工入账或扣减余额'),
  ('instance:view', '查看实例', 'instance', '查看实例列表和详情'),
  ('instance:operate', '操作实例', 'instance', '后台强制操作实例'),
  ('ticket:reply', '回复工单', 'ticket', '回复和关闭工单'),
  ('admin:manage', '管理员管理', 'admin', '管理管理员、角色和权限'),
  ('system:update', '修改系统配置', 'system', '修改站点、支付和 PVE 配置'),
  ('audit:view', '查看审计日志', 'audit', '查看后台操作审计日志')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `group_name` = VALUES(`group_name`),
  `description` = VALUES(`description`);

INSERT INTO `admin_roles` (`code`, `name`, `description`, `status`) VALUES
  ('super_admin', '超级管理员', '拥有全部后台权限', 'active')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `description` = VALUES(`description`),
  `status` = VALUES(`status`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT `admin_roles`.`id`, `admin_permissions`.`id`
FROM `admin_roles`
JOIN `admin_permissions`
WHERE `admin_roles`.`code` = 'super_admin'
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('site.name', 'pveCloud', 'string', 'site', 0, '站点名称'),
  ('order.expire_minutes', '30', 'int', 'order', 0, '待支付订单过期分钟数'),
  ('payment.enabled_channels', '[]', 'json', 'payment', 0, '启用的支付渠道'),
  ('pve.default_region_id', NULL, 'int', 'pve', 0, '默认 PVE 地域 ID'),
  ('notify.email_enabled', 'false', 'bool', 'notify', 0, '是否启用邮件通知'),
  ('notify.sms_enabled', 'false', 'bool', 'notify', 0, '是否启用短信通知')
ON DUPLICATE KEY UPDATE
  `config_value` = VALUES(`config_value`),
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

-- Default admin accounts are intentionally not inserted here.
-- Create the first admin through a setup command so no default password is stored in git.
