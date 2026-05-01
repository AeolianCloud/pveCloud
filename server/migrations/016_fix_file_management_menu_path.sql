-- Fix existing file management menu path for deployed databases.

SET NAMES utf8mb4;

USE `pvecloud`;

UPDATE `admin_permissions`
SET `path` = '/files',
    `icon` = 'FolderOpened',
    `updated_at` = CURRENT_TIMESTAMP(3)
WHERE `code` = 'page.file-management';
