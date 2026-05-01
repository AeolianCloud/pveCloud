-- pveCloud file attachments table and permission seeds.
-- Target: MariaDB 11.4.9 / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `file_attachments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '附件ID',
  `original_name` VARCHAR(255) NOT NULL COMMENT '原始文件名',
  `stored_name` VARCHAR(128) NOT NULL COMMENT '存储文件名（UUID）',
  `mime_type` VARCHAR(128) NOT NULL COMMENT 'MIME 类型',
  `extension` VARCHAR(32) NOT NULL COMMENT '文件扩展名',
  `size` BIGINT UNSIGNED NOT NULL COMMENT '文件大小（字节）',
  `storage_path` VARCHAR(500) NOT NULL COMMENT '存储路径（相对）',
  `storage_driver` VARCHAR(32) NOT NULL DEFAULT 'local' COMMENT '存储驱动',
  `checksum` VARCHAR(64) NOT NULL COMMENT '文件校验和（SHA-256）',
  `uploader_id` BIGINT UNSIGNED NOT NULL COMMENT '上传者管理员ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '状态：active/deleted',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_attachments_stored_name` (`stored_name`),
  KEY `idx_file_attachments_uploader_id` (`uploader_id`),
  KEY `idx_file_attachments_status` (`status`),
  KEY `idx_file_attachments_created_at` (`created_at`),
  CONSTRAINT `fk_file_attachments_uploader` FOREIGN KEY (`uploader_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件附件';

-- 文件管理菜单权限
INSERT IGNORE INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`)
VALUES ('page.file-management', '文件管理', 'menu', NULL, '/files', 'FolderOpened', 60, 1, 'system', '文件上传与附件管理');

-- 文件管理操作权限
INSERT IGNORE INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`)
VALUES
  ('file:upload', '上传文件', 'action', 'page.file-management', NULL, NULL, 1, 0, 'system', '允许上传文件'),
  ('file:delete', '删除文件', 'action', 'page.file-management', NULL, NULL, 2, 0, 'system', '允许删除文件'),
  ('file:*', '文件全部权限', 'action', 'page.file-management', NULL, NULL, 3, 0, 'system', '文件管理全部权限');

-- 超级管理员角色拥有文件管理全部权限
INSERT IGNORE INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT r.id, p.id
FROM `admin_roles` r, `admin_permissions` p
WHERE r.code = 'super_admin' AND p.code IN ('page.file-management', 'file:upload', 'file:delete', 'file:*');
