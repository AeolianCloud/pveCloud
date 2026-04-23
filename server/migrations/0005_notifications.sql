-- server/migrations/0005_notifications.sql

CREATE TABLE notifications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  title VARCHAR(255) NOT NULL COMMENT '通知标题',
  body TEXT NOT NULL COMMENT '通知正文',
  type VARCHAR(32) NOT NULL COMMENT '通知类型：system-系统通知，provision-开通通知，billing-账单通知',
  is_read TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_user_id_created (user_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户通知表';
