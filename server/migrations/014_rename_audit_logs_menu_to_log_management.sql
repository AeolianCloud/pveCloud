-- Rename the audit logs menu label to log management while keeping the route and permission code stable.

UPDATE `admin_permissions`
SET `name` = '日志管理',
    `description` = '显示系统设置下的日志管理入口'
WHERE `code` = 'page.system-settings.audit-logs';
