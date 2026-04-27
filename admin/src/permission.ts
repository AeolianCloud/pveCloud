import type { RouteLocationNormalized } from 'vue-router'

import { ADMIN_ROUTE_NAME, ADMIN_ROUTE_PATH, normalizeAdminRedirect } from './router/constants'
import { router } from './router'
import { useAuthStore } from './store/modules/auth'
import { usePermissionStore } from './store/modules/permission'
import { normalizePermissions } from './utils/permission'

router.beforeEach(async (to) => {
  const authStore = useAuthStore()
  const permissionStore = usePermissionStore()
  const requiresAuth = to.matched.some((record) => Boolean(record.meta.requiresAuth))

  if (authStore.hasToken && !authStore.restored) {
    await authStore.restore()
  }

  if (to.meta.guestOnly) {
    if (authStore.isLoggedIn) {
      return resolveRedirect(ADMIN_ROUTE_PATH.dashboard)
    }
    return true
  }

  if (requiresAuth && !authStore.isLoggedIn) {
    return {
      name: ADMIN_ROUTE_NAME.login,
      query: {
        redirect: normalizeAdminRedirect(to.fullPath),
      },
    }
  }

  const requiredPermissions = collectRoutePermissions(to)
  if (requiresAuth && requiredPermissions.length > 0 && !permissionStore.hasAllPermissions(requiredPermissions)) {
    return { name: ADMIN_ROUTE_NAME.forbidden }
  }

  return true
})

function collectRoutePermissions(route: RouteLocationNormalized) {
  for (let index = route.matched.length - 1; index >= 0; index -= 1) {
    const permissions = normalizePermissions(route.matched[index].meta.permission)
    if (permissions.length > 0) {
      return permissions
    }
  }
  return []
}

function resolveRedirect(path: string) {
  return {
    path,
    replace: true,
  }
}
