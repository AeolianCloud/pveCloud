-- users / wallets / wallet_logs
CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(120) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'user',
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  login_failed_count INT NOT NULL DEFAULT 0,
  locked_until DATETIME NULL,
  email_verified TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wallets (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL UNIQUE,
  balance DECIMAL(12,2) NOT NULL DEFAULT 0,
  frozen_balance DECIMAL(12,2) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_wallet_user_id (user_id)
);

CREATE TABLE IF NOT EXISTS wallet_logs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  type VARCHAR(20) NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  order_id BIGINT UNSIGNED NULL,
  remark VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_wallet_logs_user_id (user_id),
  KEY idx_wallet_logs_type (type),
  KEY idx_wallet_logs_order_id (order_id)
);

-- products / product_prices
CREATE TABLE IF NOT EXISTS products (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(120) NOT NULL,
  description TEXT NULL,
  region_id BIGINT UNSIGNED NOT NULL,
  cpu INT NOT NULL,
  memory_gb INT NOT NULL,
  disk_gb INT NOT NULL,
  bandwidth_mbps INT NOT NULL,
  disk_type VARCHAR(20) NOT NULL,
  os_options TEXT NULL,
  is_customizable TINYINT(1) NOT NULL DEFAULT 0,
  min_cpu INT NOT NULL DEFAULT 1,
  max_cpu INT NOT NULL DEFAULT 1,
  min_memory_gb INT NOT NULL DEFAULT 1,
  max_memory_gb INT NOT NULL DEFAULT 1,
  min_disk_gb INT NOT NULL DEFAULT 20,
  max_disk_gb INT NOT NULL DEFAULT 20,
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_products_region_id (region_id),
  KEY idx_products_status (status)
);

CREATE TABLE IF NOT EXISTS product_prices (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  product_id BIGINT UNSIGNED NOT NULL,
  billing_cycle VARCHAR(20) NOT NULL,
  unit_price DECIMAL(12,2) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_product_cycle (product_id, billing_cycle),
  KEY idx_product_prices_product_id (product_id)
);

-- orders / instances / instance_snapshots
CREATE TABLE IF NOT EXISTS orders (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  billing_cycle VARCHAR(20) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  config_snapshot JSON NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_orders_user_id (user_id),
  KEY idx_orders_product_id (product_id),
  KEY idx_orders_status (status)
);

CREATE TABLE IF NOT EXISTS instances (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  pve_instance_id VARCHAR(100) NULL,
  name VARCHAR(120) NULL,
  ip VARCHAR(64) NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  cpu INT NOT NULL,
  memory_gb INT NOT NULL,
  disk_gb INT NOT NULL,
  expire_at DATETIME NULL,
  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_instances_user_id (user_id),
  KEY idx_instances_order_id (order_id),
  KEY idx_instances_status (status),
  KEY idx_instances_expire_at (expire_at)
);

CREATE TABLE IF NOT EXISTS instance_snapshots (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  instance_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(120) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'available',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_instance_snapshot_name (instance_id, name),
  KEY idx_instance_snapshots_instance_id (instance_id)
);

-- tasks / tickets / ticket_replies
CREATE TABLE IF NOT EXISTS tasks (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NULL,
  instance_id BIGINT UNSIGNED NULL,
  type VARCHAR(50) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  pve_task_id VARCHAR(100) NULL,
  message VARCHAR(255) NULL,
  progress INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_tasks_user_id (user_id),
  KEY idx_tasks_order_id (order_id),
  KEY idx_tasks_instance_id (instance_id),
  KEY idx_tasks_status (status),
  KEY idx_tasks_pve_task_id (pve_task_id)
);

CREATE TABLE IF NOT EXISTS tickets (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  instance_id BIGINT UNSIGNED NULL,
  title VARCHAR(120) NOT NULL,
  content TEXT NOT NULL,
  priority VARCHAR(20) NOT NULL DEFAULT 'medium',
  status VARCHAR(20) NOT NULL DEFAULT 'open',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_tickets_user_id (user_id),
  KEY idx_tickets_instance_id (instance_id),
  KEY idx_tickets_status (status)
);

CREATE TABLE IF NOT EXISTS ticket_replies (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  ticket_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  content TEXT NOT NULL,
  is_admin TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_ticket_replies_ticket_id (ticket_id),
  KEY idx_ticket_replies_user_id (user_id)
);

-- regions / system_configs
CREATE TABLE IF NOT EXISTS regions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(32) NOT NULL UNIQUE,
  name VARCHAR(80) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_configs (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  config_key VARCHAR(80) NOT NULL UNIQUE,
  config_val TEXT NOT NULL,
  remark VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
