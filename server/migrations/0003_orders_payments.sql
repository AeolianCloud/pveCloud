CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  order_no VARCHAR(32) NOT NULL COMMENT '订单编号，业务侧唯一编号',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID，对应users.id',
  sku_id BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID，对应product_skus.id',
  region_id BIGINT UNSIGNED NOT NULL COMMENT '区域ID，对应regions.id',
  reservation_id BIGINT UNSIGNED NULL COMMENT '预占ID，对应resource_reservations.id',
  order_status VARCHAR(32) NOT NULL COMMENT '订单状态：pending_payment-待支付，paid-已支付，provisioning-开通中，active-已生效，failed-失败，closed-已关闭',
  cycle_unit VARCHAR(32) NOT NULL COMMENT '购买周期：month-月，quarter-季，year-年',
  original_amount BIGINT NOT NULL COMMENT '订单原始金额，单位分',
  discount_amount BIGINT NOT NULL DEFAULT 0 COMMENT '订单优惠金额，单位分',
  payable_amount BIGINT NOT NULL COMMENT '订单应付金额，单位分',
  paid_at DATETIME(3) NULL COMMENT '支付成功时间',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  KEY idx_user_id (user_id),
  KEY idx_order_status (order_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';

CREATE TABLE order_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID，对应orders.id',
  sku_id BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID，对应product_skus.id',
  sku_name_snapshot VARCHAR(128) NOT NULL COMMENT 'SKU名称快照',
  cpu_cores_snapshot INT NOT NULL COMMENT 'CPU核数快照，单位核',
  memory_mb_snapshot INT NOT NULL COMMENT '内存容量快照，单位MB',
  disk_gb_snapshot INT NOT NULL COMMENT '磁盘容量快照，单位GB',
  bandwidth_mbps_snapshot INT NOT NULL COMMENT '带宽峰值快照，单位Mbps',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';

CREATE TABLE payment_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  payment_order_no VARCHAR(32) NOT NULL COMMENT '支付单编号，业务侧唯一编号',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID，对应orders.id',
  pay_status VARCHAR(32) NOT NULL COMMENT '支付状态：pending-待支付，success-支付成功，failed-支付失败，refunded-已退款',
  payable_amount BIGINT NOT NULL COMMENT '应付金额，单位分',
  paid_at DATETIME(3) NULL COMMENT '支付成功时间',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_payment_order_no (payment_order_no),
  KEY idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付单主表';

CREATE TABLE payment_callback_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  payment_order_no VARCHAR(32) NOT NULL COMMENT '支付单编号，业务侧唯一编号',
  callback_status VARCHAR(32) NOT NULL COMMENT '回调处理状态：received-已接收，success-处理成功，ignored-重复忽略，failed-处理失败',
  raw_payload JSON NOT NULL COMMENT '支付回调原始报文',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_payment_order_no (payment_order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付回调日志表';

CREATE TABLE billing_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID，对应orders.id',
  billing_type VARCHAR(32) NOT NULL COMMENT '计费类型：create-新购，renew-续费，change-变更',
  cycle_unit VARCHAR(32) NOT NULL COMMENT '计费周期：month-月，quarter-季，year-年',
  original_amount BIGINT NOT NULL COMMENT '原始金额，单位分',
  discount_amount BIGINT NOT NULL DEFAULT 0 COMMENT '优惠金额，单位分',
  payable_amount BIGINT NOT NULL COMMENT '应付金额，单位分',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计费记录表';
