import { defineStore } from 'pinia'

import type { AdminMenuItem, AdminPermissionSnapshot } from '../../api/auth'
import type { SidebarMenuItem } from '../../utils/permission'
import { buildSidebarMenusFromSnapshot, checkPermission } from '../../utils/permission'

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
    sidebarMenus: (state): SidebarMenuItem[] => buildSidebarMenusFromSnapshot(state.menuSnapshot),
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
