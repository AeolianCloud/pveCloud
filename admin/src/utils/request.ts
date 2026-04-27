import axios, { AxiosError } from 'axios'

import { ADMIN_ROUTE_PATH, normalizeAdminRedirect } from '../router/constants'
import { pinia } from '../store'
import { useAuthStore } from '../store/modules/auth'
import { usePermissionStore } from '../store/modules/permission'
import { clearAuthStorage, loadAuthStorage } from './auth'

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

export const http = axios.create({
  baseURL: '/admin-api',
  timeout: 12_000,
  headers: {
    'Content-Type': 'application/json',
  },
})

http.interceptors.request.use((config) => {
  const snapshot = loadAuthStorage()
  if (snapshot?.token) {
    config.headers.Authorization = `Bearer ${snapshot.token}`
  }
  return config
})

http.interceptors.response.use(
  (response) => {
    const body = response.data as ApiEnvelope<unknown>
    if (isAuthExpiredCode(body.code)) {
      handleExpiredLogin(response.config.url)
      throw new Error(body.message || '未登录或登录已过期')
    }
    if (isForbiddenCode(body.code)) {
      handleForbidden(response.config.url)
      throw new Error(body.message || '当前账号没有访问权限')
    }
    if (body.code !== 0) {
      throw new Error(body.message || '请求失败')
    }
    return response
  },
  (error: AxiosError<ApiEnvelope<null>>) => {
    if (error.response?.status === 401 || isAuthExpiredCode(error.response?.data?.code)) {
      handleExpiredLogin(error.config?.url)
    }
    if (error.response?.status === 403 || isForbiddenCode(error.response?.data?.code)) {
      handleForbidden(error.config?.url)
    }
    throw new Error(buildRequestErrorMessage(error))
  },
)

function isAuthExpiredCode(code?: number) {
  return typeof code === 'number' && code >= 40100 && code < 40200
}

function isForbiddenCode(code?: number) {
  return typeof code === 'number' && code >= 40300 && code < 40400
}

function handleExpiredLogin(url?: string) {
  if (url?.includes('/auth/login')) {
    return
  }

  const authStore = useAuthStore(pinia)
  const permissionStore = usePermissionStore(pinia)
  permissionStore.reset()
  authStore.logout()
  clearAuthStorage()

  const current = normalizeAdminRedirect(
    `${window.location.pathname}${window.location.search}`,
    ADMIN_ROUTE_PATH.dashboard,
  )
  const redirect = current !== ADMIN_ROUTE_PATH.login ? `?redirect=${encodeURIComponent(current)}` : ''
  window.location.replace(`${ADMIN_ROUTE_PATH.login}${redirect}`)
}

function handleForbidden(url?: string) {
  if (url?.includes('/auth/login') || window.location.pathname === ADMIN_ROUTE_PATH.forbidden) {
    return
  }
  window.location.replace(ADMIN_ROUTE_PATH.forbidden)
}

function buildRequestErrorMessage(error: AxiosError<ApiEnvelope<null>>) {
  const serverMessage = error.response?.data?.message?.trim()
  if (serverMessage) {
    return serverMessage
  }

  if (error.code === 'ECONNABORTED') {
    return '请求超时，请稍后重试'
  }

  if (error.response?.status) {
    return `请求失败（${error.response.status}），请稍后重试`
  }

  return '网络请求失败，请稍后重试'
}
