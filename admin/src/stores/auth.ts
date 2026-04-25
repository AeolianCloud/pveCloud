import { defineStore } from 'pinia'

import { loginAdmin } from '../api/auth'
import type { AdminLoginRequest, AdminSummary } from '../types/auth'
import type { AdminMenuItem, DashboardResponse } from '../types/dashboard'

const STORAGE_KEY = 'pvecloud_admin_auth'

interface AuthState {
  token: string
  admin: AdminSummary | null
  roleIds: number[]
  permissionCodes: string[]
  menuItems: AdminMenuItem[]
  sidebarCollapsed: boolean
}

function loadState(): AuthState {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return emptyState()
  }

  try {
    return { ...emptyState(), ...JSON.parse(raw) } as AuthState
  } catch {
    localStorage.removeItem(STORAGE_KEY)
    return emptyState()
  }
}

function emptyState(): AuthState {
  return {
    token: '',
    admin: null,
    roleIds: [],
    permissionCodes: [],
    menuItems: [],
    sidebarCollapsed: false,
  }
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => loadState(),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.admin),
    hasPermission: (state) => (permissionCode?: string) => {
      if (!permissionCode) {
        return true
      }
      return state.permissionCodes.includes(permissionCode)
    },
  },
  actions: {
    async login(payload: AdminLoginRequest) {
      const result = await loginAdmin(payload)
      this.token = result.access_token
      this.admin = result.admin
      this.roleIds = result.role_ids
      this.permissionCodes = result.permission_codes
      this.menuItems = []
      this.persist()
    },
    applyDashboard(payload: DashboardResponse) {
      this.admin = payload.admin
      this.roleIds = payload.role_ids
      this.permissionCodes = payload.permission_codes
      this.menuItems = payload.menus
      this.persist()
    },
    logout() {
      Object.assign(this, emptyState())
      localStorage.removeItem(STORAGE_KEY)
    },
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      this.persist()
    },
    persist() {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          token: this.token,
          admin: this.admin,
          roleIds: this.roleIds,
          permissionCodes: this.permissionCodes,
          menuItems: this.menuItems,
          sidebarCollapsed: this.sidebarCollapsed,
        }),
      )
    },
  },
})
