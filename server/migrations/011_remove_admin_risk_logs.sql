-- Remove the retired admin high-risk log table and permissions from upgraded environments.
-- New environments no longer create admin_risk_logs in the baseline schema.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

DELETE `arp`
FROM `admin_role_permissions` AS `arp`
JOIN `admin_permissions` AS `ap` ON `ap`.`id` = `arp`.`permission_id`
WHERE `ap`.`code` IN (
  'audit:sensitive_view',
  'risk-log:*',
  'risk-log:view'
);

DELETE FROM `admin_permissions`
WHERE `code` IN (
  'audit:sensitive_view',
  'risk-log:*',
  'risk-log:view'
);

DROP TABLE IF EXISTS `admin_risk_logs`;
