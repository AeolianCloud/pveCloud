import { defineStore } from 'pinia'

import {
  getCurrentUser,
  login,
  logout,
  type AuthStateResponse,
  type LoginRequest,
  type SessionSummary,
  type UserSummary,
} from '../../api/auth'

const tokenKey = 'pve-web-token'

export const useWebAuthStore = defineStore('web-auth', {
  state: () => ({
    token: localStorage.getItem(tokenKey) ?? '',
    user: null as UserSummary | null,
    session: null as SessionSummary | null,
    restored: false,
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.user),
    displayName: (state) => state.user?.display_name || state.user?.username || '用户',
  },
  actions: {
    setAuth(token: string, state: AuthStateResponse) {
      this.token = token
      this.user = state.user
      this.session = state.session
      localStorage.setItem(tokenKey, token)
    },
    clearAuth() {
      this.token = ''
      this.user = null
      this.session = null
      localStorage.removeItem(tokenKey)
    },
    async login(payload: LoginRequest) {
      const result = await login(payload)
      this.setAuth(result.access_token, result)
      this.restored = true
      return result
    },
    async restore() {
      if (this.restored) return this.isLoggedIn
      if (!this.token) {
        this.restored = true
        return false
      }
      try {
        const result = await getCurrentUser()
        this.user = result.user
        this.session = result.session
        return true
      } catch {
        this.clearAuth()
        return false
      } finally {
        this.restored = true
      }
    },
    async logout() {
      try {
        await logout()
      } finally {
        this.clearAuth()
        this.restored = true
      }
    },
  },
})
