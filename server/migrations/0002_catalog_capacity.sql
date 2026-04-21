CREATE TABLE products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  product_no VARCHAR(32) NOT NULL COMMENT '商品编号，业务侧唯一编号',
  product_name VARCHAR(128) NOT NULL COMMENT '商品名称',
  product_type VARCHAR(32) NOT NULL COMMENT '商品类型：cloud_host-云服务器',
  status VARCHAR(32) NOT NULL COMMENT '商品状态：draft-草稿，active-上架，disabled-下架',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_product_no (product_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品主表';

CREATE TABLE product_skus (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  sku_no VARCHAR(32) NOT NULL COMMENT 'SKU编号，业务侧唯一编号',
  product_id BIGINT UNSIGNED NOT NULL COMMENT '商品ID，对应products.id',
  sku_name VARCHAR(128) NOT NULL COMMENT 'SKU名称',
  cpu_cores INT NOT NULL COMMENT 'CPU核数，单位核',
  memory_mb INT NOT NULL COMMENT '内存容量，单位MB',
  disk_gb INT NOT NULL COMMENT '磁盘容量，单位GB',
  bandwidth_mbps INT NOT NULL COMMENT '带宽峰值，单位Mbps',
  status VARCHAR(32) NOT NULL COMMENT 'SKU状态：draft-草稿，active-可售，disabled-停用',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_sku_no (sku_no),
  KEY idx_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品SKU主表';

CREATE TABLE regions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  region_no VARCHAR(32) NOT NULL COMMENT '区域编号，业务侧唯一编号',
  region_name VARCHAR(64) NOT NULL COMMENT '区域名称',
  status VARCHAR(32) NOT NULL COMMENT '区域状态：active-可用，disabled-停用',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_region_no (region_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='区域主表';

CREATE TABLE resource_nodes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  node_no VARCHAR(32) NOT NULL COMMENT '节点编号，业务侧唯一编号',
  region_id BIGINT UNSIGNED NOT NULL COMMENT '区域ID，对应regions.id',
  node_name VARCHAR(128) NOT NULL COMMENT '节点名称',
  total_instances INT NOT NULL COMMENT '实例总容量，单位台',
  used_instances INT NOT NULL DEFAULT 0 COMMENT '已分配实例数，单位台',
  reserved_instances INT NOT NULL DEFAULT 0 COMMENT '已预占实例数，单位台',
  status VARCHAR(32) NOT NULL COMMENT '节点状态：active-可售，disabled-停用',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_node_no (node_no),
  KEY idx_region_id (region_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资源节点主表';

CREATE TABLE sku_region_node_bindings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  sku_id BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID，对应product_skus.id',
  region_id BIGINT UNSIGNED NOT NULL COMMENT '区域ID，对应regions.id',
  node_id BIGINT UNSIGNED NOT NULL COMMENT '节点ID，对应resource_nodes.id',
  sale_status VARCHAR(32) NOT NULL COMMENT '可售状态：saleable-可售，unsaleable-不可售',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_sku_region_node (sku_id, region_id, node_id),
  KEY idx_region_node (region_id, node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SKU区域节点映射表';

CREATE TABLE resource_reservations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  reservation_no VARCHAR(32) NOT NULL COMMENT '预占编号，业务侧唯一编号',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID，对应users.id',
  sku_id BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID，对应product_skus.id',
  region_id BIGINT UNSIGNED NOT NULL COMMENT '区域ID，对应regions.id',
  node_id BIGINT UNSIGNED NOT NULL COMMENT '节点ID，对应resource_nodes.id',
  status VARCHAR(32) NOT NULL COMMENT '预占状态：reserved-已预占，consumed-已消费，released-已释放，expired-已过期',
  expires_at DATETIME(3) NOT NULL COMMENT '预占过期时间',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_reservation_no (reservation_no),
  KEY idx_status_expires_at (status, expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资源预占主表';
