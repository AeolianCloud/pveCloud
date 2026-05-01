-- File attachment references.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `file_attachment_references` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '引用ID',
  `file_id` BIGINT UNSIGNED NOT NULL COMMENT '附件ID',
  `ref_type` VARCHAR(64) NOT NULL COMMENT '引用业务类型',
  `ref_id` VARCHAR(128) NOT NULL COMMENT '引用业务ID',
  `ref_name` VARCHAR(255) NULL COMMENT '引用名称',
  `ref_path` VARCHAR(255) NULL COMMENT '引用路径或位置',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_file_attachment_references_unique` (`file_id`, `ref_type`, `ref_id`),
  KEY `idx_file_attachment_references_file_id` (`file_id`),
  KEY `idx_file_attachment_references_ref_type` (`ref_type`),
  CONSTRAINT `fk_file_attachment_references_file` FOREIGN KEY (`file_id`) REFERENCES `file_attachments` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件引用关系';
