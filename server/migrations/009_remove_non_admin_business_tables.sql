-- Legacy upgrade cleanup for repositories that previously used the old
-- full business schema baseline. New environments should already start
-- from the admin-only baseline in 001_init.sql.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

DELETE `arp`
FROM `admin_role_permissions` AS `arp`
JOIN `admin_permissions` AS `ap` ON `ap`.`id` = `arp`.`permission_id`
WHERE `ap`.`code` IN (
  'user:view',
  'user:update',
  'product:create',
  'product:update',
  'order:view',
  'order:cancel',
  'payment:view',
  'payment:manual_credit',
  'instance:view',
  'instance:operate',
  'ticket:reply'
);

DELETE FROM `admin_permissions`
WHERE `code` IN (
  'user:view',
  'user:update',
  'product:create',
  'product:update',
  'order:view',
  'order:cancel',
  'payment:view',
  'payment:manual_credit',
  'instance:view',
  'instance:operate',
  'ticket:reply'
);

DELETE FROM `system_configs`
WHERE `config_key` IN (
  'order.expire_minutes',
  'payment.enabled_channels',
  'pve.default_region_id'
);

DROP TABLE IF EXISTS `wallet_transactions`;
DROP TABLE IF EXISTS `payment_notify_logs`;
DROP TABLE IF EXISTS `ticket_messages`;
DROP TABLE IF EXISTS `async_tasks`;
DROP TABLE IF EXISTS `instances`;
DROP TABLE IF EXISTS `payment_orders`;
DROP TABLE IF EXISTS `wallet_accounts`;
DROP TABLE IF EXISTS `tickets`;
DROP TABLE IF EXISTS `orders`;
DROP TABLE IF EXISTS `plan_prices`;
DROP TABLE IF EXISTS `region_images`;
DROP TABLE IF EXISTS `pve_nodes`;
DROP TABLE IF EXISTS `product_plans`;
DROP TABLE IF EXISTS `images`;
DROP TABLE IF EXISTS `regions`;
DROP TABLE IF EXISTS `products`;
DROP TABLE IF EXISTS `users`;
