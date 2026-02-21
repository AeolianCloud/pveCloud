-- =============================================================
-- pveCloud 管理后台数据库初始化脚本
-- 数据库：MariaDB / MySQL 5.7+
-- 字符集：utf8mb4
-- 执行方式：mysql -u用户名 -p 数据库名 < init.sql
-- 创建时间：2026-02-20
-- =============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- -------------------------------------------------------------
-- 1. 管理员账号表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `admin_users` (
  `id`            BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT        COMMENT '主键',
  `username`      VARCHAR(64)      NOT NULL                       COMMENT '登录用户名，全局唯一',
  `password`      VARCHAR(128)     NOT NULL                       COMMENT 'bcrypt 哈希密码，不可逆',
  `nickname`      VARCHAR(64)      NOT NULL DEFAULT ''            COMMENT '显示昵称',
  `avatar`        VARCHAR(255)     NOT NULL DEFAULT ''            COMMENT '头像图片 URL',
  `email`         VARCHAR(128)              DEFAULT NULL          COMMENT '邮箱地址，唯一，可为 NULL（未填写）',
  `status`        TINYINT(1)       NOT NULL DEFAULT 1             COMMENT '账号状态：1 启用  0 禁用',
  `last_login_at` DATETIME                  DEFAULT NULL          COMMENT '最后一次登录时间',
  `created_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`    DATETIME                  DEFAULT NULL          COMMENT '软删除时间，NULL 表示未删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email`    (`email`),
  KEY `idx_deleted_at`     (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理后台账号表';


-- -------------------------------------------------------------
-- 2. 角色表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `admin_roles` (
  `id`          BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT          COMMENT '主键',
  `name`        VARCHAR(64)      NOT NULL                         COMMENT '角色标识，全局唯一，如 super_admin / admin',
  `label`       VARCHAR(64)      NOT NULL                         COMMENT '角色显示名，如 超级管理员',
  `description` VARCHAR(255)     NOT NULL DEFAULT ''              COMMENT '角色描述',
  `sort`        INT              NOT NULL DEFAULT 0               COMMENT '排序权重，值越小排序越靠前',
  `created_at`  DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`  DATETIME                  DEFAULT NULL            COMMENT '软删除时间，NULL 表示未删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name`     (`name`),
  KEY `idx_deleted_at`     (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';


-- -------------------------------------------------------------
-- 3. 权限表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `admin_permissions` (
  `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT           COMMENT '主键',
  `name`       VARCHAR(64)      NOT NULL                          COMMENT '权限标识，全局唯一，格式：模块:操作，如 admin:create',
  `label`      VARCHAR(64)      NOT NULL                          COMMENT '权限显示名，如 创建管理员',
  `group`      VARCHAR(64)      NOT NULL                          COMMENT '所属分组，如 admin / order / user',
  `created_at` DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_group`      (`group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表，权限为系统操作的最小粒度';


-- -------------------------------------------------------------
-- 4. 用户 ↔ 角色 关联表（多对多）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `admin_user_roles` (
  `admin_user_id` BIGINT UNSIGNED NOT NULL                        COMMENT '管理员 ID，关联 admin_users.id',
  `admin_role_id` BIGINT UNSIGNED NOT NULL                        COMMENT '角色 ID，关联 admin_roles.id',
  PRIMARY KEY (`admin_user_id`, `admin_role_id`),
  KEY `idx_role_id` (`admin_role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员与角色多对多关联表';


-- -------------------------------------------------------------
-- 5. 角色 ↔ 权限 关联表（多对多）
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `admin_role_permissions` (
  `admin_role_id`       BIGINT UNSIGNED NOT NULL                  COMMENT '角色 ID，关联 admin_roles.id',
  `admin_permission_id` BIGINT UNSIGNED NOT NULL                  COMMENT '权限 ID，关联 admin_permissions.id',
  PRIMARY KEY (`admin_role_id`, `admin_permission_id`),
  KEY `idx_permission_id` (`admin_permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色与权限多对多关联表';


-- -------------------------------------------------------------
-- 6. 管理员登录日志表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `admin_login_logs` (
  `id`            BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT         COMMENT '主键',
  `admin_user_id` BIGINT UNSIGNED  NOT NULL DEFAULT 0             COMMENT '管理员 ID，0 表示用户名不存在时的登录尝试',
  `username`      VARCHAR(64)      NOT NULL                        COMMENT '冗余存储登录用户名，防止账号删除后丢失上下文',
  `ip`            VARCHAR(64)      NOT NULL DEFAULT ''             COMMENT '登录来源 IP',
  `user_agent`    VARCHAR(255)     NOT NULL DEFAULT ''             COMMENT '浏览器 User-Agent',
  `status`        TINYINT(1)       NOT NULL DEFAULT 0              COMMENT '登录结果：1 成功  0 失败',
  `remark`        VARCHAR(128)     NOT NULL DEFAULT ''             COMMENT '备注，失败时记录原因，如 密码错误 / 账号禁用',
  `created_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_user_id` (`admin_user_id`),
  KEY `idx_created_at`    (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员登录日志表，不做软删除，保留完整历史';


-- -------------------------------------------------------------
-- 7. 管理员操作日志表
-- -------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `admin_op_logs` (
  `id`            BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT         COMMENT '主键',
  `admin_user_id` BIGINT UNSIGNED  NOT NULL DEFAULT 0             COMMENT '操作人管理员 ID',
  `username`      VARCHAR(64)      NOT NULL DEFAULT ''             COMMENT '冗余操作人用户名',
  `module`        VARCHAR(64)      NOT NULL DEFAULT ''             COMMENT '操作模块，如 admin / role',
  `action`        VARCHAR(64)      NOT NULL DEFAULT ''             COMMENT '操作动作，如 create / update / delete',
  `target_id`     BIGINT UNSIGNED  NOT NULL DEFAULT 0             COMMENT '操作目标的主键 ID',
  `target_label`  VARCHAR(128)     NOT NULL DEFAULT ''             COMMENT '操作目标描述，冗余记录防丢失',
  `status`        TINYINT(1)       NOT NULL DEFAULT 1              COMMENT '执行结果：1 成功  0 失败',
  `ip`            VARCHAR(64)      NOT NULL DEFAULT ''             COMMENT '操作来源 IP',
  `created_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_user_id` (`admin_user_id`),
  KEY `idx_module`        (`module`),
  KEY `idx_created_at`    (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员操作日志表，记录所有增删改审计记录';


SET FOREIGN_KEY_CHECKS = 1;
