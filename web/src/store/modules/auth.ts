import { defineStore } from 'pinia'

import {
  getCurrentUser,
  login,
  logout,
  refreshToken,
  register,
  type AuthStateResponse,
  type LoginRequest,
  type LoginResponse,
  type RegisterRequest,
  type SessionSummary,
  type UserSummary,
} from '../../api/auth'
import {
  clearStoredWebToken,
  getStoredWebTokenExpiresAt,
  getStoredWebToken,
  setStoredWebTokenExpiresAt,
  setStoredWebToken,
} from '../../utils/web-auth'

let refreshTimer: ReturnType<typeof window.setTimeout> | undefined
let refreshPromise: Promise<boolean> | undefined
const refreshLeadMs = 60_000

export const useWebAuthStore = defineStore('web-auth', {
  state: () => ({
    token: getStoredWebToken(),
    tokenExpiresAt: getStoredWebTokenExpiresAt(),
    user: null as UserSummary | null,
    session: null as SessionSummary | null,
    restored: false,
  }),
  getters: {
    isLoggedIn: (state) => Boolean(state.token && state.user),
    displayName: (state) => state.user?.display_name || state.user?.username || '用户',
  },
  actions: {
    setAuth(token: string, state: AuthStateResponse, expiresAt?: number) {
      this.token = token
      this.user = state.user
      this.session = state.session
      this.tokenExpiresAt = expiresAt ?? Date.parse(state.session.expires_at)
      setStoredWebToken(token)
      setStoredWebTokenExpiresAt(this.tokenExpiresAt)
      this.scheduleRefresh()
    },
    setAuthState(state: AuthStateResponse) {
      this.user = state.user
      this.session = state.session
    },
    clearAuth() {
      this.token = ''
      this.tokenExpiresAt = 0
      this.user = null
      this.session = null
      clearStoredWebToken()
      this.clearRefreshTimer()
    },
    handleUnauthorized() {
      this.clearAuth()
      this.restored = true
    },
    async login(payload: LoginRequest) {
      const result = await login(payload)
      this.applyLoginResponse(result)
      this.restored = true
      return result
    },
    async register(payload: RegisterRequest) {
      const result = await register(payload)
      this.applyLoginResponse(result)
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
        this.tokenExpiresAt = Date.parse(result.session.expires_at)
        setStoredWebTokenExpiresAt(this.tokenExpiresAt)
        this.scheduleRefresh()
        this.restored = true
        return true
      } catch {
        this.handleUnauthorized()
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
    applyLoginResponse(result: LoginResponse) {
      const expiresAt = Date.now() + result.expires_in * 1000
      this.setAuth(result.access_token, result, expiresAt)
    },
    clearRefreshTimer() {
      if (refreshTimer) {
        window.clearTimeout(refreshTimer)
        refreshTimer = undefined
      }
    },
    scheduleRefresh() {
      this.clearRefreshTimer()
      if (!this.token || !this.tokenExpiresAt) return
      const delay = Math.max(this.tokenExpiresAt - Date.now() - refreshLeadMs, 5_000)
      refreshTimer = window.setTimeout(() => {
        void this.refresh()
      }, delay)
    },
    async refresh() {
      if (refreshPromise) return refreshPromise
      if (!this.token) return false
      refreshPromise = refreshToken()
        .then((result) => {
          this.applyLoginResponse(result)
          this.restored = true
          return true
        })
        .catch(() => {
          this.handleUnauthorized()
          return false
        })
        .finally(() => {
          refreshPromise = undefined
        })
      return refreshPromise
    },
  },
})
