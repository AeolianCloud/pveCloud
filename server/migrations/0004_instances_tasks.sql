CREATE TABLE instances (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  instance_no VARCHAR(32) NOT NULL COMMENT '实例编号，业务侧唯一编号',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID，对应users.id',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID，对应orders.id',
  node_id BIGINT UNSIGNED NOT NULL COMMENT '节点ID，对应resource_nodes.id',
  instance_status VARCHAR(32) NOT NULL COMMENT '实例状态：creating-创建中，running-运行中，stopped-已关机，reinstalling-重装中，starting-开机中，stopping-关机中，failed-失败，expired-已到期',
  instance_ref VARCHAR(128) NULL COMMENT '底层实例标识',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_instance_no (instance_no),
  KEY idx_user_id (user_id),
  KEY idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务实例主表';

CREATE TABLE instance_services (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  instance_id BIGINT UNSIGNED NOT NULL COMMENT '实例ID，对应instances.id',
  current_period_start_at DATETIME(3) NOT NULL COMMENT '当前服务周期开始时间',
  current_period_end_at DATETIME(3) NOT NULL COMMENT '当前服务周期结束时间',
  billing_status VARCHAR(32) NOT NULL COMMENT '服务计费状态：active-生效中，expired-已过期, suspended-已暂停',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_instance_id (instance_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='实例服务周期事实表';

CREATE TABLE instance_actions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  action_no VARCHAR(32) NOT NULL COMMENT '操作编号，业务侧唯一编号',
  instance_id BIGINT UNSIGNED NOT NULL COMMENT '实例ID，对应instances.id',
  action_type VARCHAR(32) NOT NULL COMMENT '操作类型：start-开机，stop-关机，reboot-重启，reinstall-重装',
  action_status VARCHAR(32) NOT NULL COMMENT '操作状态：pending-待执行，processing-执行中，success-成功，failed-失败',
  operator_type VARCHAR(32) NOT NULL COMMENT '操作人类型：user-用户，admin-管理员，system-系统',
  operator_id BIGINT UNSIGNED NULL COMMENT '操作人ID',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_action_no (action_no),
  KEY idx_instance_id (instance_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='实例操作记录表';

CREATE TABLE async_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  task_no VARCHAR(32) NOT NULL COMMENT '任务编号，业务侧唯一编号',
  task_type VARCHAR(32) NOT NULL COMMENT '任务类型：create_instance-开通实例，start_instance-开机，stop_instance-关机，reboot_instance-重启，reinstall_instance-重装',
  business_type VARCHAR(32) NOT NULL COMMENT '业务类型：order-订单，instance-实例',
  business_id BIGINT UNSIGNED NOT NULL COMMENT '业务ID',
  status VARCHAR(32) NOT NULL COMMENT '任务状态：pending-待执行，processing-执行中，success-成功，failed-失败，retrying-重试中',
  payload JSON NULL COMMENT '任务载荷',
  next_run_at DATETIME(3) NOT NULL COMMENT '下次可执行时间',
  retry_count INT NOT NULL DEFAULT 0 COMMENT '当前重试次数',
  max_retry_count INT NOT NULL DEFAULT 5 COMMENT '最大重试次数',
  locked_by VARCHAR(64) NULL COMMENT '任务抢占者标识',
  locked_at DATETIME(3) NULL COMMENT '任务抢占时间',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_task_business (task_type, business_type, business_id),
  UNIQUE KEY uk_task_no (task_no),
  KEY idx_status_next_run (status, next_run_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异步任务主表';

CREATE TABLE async_task_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  task_id BIGINT UNSIGNED NOT NULL COMMENT '任务ID，对应async_tasks.id',
  log_level VARCHAR(16) NOT NULL COMMENT '日志级别：info-信息，warn-警告，error-错误',
  message VARCHAR(255) NOT NULL COMMENT '日志内容',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_task_id (task_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异步任务日志表';
