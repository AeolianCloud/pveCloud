import axios, { AxiosError } from 'axios'

import { useAuthStore } from '../stores/auth'
import type { ApiEnvelope } from '../types/auth'

export const http = axios.create({
  baseURL: '/admin-api',
  timeout: 12_000,
  headers: {
    'Content-Type': 'application/json',
  },
})

http.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

http.interceptors.response.use(
  (response) => {
    const body = response.data as ApiEnvelope<unknown>
    if (body.code >= 40100 && body.code < 40200) {
      handleExpiredLogin(response.config.url)
      throw new Error(body.message || '未登录或登录已过期')
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
    const message = error.response?.data?.message || error.message || '网络请求失败'
    throw new Error(message)
  },
)

function isAuthExpiredCode(code?: number) {
  return typeof code === 'number' && code >= 40100 && code < 40200
}

function handleExpiredLogin(url?: string) {
  if (url?.includes('/auth/login')) {
    return
  }

  const auth = useAuthStore()
  if (auth.isLoggedIn) {
    auth.logout()
  }

  const current = `${window.location.pathname}${window.location.search}`
  const redirect = current && current !== '/login' ? `?redirect=${encodeURIComponent(current)}` : ''
  window.location.replace(`/login${redirect}`)
}
