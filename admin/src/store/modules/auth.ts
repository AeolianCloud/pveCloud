import { defineStore } from 'pinia'

import {
  getCurrentAdmin,
  loginAdmin,
  logoutAdmin,
  refreshAdminToken,
  type AdminAuthStateResponse,
  type AdminLoginRequest,
  type AdminLoginResponse,
  type AdminPermissionSnapshot,
  type AdminSessionSummary,
  type AdminSummary,
} from '../../api/auth'
import type { DashboardResponse } from '../../api/dashboard'
import { clearAuthStorage, loadAuthStorage, saveAuthStorage } from '../../utils/auth'
import { usePermissionStore } from './permission'

interface AuthState {
  token: string
  admin: AdminSummary | null
  session: AdminSessionSummary | null
  restored: boolean
}

let restorePromise: Promise<boolean> | null = null

function buildInitialState(): AuthState {
  const snapshot = loadAuthStorage()
  return {
    token: snapshot?.token || '',
    admin: snapshot?.admin || null,
    session: snapshot?.session || null,
    restored: !snapshot?.token,
  }
}

export const useAuthStore = defineStore('admin-auth', {
  state: (): AuthState => buildInitialState(),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.admin),
    hasToken: (state) => Boolean(state.token),
  },
  actions: {
    async login(payload: AdminLoginRequest) {
      const result = await loginAdmin(payload)
      this.applyLoginPayload(result)
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
          this.applyAuthStatePayload(result)
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
    async refreshToken() {
      const result = await refreshAdminToken()
      this.applyLoginPayload(result)
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
    applyLoginPayload(payload: AdminLoginResponse) {
      this.token = payload.access_token
      this.admin = payload.admin
      this.session = payload.session
      this.restored = true
      this.persist()
      this.applyPermissionSnapshot({
        role_ids: payload.role_ids,
        permission_codes: payload.permission_codes,
        menus: [],
      })
    },
    applyAuthStatePayload(payload: AdminAuthStateResponse) {
      this.admin = payload.admin
      this.session = payload.session
      this.restored = true
      this.persist()
      this.applyPermissionSnapshot(payload)
    },
    applyDashboardPayload(payload: DashboardResponse) {
      this.admin = payload.admin
      this.session = payload.session
      this.restored = true
      this.persist()
      this.applyPermissionSnapshot(payload)
    },
    applyPermissionSnapshot(payload: AdminPermissionSnapshot) {
      const permissionStore = usePermissionStore()
      permissionStore.setSnapshot(payload)
    },
    logout() {
      this.token = ''
      this.admin = null
      this.session = null
      this.restored = true
      clearAuthStorage()
      const permissionStore = usePermissionStore()
      permissionStore.reset()
    },
    persist() {
      if (!this.token) {
        clearAuthStorage()
        return
      }

      saveAuthStorage({
        token: this.token,
        admin: this.admin,
        session: this.session,
      })
    },
  },
})
