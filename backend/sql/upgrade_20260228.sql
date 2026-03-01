-- =============================================================
-- pveCloud 管理后台数据库升级脚本（补齐 admin_menus 菜单表）
--
-- 适用场景：
-- - 已在历史版本执行过 init.sql（未包含 admin_menus）
-- - 需要新增“动态菜单下发 + 菜单管理”能力
--
-- 执行方式：
--   mysql -u用户名 -p 数据库名 < upgrade_20260228.sql
-- =============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- -------------------------------------------------------------
-- 1) 新增 admin_menus 表（若不存在）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `admin_menus` (
  `id`               BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT      COMMENT '主键',
  `parent_id`        BIGINT UNSIGNED  NOT NULL DEFAULT 0           COMMENT '父菜单 ID，0 表示顶级菜单',
  `title`            VARCHAR(64)      NOT NULL                     COMMENT '菜单标题',
  `path`             VARCHAR(128)              DEFAULT NULL        COMMENT '前端路由路径（目录节点可为 NULL）',
  `permission`       VARCHAR(64)               DEFAULT NULL        COMMENT '可见权限标识（为空表示无需权限）',
  `super_admin_only` TINYINT(1)       NOT NULL DEFAULT 0           COMMENT '是否仅超级管理员可见：1 是 0 否',
  `icon`             VARCHAR(64)               DEFAULT NULL        COMMENT '菜单图标标识（前端按约定映射）',
  `sort`             INT              NOT NULL DEFAULT 0           COMMENT '排序权重，值越小越靠前',
  `visible`          TINYINT(1)       NOT NULL DEFAULT 1           COMMENT '是否显示：1 显示 0 隐藏',
  `created_at`       DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`       DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`       DATETIME                  DEFAULT NULL        COMMENT '软删除时间，NULL 表示未删除',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id`    (`parent_id`),
  KEY `idx_deleted_at`   (`deleted_at`),
  UNIQUE KEY `uk_path`   (`path`),
  UNIQUE KEY `uk_parent_title` (`parent_id`, `title`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理后台菜单表（树形结构 + 动态下发）';

SET FOREIGN_KEY_CHECKS = 1;
