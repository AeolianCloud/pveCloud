export const ADMIN_ROUTE_PATH = {
  root: '/',
  login: '/login',
  dashboard: '/dashboard',
  systemSettings: '/system',
  systemSettingsConfig: '/system/settings',
  adminUsers: '/system/admin-users',
  forbidden: '/403',
} as const

export const ADMIN_ROUTE_NAME = {
  login: 'login',
  dashboard: 'dashboard',
  systemSettings: 'system-settings',
  systemSettingsConfig: 'system-settings-config',
  adminUsers: 'admin-users',
  forbidden: 'forbidden',
} as const

export function normalizeAdminRedirect(target: unknown, fallback = ADMIN_ROUTE_PATH.dashboard) {
  if (typeof target !== 'string') {
    return fallback
  }

  const value = target.trim()
  if (!value.startsWith('/') || value.startsWith('//') || value.includes('://')) {
    return fallback
  }

  return value
}
