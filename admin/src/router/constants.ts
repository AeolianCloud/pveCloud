export const ADMIN_ROUTE_PATH = {
  root: '/',
  login: '/login',
  dashboard: '/dashboard',
  systemSettings: '/system',
  systemSettingsConfig: '/system/settings',
  adminUsers: '/system/admin-users',
  auditLogs: '/system/audit-logs',
  logs: '/logs',
  adminOperationLogs: '/logs/admin-operations',
  adminSecurityLogs: '/logs/admin-security',
  userSecurityLogs: '/logs/user-security',
  userBusinessLogs: '/logs/user-business',
  frontendErrorLogs: '/logs/frontend-errors',
  backendRuntimeLogs: '/logs/backend-runtime',
  files: '/files',
  webUsers: '/web/users',
  realNames: '/web/real-names',
  products: '/products',
  orders: '/orders',
  tickets: '/tickets',
  forbidden: '/403',
} as const

export const ADMIN_ROUTE_NAME = {
  login: 'login',
  dashboard: 'dashboard',
  systemSettings: 'system-settings',
  systemSettingsConfig: 'system-settings-config',
  adminUsers: 'admin-users',
  auditLogs: 'audit-logs',
  logs: 'logs',
  adminOperationLogs: 'logs-admin-operations',
  adminSecurityLogs: 'logs-admin-security',
  userSecurityLogs: 'logs-user-security',
  userBusinessLogs: 'logs-user-business',
  frontendErrorLogs: 'logs-frontend-errors',
  backendRuntimeLogs: 'logs-backend-runtime',
  files: 'files',
  webUsers: 'web-users',
  realNames: 'real-names',
  products: 'products',
  orders: 'orders',
  tickets: 'tickets',
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
