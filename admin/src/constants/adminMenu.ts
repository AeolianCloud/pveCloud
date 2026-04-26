import type { AdminMenuItem } from '../types/dashboard'

export const fallbackAdminMenus: AdminMenuItem[] = [
  {
    key: 'dashboard',
    title: '控制台',
    path: '/dashboard',
    icon: 'layout-dashboard',
    permission_code: 'dashboard:view',
  },
  {
    key: 'admin_users',
    title: '管理员账号',
    path: '/admin-users',
    icon: 'users',
    permission_code: 'admin:manage',
  },
  {
    key: 'admin_roles',
    title: '角色权限',
    path: '/admin-roles',
    icon: 'shield-check',
    permission_code: 'admin:manage',
  },
  {
    key: 'admin_sessions',
    title: '登录会话',
    path: '/admin-sessions',
    icon: 'monitor-check',
    permission_code: 'admin:manage',
  },
  {
    key: 'system_configs',
    title: '系统设置',
    path: '/system-configs',
    icon: 'settings',
    permission_code: 'system:update',
  },
  {
    key: 'audit_logs',
    title: '审计日志',
    path: '/audit-logs',
    icon: 'clipboard-list',
    permission_code: 'audit:view',
  },
  {
    key: 'risk_logs',
    title: '高危操作日志',
    path: '/risk-logs',
    icon: 'shield-alert',
    permission_code: 'audit:view',
  },
]

export function visibleAdminMenus(permissionCodes: string[], menus = fallbackAdminMenus) {
  const permissionSet = new Set(permissionCodes)
  return menus.filter((menu) => !menu.permission_code || permissionSet.has(menu.permission_code))
}
