-- Server product catalog schema.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `products` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '产品ID',
  `product_no` VARCHAR(64) NOT NULL COMMENT '对外产品编号',
  `type` VARCHAR(32) NOT NULL COMMENT '产品类型：server',
  `slug` VARCHAR(96) NOT NULL COMMENT 'Web 展示标识',
  `name` VARCHAR(128) NOT NULL COMMENT '产品名称',
  `summary` VARCHAR(255) NULL COMMENT '产品简介',
  `description` TEXT NULL COMMENT '产品详情',
  `status` VARCHAR(32) NOT NULL DEFAULT 'draft' COMMENT '产品状态：draft/active/inactive',
  `visible` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否 Web 展示',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_products_product_no` (`product_no`),
  UNIQUE KEY `uk_products_slug` (`slug`),
  KEY `idx_products_type_status` (`type`, `status`),
  KEY `idx_products_visible_sort` (`visible`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品目录';

CREATE TABLE IF NOT EXISTS `product_plans` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '套餐ID',
  `plan_no` VARCHAR(64) NOT NULL COMMENT '对外套餐编号',
  `product_id` BIGINT UNSIGNED NOT NULL COMMENT '所属产品ID',
  `code` VARCHAR(96) NOT NULL COMMENT '套餐编码',
  `name` VARCHAR(128) NOT NULL COMMENT '套餐名称',
  `summary` VARCHAR(255) NULL COMMENT '套餐简介',
  `cpu_cores` INT NOT NULL COMMENT 'CPU 核数',
  `memory_mb` INT NOT NULL COMMENT '内存 MB',
  `system_disk_gb` INT NOT NULL COMMENT '系统盘 GB',
  `data_disk_gb` INT NOT NULL DEFAULT 0 COMMENT '默认数据盘 GB',
  `bandwidth_mbps` INT NOT NULL COMMENT '默认带宽 Mbps',
  `traffic_gb` INT NULL COMMENT '月流量 GB，NULL 表示暂不承诺',
  `public_ip_count` INT NOT NULL DEFAULT 1 COMMENT '公网 IP 数',
  `virtualization` VARCHAR(32) NOT NULL DEFAULT 'kvm' COMMENT '虚拟化方式',
  `architecture` VARCHAR(32) NOT NULL DEFAULT 'x86_64' COMMENT 'CPU 架构',
  `is_featured` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否推荐',
  `status` VARCHAR(32) NOT NULL DEFAULT 'draft' COMMENT '套餐状态：draft/active/inactive/sold_out',
  `visible` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否 Web 展示',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_plans_plan_no` (`plan_no`),
  UNIQUE KEY `uk_product_plans_code` (`code`),
  KEY `idx_product_plans_product_status` (`product_id`, `status`),
  KEY `idx_product_plans_visible_sort` (`visible`, `sort_order`),
  CONSTRAINT `fk_product_plans_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务器套餐';

CREATE TABLE IF NOT EXISTS `plan_prices` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '价格ID',
  `plan_id` BIGINT UNSIGNED NOT NULL COMMENT '套餐ID',
  `billing_cycle` VARCHAR(32) NOT NULL COMMENT '计费周期：monthly/quarterly/semi_yearly/yearly',
  `price_cents` BIGINT UNSIGNED NOT NULL COMMENT '售价，单位分',
  `original_price_cents` BIGINT UNSIGNED NULL COMMENT '划线价，单位分',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '价格状态：active/inactive',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plan_prices_plan_cycle` (`plan_id`, `billing_cycle`),
  KEY `idx_plan_prices_status_sort` (`status`, `sort_order`),
  CONSTRAINT `fk_plan_prices_plan` FOREIGN KEY (`plan_id`) REFERENCES `product_plans` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='套餐周期价格';

CREATE TABLE IF NOT EXISTS `sales_regions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '销售地域ID',
  `region_no` VARCHAR(64) NOT NULL COMMENT '对外地域编号',
  `code` VARCHAR(64) NOT NULL COMMENT '地域编码',
  `name` VARCHAR(128) NOT NULL COMMENT '地域名称',
  `country` VARCHAR(64) NULL COMMENT '国家或地区',
  `city` VARCHAR(64) NULL COMMENT '城市',
  `summary` VARCHAR(255) NULL COMMENT '地域简介',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '地域状态：active/inactive',
  `visible` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否 Web 展示',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sales_regions_region_no` (`region_no`),
  UNIQUE KEY `uk_sales_regions_code` (`code`),
  KEY `idx_sales_regions_status_visible` (`status`, `visible`),
  KEY `idx_sales_regions_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售地域';

CREATE TABLE IF NOT EXISTS `server_os_templates` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '服务器系统模板ID',
  `template_no` VARCHAR(64) NOT NULL COMMENT '对外模板编号',
  `code` VARCHAR(96) NOT NULL COMMENT '模板编码',
  `name` VARCHAR(128) NOT NULL COMMENT '展示名称',
  `os_family` VARCHAR(32) NOT NULL COMMENT '系统族：linux/windows/bsd',
  `distribution` VARCHAR(64) NOT NULL COMMENT '发行版',
  `version` VARCHAR(64) NOT NULL COMMENT '版本号',
  `architecture` VARCHAR(32) NOT NULL DEFAULT 'x86_64' COMMENT 'CPU 架构',
  `summary` VARCHAR(255) NULL COMMENT '模板简介',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '模板状态：active/inactive',
  `visible` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否 Web 展示',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_server_os_templates_template_no` (`template_no`),
  UNIQUE KEY `uk_server_os_templates_code` (`code`),
  KEY `idx_server_os_templates_status_visible` (`status`, `visible`),
  KEY `idx_server_os_templates_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务器系统模板';

CREATE TABLE IF NOT EXISTS `plan_regions` (
  `plan_id` BIGINT UNSIGNED NOT NULL COMMENT '套餐ID',
  `region_id` BIGINT UNSIGNED NOT NULL COMMENT '销售地域ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '关联状态：active/inactive',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`plan_id`, `region_id`),
  KEY `idx_plan_regions_region` (`region_id`),
  CONSTRAINT `fk_plan_regions_plan` FOREIGN KEY (`plan_id`) REFERENCES `product_plans` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_plan_regions_region` FOREIGN KEY (`region_id`) REFERENCES `sales_regions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='套餐销售地域关联';

CREATE TABLE IF NOT EXISTS `plan_os_templates` (
  `plan_id` BIGINT UNSIGNED NOT NULL COMMENT '套餐ID',
  `template_id` BIGINT UNSIGNED NOT NULL COMMENT '服务器系统模板ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '关联状态：active/inactive',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`plan_id`, `template_id`),
  KEY `idx_plan_os_templates_template` (`template_id`),
  CONSTRAINT `fk_plan_os_templates_plan` FOREIGN KEY (`plan_id`) REFERENCES `product_plans` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_plan_os_templates_template` FOREIGN KEY (`template_id`) REFERENCES `server_os_templates` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='套餐服务器系统模板关联';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.products', '产品管理', 'menu', NULL, '/products', 'Box', 50, 1, '菜单', '显示产品管理菜单和页面入口'),
  ('product:*', '产品全权限', 'action', 'page.products', NULL, NULL, 100, 0, '产品管理', '产品目录模块全部能力'),
  ('product:view', '查看产品', 'action', 'page.products', NULL, NULL, 110, 0, '产品管理', '查看产品、套餐、价格、销售地域和系统模板'),
  ('product:create', '创建产品', 'action', 'page.products', NULL, NULL, 120, 0, '产品管理', '创建产品、套餐、销售地域和系统模板'),
  ('product:update', '编辑产品', 'action', 'page.products', NULL, NULL, 130, 0, '产品管理', '编辑产品、套餐、价格和关联关系'),
  ('product:publish', '发布产品', 'action', 'page.products', NULL, NULL, 140, 0, '产品管理', '上架、下架、售罄和恢复产品目录')
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

INSERT INTO `products` (`product_no`, `type`, `slug`, `name`, `summary`, `description`, `status`, `visible`, `sort_order`) VALUES
  ('PROD-SERVER-001', 'server', 'cloud-server', '云服务器', '固定规格云服务器产品目录', '用于展示固定套餐、周期价格、销售地域和服务器系统模板。当前阶段不包含订单、支付或实例开通。', 'active', 1, 10)
ON DUPLICATE KEY UPDATE
  `type` = VALUES(`type`),
  `slug` = VALUES(`slug`),
  `name` = VALUES(`name`),
  `summary` = VALUES(`summary`),
  `description` = VALUES(`description`),
  `status` = VALUES(`status`),
  `visible` = VALUES(`visible`),
  `sort_order` = VALUES(`sort_order`);

INSERT INTO `sales_regions` (`region_no`, `code`, `name`, `country`, `city`, `summary`, `status`, `visible`, `sort_order`) VALUES
  ('REG-HK-001', 'hk', '香港', '中国香港', '香港', '面向亚太访问优化的销售地域', 'active', 1, 10),
  ('REG-US-001', 'us-west', '美国西部', '美国', '洛杉矶', '面向海外访问的销售地域', 'active', 1, 20),
  ('REG-CN-001', 'cn-east', '华东', '中国', '上海', '面向国内业务的销售地域', 'active', 1, 30)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `country` = VALUES(`country`),
  `city` = VALUES(`city`),
  `summary` = VALUES(`summary`),
  `status` = VALUES(`status`),
  `visible` = VALUES(`visible`),
  `sort_order` = VALUES(`sort_order`);

INSERT INTO `server_os_templates` (`template_no`, `code`, `name`, `os_family`, `distribution`, `version`, `architecture`, `summary`, `status`, `visible`, `sort_order`) VALUES
  ('TPL-UBUNTU-2204', 'ubuntu-22-04', 'Ubuntu 22.04 LTS', 'linux', 'ubuntu', '22.04', 'x86_64', '长期支持版本，适合通用 Linux 服务', 'active', 1, 10),
  ('TPL-DEBIAN-12', 'debian-12', 'Debian 12', 'linux', 'debian', '12', 'x86_64', '稳定轻量的 Linux 发行版', 'active', 1, 20),
  ('TPL-CENTOS-STREAM-9', 'centos-stream-9', 'CentOS Stream 9', 'linux', 'centos-stream', '9', 'x86_64', '适合 RHEL 生态测试和服务部署', 'active', 1, 30),
  ('TPL-WINDOWS-2022', 'windows-server-2022', 'Windows Server 2022', 'windows', 'windows-server', '2022', 'x86_64', 'Windows 服务器系统模板展示项', 'active', 1, 40)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `os_family` = VALUES(`os_family`),
  `distribution` = VALUES(`distribution`),
  `version` = VALUES(`version`),
  `architecture` = VALUES(`architecture`),
  `summary` = VALUES(`summary`),
  `status` = VALUES(`status`),
  `visible` = VALUES(`visible`),
  `sort_order` = VALUES(`sort_order`);
