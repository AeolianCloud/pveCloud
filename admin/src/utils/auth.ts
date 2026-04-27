import type { AdminSessionSummary, AdminSummary } from '../api/auth'

const AUTH_STORAGE_KEY = 'pvecloud_admin_auth'

export interface AuthStorageSnapshot {
  token: string
  admin: AdminSummary | null
  session: AdminSessionSummary | null
}

/**
 * 读取本地登录态快照。
 */
export function loadAuthStorage(): AuthStorageSnapshot | null {
  const raw = localStorage.getItem(AUTH_STORAGE_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as AuthStorageSnapshot
  } catch {
    localStorage.removeItem(AUTH_STORAGE_KEY)
    return null
  }
}

/**
 * 持久化最小登录态快照。
 */
export function saveAuthStorage(snapshot: AuthStorageSnapshot) {
  localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(snapshot))
}

/**
 * 清理本地登录态快照。
 */
export function clearAuthStorage() {
  localStorage.removeItem(AUTH_STORAGE_KEY)
}
