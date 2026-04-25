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
    key: 'instances',
    title: '云服务器',
    path: '/instances',
    icon: 'cloud',
    permission_code: 'instance:view',
  },
  {
    key: 'products',
    title: '产品套餐',
    path: '/products',
    icon: 'package',
    permission_code: 'product:update',
  },
  {
    key: 'orders',
    title: '订单管理',
    path: '/orders',
    icon: 'receipt-text',
    permission_code: 'order:view',
  },
  {
    key: 'users',
    title: '客户管理',
    path: '/users',
    icon: 'users',
    permission_code: 'user:view',
  },
  {
    key: 'tickets',
    title: '工单服务',
    path: '/tickets',
    icon: 'message-square',
    permission_code: 'ticket:reply',
  },
  {
    key: 'audit',
    title: '资源监控',
    path: '/audit',
    icon: 'clipboard-list',
    permission_code: 'audit:view',
  },
  {
    key: 'payments',
    title: '财务中心',
    path: '/payments',
    icon: 'wallet-cards',
    permission_code: 'payment:view',
  },
  {
    key: 'admins',
    title: '营销活动',
    path: '/admins',
    icon: 'megaphone',
    permission_code: 'admin:manage',
  },
  {
    key: 'system',
    title: '系统设置',
    path: '/system',
    icon: 'settings',
    permission_code: 'system:update',
  },
]

export function visibleAdminMenus(permissionCodes: string[], menus = fallbackAdminMenus) {
  const permissionSet = new Set(permissionCodes)
  return menus.filter((menu) => !menu.permission_code || permissionSet.has(menu.permission_code))
}
