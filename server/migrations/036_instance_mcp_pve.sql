-- Instance delivery schema backed by MCP PVE client API.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @orders_status_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'status'
);
SET @modify_orders_status_sql := IF(
  @orders_status_column_exists = 1,
  'ALTER TABLE `orders` MODIFY COLUMN `status` VARCHAR(32) NOT NULL DEFAULT ''pending'' COMMENT ''订单状态：pending/provisioning/fulfilled/cancelled/closed''',
  'SELECT 1'
);
PREPARE modify_orders_status_stmt FROM @modify_orders_status_sql;
EXECUTE modify_orders_status_stmt;
DEALLOCATE PREPARE modify_orders_status_stmt;

CREATE TABLE IF NOT EXISTS `instance_provision_mappings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '交付映射ID',
  `mapping_no` VARCHAR(64) NOT NULL COMMENT '对外交付映射编号',
  `product_no` VARCHAR(64) NULL COMMENT '产品编号，NULL 表示不限定产品',
  `plan_no` VARCHAR(64) NOT NULL COMMENT '套餐编号',
  `region_no` VARCHAR(64) NOT NULL COMMENT '销售地域编号',
  `template_no` VARCHAR(64) NOT NULL COMMENT '系统模板编号',
  `network_type_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '网络类型编号，空字符串表示不限定网络类型',
  `node` VARCHAR(128) NOT NULL COMMENT 'MCP/PVE 节点名称',
  `storage` VARCHAR(128) NOT NULL COMMENT 'MCP/PVE 目标存储池',
  `disk_source` VARCHAR(255) NOT NULL COMMENT 'MCP/PVE 导入磁盘来源，格式 storage:path',
  `disk_format` VARCHAR(32) NULL COMMENT '源磁盘格式，如 qcow2/raw',
  `disk_interface` VARCHAR(32) NULL COMMENT '磁盘接口类型，如 scsi0/virtio0',
  `snippets_storage` VARCHAR(128) NULL COMMENT 'cloud-init snippets 存储名称',
  `ci_user` VARCHAR(64) NULL COMMENT 'CloudInit 默认登录用户',
  `ssh_keys` TEXT NULL COMMENT 'CloudInit SSH 公钥，多个公钥换行分隔',
  `ip_config0` VARCHAR(255) NULL COMMENT 'CloudInit ipconfig0 配置',
  `nameserver` VARCHAR(128) NULL COMMENT 'CloudInit DNS 服务器',
  `search_domain` VARCHAR(128) NULL COMMENT 'CloudInit DNS 搜索域',
  `ci_packages` TEXT NULL COMMENT 'CloudInit 首次开机安装包 JSON 数组',
  `apt_mirror` VARCHAR(255) NULL COMMENT 'CloudInit APT 镜像源',
  `vmid_start` INT UNSIGNED NOT NULL COMMENT '本映射可分配 VMID 起始值',
  `vmid_end` INT UNSIGNED NOT NULL COMMENT '本映射可分配 VMID 结束值',
  `next_vmid` INT UNSIGNED NOT NULL COMMENT '下一个待分配 VMID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '映射状态：active/inactive',
  `remark` VARCHAR(500) NULL COMMENT '后台备注',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_instance_provision_mappings_mapping_no` (`mapping_no`),
  UNIQUE KEY `uk_instance_provision_mappings_scope` (`plan_no`, `region_no`, `template_no`, `network_type_no`, `status`),
  KEY `idx_instance_provision_mappings_product` (`product_no`),
  KEY `idx_instance_provision_mappings_node` (`node`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实例交付映射';

CREATE TABLE IF NOT EXISTS `instances` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '实例ID',
  `instance_no` VARCHAR(64) NOT NULL COMMENT '对外实例编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '来源订单ID',
  `order_no` VARCHAR(64) NOT NULL COMMENT '来源订单号快照',
  `status` VARCHAR(32) NOT NULL DEFAULT 'creating' COMMENT '实例状态：creating/running/stopped/error/releasing/released',
  `product_no` VARCHAR(64) NOT NULL COMMENT '产品编号快照',
  `product_name` VARCHAR(128) NOT NULL COMMENT '产品名称快照',
  `plan_no` VARCHAR(64) NOT NULL COMMENT '套餐编号快照',
  `plan_name` VARCHAR(128) NOT NULL COMMENT '套餐名称快照',
  `cpu_cores` INT NOT NULL COMMENT 'CPU 核数快照',
  `memory_mb` INT NOT NULL COMMENT '内存 MB 快照',
  `system_disk_gb` INT NOT NULL COMMENT '系统盘 GB 快照',
  `data_disk_gb` INT NOT NULL DEFAULT 0 COMMENT '数据盘 GB 快照',
  `bandwidth_mbps` INT NOT NULL COMMENT '带宽 Mbps 快照',
  `region_no` VARCHAR(64) NOT NULL COMMENT '销售地域编号快照',
  `region_name` VARCHAR(128) NOT NULL COMMENT '销售地域名称快照',
  `network_type_no` VARCHAR(64) NULL COMMENT '网络类型编号快照',
  `network_type_name` VARCHAR(128) NULL COMMENT '网络类型名称快照',
  `template_no` VARCHAR(64) NOT NULL COMMENT '系统模板编号快照',
  `template_name` VARCHAR(128) NOT NULL COMMENT '系统模板名称快照',
  `os_family` VARCHAR(32) NOT NULL COMMENT '系统族快照',
  `os_distribution` VARCHAR(64) NOT NULL COMMENT '系统发行版快照',
  `os_version` VARCHAR(64) NOT NULL COMMENT '系统版本快照',
  `external_node` VARCHAR(128) NOT NULL COMMENT 'MCP/PVE 节点名称，仅管理端和服务端内部使用',
  `external_vmid` INT UNSIGNED NOT NULL COMMENT 'MCP/PVE VMID，仅管理端和服务端内部使用',
  `external_resource_location` VARCHAR(255) NULL COMMENT 'MCP 资源位置',
  `last_error_code` VARCHAR(64) NULL COMMENT '最近一次错误码',
  `last_error_message` VARCHAR(500) NULL COMMENT '最近一次错误说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `released_at` DATETIME(3) NULL COMMENT '释放完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_instances_instance_no` (`instance_no`),
  UNIQUE KEY `uk_instances_order_id` (`order_id`),
  UNIQUE KEY `uk_instances_external_vm` (`external_node`, `external_vmid`),
  KEY `idx_instances_user_status_created` (`user_id`, `status`, `created_at`),
  KEY `idx_instances_status_created` (`status`, `created_at`),
  CONSTRAINT `fk_instances_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_instances_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='云主机实例';

CREATE TABLE IF NOT EXISTS `instance_operations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '实例操作ID',
  `operation_no` VARCHAR(64) NOT NULL COMMENT '对外操作编号',
  `instance_id` BIGINT UNSIGNED NOT NULL COMMENT '实例ID',
  `order_id` BIGINT UNSIGNED NULL COMMENT '关联订单ID',
  `admin_id` BIGINT UNSIGNED NULL COMMENT '触发管理员ID',
  `user_id` BIGINT UNSIGNED NULL COMMENT '触发用户ID',
  `action` VARCHAR(32) NOT NULL COMMENT '操作：provision/start/stop/release/sync',
  `status` VARCHAR(32) NOT NULL DEFAULT 'running' COMMENT '操作状态：running/succeeded/failed',
  `external_operation_id` VARCHAR(128) NULL COMMENT 'MCP operation ID',
  `operation_location` VARCHAR(255) NULL COMMENT 'MCP Operation-Location',
  `resource_location` VARCHAR(255) NULL COMMENT 'MCP Location/resourceLocation',
  `error_code` VARCHAR(64) NULL COMMENT '失败错误码',
  `error_message` VARCHAR(500) NULL COMMENT '失败错误说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `completed_at` DATETIME(3) NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_instance_operations_operation_no` (`operation_no`),
  KEY `idx_instance_operations_instance_created` (`instance_id`, `created_at`),
  KEY `idx_instance_operations_external` (`external_operation_id`),
  KEY `idx_instance_operations_status` (`status`, `created_at`),
  CONSTRAINT `fk_instance_operations_instance` FOREIGN KEY (`instance_id`) REFERENCES `instances` (`id`),
  CONSTRAINT `fk_instance_operations_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_instance_operations_admin` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`),
  CONSTRAINT `fk_instance_operations_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实例异步操作';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.instances', '实例管理', 'menu', NULL, '/instances', 'Server', 80, 1, '菜单', '显示实例管理菜单和页面入口'),
  ('instance:*', '实例全权限', 'action', 'page.instances', NULL, NULL, 100, 0, '实例管理', '实例交付、操作、释放和同步全部能力'),
  ('instance:view', '查看实例', 'action', 'page.instances', NULL, NULL, 110, 0, '实例管理', '查看实例列表、详情、交付映射和 MCP 只读资源'),
  ('instance:provision', '交付实例', 'action', 'page.instances', NULL, NULL, 120, 0, '实例管理', '从订单触发实例交付'),
  ('instance:operate', '操作实例', 'action', 'page.instances', NULL, NULL, 130, 0, '实例管理', '启动、停止实例'),
  ('instance:release', '释放实例', 'action', 'page.instances', NULL, NULL, 140, 0, '实例管理', '释放实例并删除上游 VM'),
  ('instance:sync', '同步实例', 'action', 'page.instances', NULL, NULL, 150, 0, '实例管理', '同步 MCP operation 和 VM 状态')
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
