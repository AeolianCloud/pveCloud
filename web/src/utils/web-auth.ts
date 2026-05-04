export const webAuthTokenKey = 'pve-web-token'
export const webAuthTokenExpiresAtKey = 'pve-web-token-expires-at'
export const webAuthUnauthorizedEvent = 'pve-web-auth:unauthorized'

export function getStoredWebToken() {
  return localStorage.getItem(webAuthTokenKey) ?? ''
}

export function setStoredWebToken(token: string) {
  localStorage.setItem(webAuthTokenKey, token)
}

export function getStoredWebTokenExpiresAt() {
  const value = Number(localStorage.getItem(webAuthTokenExpiresAtKey))
  return Number.isFinite(value) ? value : 0
}

export function setStoredWebTokenExpiresAt(expiresAt: number) {
  localStorage.setItem(webAuthTokenExpiresAtKey, String(expiresAt))
}

export function clearStoredWebToken() {
  localStorage.removeItem(webAuthTokenKey)
  localStorage.removeItem(webAuthTokenExpiresAtKey)
}

export function notifyWebUnauthorized() {
  window.dispatchEvent(new CustomEvent(webAuthUnauthorizedEvent))
}

export function resolveWebRedirect(value: unknown, fallback = '/user') {
  if (typeof value !== 'string') return fallback
  if (!value.startsWith('/') || value.startsWith('//')) return fallback
  return value
}
