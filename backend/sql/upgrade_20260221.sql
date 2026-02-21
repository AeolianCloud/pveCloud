-- =============================================================
-- pveCloud 管理后台数据库升级脚本（从旧版 init.sql 升级到 2026-02-21 版本）
--
-- 适用场景：
-- - 你在 2026-02-21 之前已执行过 init.sql
-- - 需要把 admin_users.email 从 “空字符串” 迁移为 NULL（与后端 Model 保持一致）
--
-- 执行方式：
--   mysql -u用户名 -p 数据库名 < upgrade_20260221.sql
-- =============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- -------------------------------------------------------------
-- 1) 修复 admin_users.email 字段：允许 NULL
--    说明：后端 Model 使用 *string（NULL 表示未填写），避免多个空字符串触发唯一索引冲突
-- -------------------------------------------------------------
ALTER TABLE `admin_users`
  MODIFY COLUMN `email` VARCHAR(128) NULL DEFAULT NULL COMMENT '邮箱地址，唯一，可为 NULL（未填写）';

-- 将历史空字符串统一转为 NULL（便于唯一约束与后端逻辑一致）
UPDATE `admin_users`
   SET `email` = NULL
 WHERE `email` = '';

SET FOREIGN_KEY_CHECKS = 1;
