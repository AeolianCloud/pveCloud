CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  user_no VARCHAR(32) NOT NULL COMMENT '用户编号，业务侧唯一编号',
  email VARCHAR(128) NULL COMMENT '邮箱地址',
  phone VARCHAR(32) NOT NULL COMMENT '手机号',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
  status VARCHAR(32) NOT NULL COMMENT '用户状态：active-正常，disabled-禁用',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_no (user_no),
  UNIQUE KEY uk_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='前台用户主表';

CREATE TABLE admins (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  admin_no VARCHAR(32) NOT NULL COMMENT '管理员编号，业务侧唯一编号',
  username VARCHAR(64) NOT NULL COMMENT '登录用户名',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
  status VARCHAR(32) NOT NULL COMMENT '管理员状态：active-正常，disabled-禁用',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_no (admin_no),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='后台管理员主表';
