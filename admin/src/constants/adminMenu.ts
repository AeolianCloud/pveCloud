import type { AdminMenuItem } from '../types/dashboard'

export const fallbackAdminMenus: AdminMenuItem[] = [
  {
    key: 'dashboard',
    title: '首页',
    path: '/dashboard',
    icon: 'layout-dashboard',
    permission_code: 'dashboard:view',
  },
  {
    key: 'users',
    title: '用户',
    path: '/users',
    icon: 'users',
    permission_code: 'user:view',
  },
  {
    key: 'products',
    title: '产品配置',
    path: '/products',
    icon: 'package',
    permission_code: 'product:update',
  },
  {
    key: 'orders',
    title: '订单',
    path: '/orders',
    icon: 'receipt-text',
    permission_code: 'order:view',
  },
  {
    key: 'payments',
    title: '支付与钱包',
    path: '/payments',
    icon: 'wallet-cards',
    permission_code: 'payment:view',
  },
  {
    key: 'instances',
    title: '实例',
    path: '/instances',
    icon: 'server',
    permission_code: 'instance:view',
  },
  {
    key: 'tickets',
    title: '工单',
    path: '/tickets',
    icon: 'message-square',
    permission_code: 'ticket:reply',
  },
  {
    key: 'admins',
    title: '管理员',
    path: '/admins',
    icon: 'shield-check',
    permission_code: 'admin:manage',
  },
  {
    key: 'system',
    title: '系统设置',
    path: '/system',
    icon: 'settings',
    permission_code: 'system:update',
  },
  {
    key: 'audit',
    title: '审计日志',
    path: '/audit',
    icon: 'clipboard-list',
    permission_code: 'audit:view',
  },
]

export function visibleAdminMenus(permissionCodes: string[], menus = fallbackAdminMenus) {
  const permissionSet = new Set(permissionCodes)
  return menus.filter((menu) => !menu.permission_code || permissionSet.has(menu.permission_code))
}
