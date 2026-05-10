import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import {
  getCurrentUser,
  login,
  logout as logoutApi,
  refreshToken,
  register,
  type AuthStateResponse,
  type LoginRequest,
  type LoginResponse,
  type RegisterRequest,
  type SessionSummary,
  type UserSummary,
} from '../api/auth'

const TOKEN_KEY = 'token'
const USER_KEY = 'user'
const SESSION_KEY = 'session'
const TOKEN_EXPIRES_AT_KEY = 'token_expires_at'

function readJson<T>(key: string): T | null {
  const raw = localStorage.getItem(key)
  if (!raw) {
    return null
  }
  try {
    return JSON.parse(raw) as T
  } catch {
    localStorage.removeItem(key)
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<UserSummary | null>(readJson<UserSummary>(USER_KEY))
  const session = ref<SessionSummary | null>(readJson<SessionSummary>(SESSION_KEY))
  const tokenExpiresAt = ref<number>(Number(localStorage.getItem(TOKEN_EXPIRES_AT_KEY) || 0))
  const restoring = ref(false)

  const isAuthenticated = computed(() => Boolean(token.value && user.value))

  const setToken = (newToken: string | null) => {
    token.value = newToken
    if (newToken) {
      localStorage.setItem(TOKEN_KEY, newToken)
    } else {
      localStorage.removeItem(TOKEN_KEY)
    }
  }

  const setUser = (newUser: UserSummary | null) => {
    user.value = newUser
    if (newUser) {
      localStorage.setItem(USER_KEY, JSON.stringify(newUser))
    } else {
      localStorage.removeItem(USER_KEY)
    }
  }

  const setSession = (newSession: SessionSummary | null) => {
    session.value = newSession
    if (newSession) {
      localStorage.setItem(SESSION_KEY, JSON.stringify(newSession))
    } else {
      localStorage.removeItem(SESSION_KEY)
    }
  }

  const setTokenExpiresAt = (expiresAt: number) => {
    tokenExpiresAt.value = expiresAt
    if (expiresAt > 0) {
      localStorage.setItem(TOKEN_EXPIRES_AT_KEY, String(expiresAt))
    } else {
      localStorage.removeItem(TOKEN_EXPIRES_AT_KEY)
    }
  }

  const applyAuth = (data: LoginResponse) => {
    setToken(data.access_token)
    setUser(data.user)
    setSession(data.session)
    setTokenExpiresAt(Date.now() + data.expires_in * 1000)
  }

  const applyAuthState = (data: AuthStateResponse) => {
    setUser(data.user)
    setSession(data.session)
  }

  const clearAuth = () => {
    setToken(null)
    setUser(null)
    setSession(null)
    setTokenExpiresAt(0)
  }

  const loginWithPassword = async (payload: LoginRequest) => {
    const data = await login(payload)
    applyAuth(data)
  }

  const registerAccount = async (payload: RegisterRequest) => {
    const data = await register(payload)
    applyAuth(data)
  }

  const refreshCurrentToken = async () => {
    if (!token.value) {
      return false
    }
    try {
      const data = await refreshToken()
      applyAuth(data)
      return true
    } catch {
      clearAuth()
      return false
    }
  }

  const restoreAuth = async () => {
    if (!token.value) {
      clearAuth()
      return false
    }
    if (restoring.value) {
      return isAuthenticated.value
    }
    restoring.value = true
    try {
      if (tokenExpiresAt.value && tokenExpiresAt.value - Date.now() < 60_000) {
        const refreshed = await refreshCurrentToken()
        if (!refreshed) {
          return false
        }
      }
      const data = await getCurrentUser()
      applyAuthState(data)
      return true
    } catch {
      clearAuth()
      return false
    } finally {
      restoring.value = false
    }
  }

  const logout = async () => {
    try {
      if (token.value) {
        await logoutApi()
      }
    } finally {
      clearAuth()
    }
  }

  return {
    token,
    user,
    session,
    restoring,
    isAuthenticated,
    setToken,
    setUser,
    setSession,
    applyAuthState,
    clearAuth,
    loginWithPassword,
    registerAccount,
    refreshCurrentToken,
    restoreAuth,
    logout,
  }
})
