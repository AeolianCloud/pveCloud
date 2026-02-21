-- =============================================================
-- pveCloud 管理后台种子数据
-- 执行前请确保已执行 init.sql
-- 执行方式：mysql -u pvecloud -p pvecloud < seed.sql
-- =============================================================

SET NAMES utf8mb4;

-- -------------------------------------------------------------
-- 1. 初始权限数据（按模块分组）
-- -------------------------------------------------------------
INSERT INTO `admin_permissions` (`name`, `label`, `group`) VALUES
  -- 管理员账号模块
  ('admin:list',   '查看管理员列表', 'admin'),
  ('admin:create', '创建管理员',     'admin'),
  ('admin:update', '编辑管理员',     'admin'),
  ('admin:delete', '删除管理员',     'admin'),
  ('admin:status', '启用/禁用管理员','admin'),
  -- 角色管理模块
  ('role:list',    '查看角色列表',   'role'),
  ('role:create',  '创建角色',       'role'),
  ('role:update',  '编辑角色',       'role'),
  ('role:delete',  '删除角色',       'role'),
  ('role:assign',  '分配权限',       'role'),
  -- 日志模块
  ('log:list',     '查看登录日志',   'log'),
  -- 操作日志模块
  ('op:list',      '查看操作日志',   'op')
ON DUPLICATE KEY UPDATE `label` = VALUES(`label`);


-- -------------------------------------------------------------
-- 2. 默认角色
-- -------------------------------------------------------------
INSERT INTO `admin_roles` (`name`, `label`, `description`, `sort`) VALUES
  ('super_admin', '超级管理员', '拥有全部权限，不受权限控制', 0),
  ('admin',       '普通管理员', '拥有大部分操作权限',         10)
ON DUPLICATE KEY UPDATE `label` = VALUES(`label`), `description` = VALUES(`description`);


-- -------------------------------------------------------------
-- 3. 角色权限关联：普通管理员拥有查看类权限
-- -------------------------------------------------------------
INSERT IGNORE INTO `admin_role_permissions` (`admin_role_id`, `admin_permission_id`)
SELECT r.id, p.id
FROM `admin_roles` r
JOIN `admin_permissions` p ON p.`name` IN ('admin:list', 'role:list', 'log:list')
WHERE r.`name` = 'admin';

-- 超级管理员拥有全部权限
INSERT IGNORE INTO `admin_role_permissions` (`admin_role_id`, `admin_permission_id`)
SELECT r.id, p.id
FROM `admin_roles` r
JOIN `admin_permissions` p ON 1=1
WHERE r.`name` = 'super_admin';


-- -------------------------------------------------------------
-- 4. 初始超管账号 admin / Admin@123
--    密码为 bcrypt hash，可用 backend/tools/genpwd 重新生成
-- -------------------------------------------------------------
INSERT INTO `admin_users` (`username`, `password`, `nickname`, `status`) VALUES
  ('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '超级管理员', 1)
ON DUPLICATE KEY UPDATE `nickname` = VALUES(`nickname`);

-- 给 admin 账号分配超级管理员角色
INSERT IGNORE INTO `admin_user_roles` (`admin_user_id`, `admin_role_id`)
SELECT u.id, r.id
FROM `admin_users` u
JOIN `admin_roles` r ON r.`name` = 'super_admin'
WHERE u.`username` = 'admin';
