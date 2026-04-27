import type { RouteRecordRaw } from 'vue-router'

import type { RoutePermission } from '../router/types'

export interface SidebarMenuItem {
  key: string
  title: string
  path: string
  icon: string | null
  affix: boolean
}

export interface PermissionCheckOptions {
  mode?: 'all' | 'any'
}

type PermissionValue = RoutePermission | undefined

export function normalizePermissions(value: PermissionValue): string[] {
  if (!value) {
    return []
  }

  return Array.isArray(value) ? value.filter(Boolean) : [value]
}

export function checkPermission(
  permissionCodes: string[],
  required: PermissionValue,
  options: PermissionCheckOptions = {},
) {
  const normalized = normalizePermissions(required)
  if (normalized.length === 0) {
    return true
  }

  const permissionSet = new Set(permissionCodes)
  if (options.mode === 'all') {
    return normalized.every((permission) => permissionSet.has(permission))
  }

  return normalized.some((permission) => permissionSet.has(permission))
}

export function filterViewRoutes(routes: RouteRecordRaw[], permissionCodes: string[]) {
  return routes.filter((route) => checkPermission(permissionCodes, route.meta?.permission, { mode: 'all' }))
}

export function buildSidebarMenus(routes: RouteRecordRaw[]): SidebarMenuItem[] {
  return routes
    .filter((route) => !route.meta?.hidden)
    .map((route) => ({
      key: String(route.name || route.path),
      title: String(route.meta?.title || route.name || route.path),
      path: route.path,
      icon: typeof route.meta?.icon === 'string' ? route.meta.icon : null,
      affix: Boolean(route.meta?.affix),
    }))
}
