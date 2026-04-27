import type { RouteRecordRaw } from 'vue-router'

import type { RoutePermission } from '../router/types'

export interface SidebarMenuItem {
  key: string
  title: string
  path: string
  icon: string | null
  affix: boolean
  children?: SidebarMenuItem[]
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

  if (options.mode === 'all') {
    return normalized.every((permission) => hasPermissionCode(permissionCodes, permission))
  }

  return normalized.some((permission) => hasPermissionCode(permissionCodes, permission))
}

export function hasPermissionCode(permissionCodes: string[], required: string) {
  const value = required.trim()
  if (!value) {
    return true
  }

  const permissionSet = new Set(permissionCodes.map((code) => code.trim()).filter(Boolean))
  if (permissionSet.has(value)) {
    return true
  }

  const separatorIndex = value.indexOf(':')
  if (separatorIndex <= 0 || separatorIndex === value.length - 1) {
    return false
  }

  const module = value.slice(0, separatorIndex)
  const action = value.slice(separatorIndex + 1)
  if (!module || !action || action === '*') {
    return false
  }

  return permissionSet.has(`${module}:*`)
}

function resolvePermissionMode(route: RouteRecordRaw): 'all' | 'any' {
  return route.meta?.permissionMode === 'any' ? 'any' : 'all'
}

function filterChildRoutes(routes: RouteRecordRaw[], permissionCodes: string[]) {
  return routes.filter((route) => checkPermission(permissionCodes, route.meta?.permission, { mode: resolvePermissionMode(route) }))
}

export function filterViewRoutes(routes: RouteRecordRaw[], permissionCodes: string[]) {
  return routes
    .filter((route) => {
      if (route.children?.length) {
        return route.children.some((child) =>
          checkPermission(permissionCodes, child.meta?.permission, { mode: resolvePermissionMode(child) }),
        )
      }
      return checkPermission(permissionCodes, route.meta?.permission, { mode: resolvePermissionMode(route) })
    })
    .map((route) => {
      if (route.children?.length) {
        return { ...route, children: filterChildRoutes(route.children, permissionCodes) }
      }
      return route
    })
}

function buildChildMenus(routes: RouteRecordRaw[], parentPath: string): SidebarMenuItem[] {
  return routes
    .filter((route) => !route.meta?.hidden)
    .map((route) => ({
      key: String(route.name || route.path),
      title: String(route.meta?.title || route.name || route.path),
      path: route.path.startsWith('/') ? route.path : `${parentPath}/${route.path}`,
      icon: typeof route.meta?.icon === 'string' ? route.meta.icon : null,
      affix: Boolean(route.meta?.affix),
    }))
}

export function buildSidebarMenus(routes: RouteRecordRaw[]): SidebarMenuItem[] {
  return routes
    .filter((route) => !route.meta?.hidden)
    .map((route) => {
      const base: SidebarMenuItem = {
        key: String(route.name || route.path),
        title: String(route.meta?.title || route.name || route.path),
        path: route.path,
        icon: typeof route.meta?.icon === 'string' ? route.meta.icon : null,
        affix: Boolean(route.meta?.affix),
      }

      if (route.children?.length) {
        const children = buildChildMenus(route.children, route.path)
        if (route.meta?.alwaysShow || children.length > 1) {
          return { ...base, children }
        }
        if (children.length === 1) {
          return children[0]
        }
      }

      return base
    })
}
