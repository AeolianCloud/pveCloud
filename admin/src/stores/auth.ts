import { defineStore } from 'pinia'

import { getCurrentAdmin, loginAdmin, logoutAdmin, refreshAdminToken } from '../api/auth'
import type {
  AdminAuthStateResponse,
  AdminLoginRequest,
  AdminLoginResponse,
  AdminSessionSummary,
  AdminSummary,
} from '../types/auth'
import type { AdminMenuItem, DashboardResponse } from '../types/dashboard'

const STORAGE_KEY = 'pvecloud_admin_auth'

interface AuthState {
  token: string
  admin: AdminSummary | null
  roleIds: number[]
  permissionCodes: string[]
  menuItems: AdminMenuItem[]
  session: AdminSessionSummary | null
  sidebarCollapsed: boolean
  restored: boolean
}

let restorePromise: Promise<boolean> | null = null

function loadState(): AuthState {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return emptyState(true)
  }

  try {
    return { ...emptyState(false), ...JSON.parse(raw) } as AuthState
  } catch {
    localStorage.removeItem(STORAGE_KEY)
    return emptyState(true)
  }
}

function emptyState(restored = true): AuthState {
  return {
    token: '',
    admin: null,
    roleIds: [],
    permissionCodes: [],
    menuItems: [],
    session: null,
    sidebarCollapsed: false,
    restored,
  }
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => loadState(),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.admin),
    hasLocalToken: (state) => Boolean(state.token),
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
      this.applyLogin(result)
    },
    async restore() {
      if (!this.token) {
        this.restored = true
        return false
      }
      if (this.restored && this.isLoggedIn) {
        return true
      }
      if (restorePromise) {
        return restorePromise
      }

      restorePromise = getCurrentAdmin()
        .then((result) => {
          this.applyAuthState(result)
          return true
        })
        .catch(() => {
          this.logout()
          return false
        })
        .finally(() => {
          restorePromise = null
          this.restored = true
        })

      return restorePromise
    },
    async logoutRemote() {
      try {
        if (this.token) {
          await logoutAdmin()
        }
      } finally {
        this.logout()
      }
    },
    async refresh() {
      const result = await refreshAdminToken()
      this.applyLogin(result)
    },
    applyLogin(payload: AdminLoginResponse) {
      this.token = payload.access_token
      this.admin = payload.admin
      this.roleIds = payload.role_ids
      this.permissionCodes = payload.permission_codes
      this.session = payload.session
      this.menuItems = []
      this.restored = true
      this.persist()
    },
    applyAuthState(payload: AdminAuthStateResponse) {
      this.admin = payload.admin
      this.roleIds = payload.role_ids
      this.permissionCodes = payload.permission_codes
      this.menuItems = payload.menus
      this.session = payload.session
      this.restored = true
      this.persist()
    },
    applyDashboard(payload: DashboardResponse) {
      this.admin = payload.admin
      this.roleIds = payload.role_ids
      this.permissionCodes = payload.permission_codes
      this.menuItems = payload.menus
      this.session = payload.session
      this.restored = true
      this.persist()
    },
    logout() {
      const sidebarCollapsed = this.sidebarCollapsed
      Object.assign(this, emptyState(true), { sidebarCollapsed })
      localStorage.removeItem(STORAGE_KEY)
    },
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      this.persist()
    },
    persist() {
      if (!this.token) {
        localStorage.removeItem(STORAGE_KEY)
        return
      }
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          token: this.token,
          admin: this.admin,
          roleIds: this.roleIds,
          permissionCodes: this.permissionCodes,
          menuItems: this.menuItems,
          session: this.session,
          sidebarCollapsed: this.sidebarCollapsed,
        }),
      )
    },
  },
})
