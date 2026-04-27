import { defineStore } from 'pinia'

import type { AdminMenuItem, AdminPermissionSnapshot } from '../../api/auth'
import { viewRoutes } from '../../router/view-routes'
import { buildSidebarMenus, checkPermission, filterViewRoutes } from '../../utils/permission'

interface PermissionState {
  roleIds: number[]
  permissionCodes: string[]
  menuSnapshot: AdminMenuItem[]
}

export const usePermissionStore = defineStore('admin-permission', {
  state: (): PermissionState => ({
    roleIds: [],
    permissionCodes: [],
    menuSnapshot: [],
  }),
  getters: {
    sidebarMenus: (state) => buildSidebarMenus(filterViewRoutes(viewRoutes, state.permissionCodes)),
    hasPermission: (state) => (required?: string | string[]) =>
      checkPermission(state.permissionCodes, required),
    hasAllPermissions: (state) => (required?: string | string[]) =>
      checkPermission(state.permissionCodes, required, { mode: 'all' }),
  },
  actions: {
    setSnapshot(payload: AdminPermissionSnapshot) {
      this.roleIds = [...payload.role_ids]
      this.permissionCodes = [...payload.permission_codes]
      this.menuSnapshot = [...payload.menus]
    },
    reset() {
      this.roleIds = []
      this.permissionCodes = []
      this.menuSnapshot = []
    },
  },
})
