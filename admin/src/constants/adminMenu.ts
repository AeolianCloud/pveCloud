import type { AdminMenuItem } from '../types/dashboard'

export const fallbackAdminMenus: AdminMenuItem[] = [
  {
    key: 'dashboard',
    title: '控制台',
    path: '/dashboard',
    icon: 'layout-dashboard',
    permission_code: 'dashboard:view',
  },
]

export function visibleAdminMenus(permissionCodes: string[], menus = fallbackAdminMenus) {
  const permissionSet = new Set(permissionCodes)
  return menus.filter((menu) => !menu.permission_code || permissionSet.has(menu.permission_code))
}
