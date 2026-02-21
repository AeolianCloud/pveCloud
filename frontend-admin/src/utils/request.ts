import axios from 'axios'
import { useAuthStore } from '@/store/auth'
import { router } from '@/router'
import type { ApiResponse } from '@/types'

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// refreshClient 专用于刷新 token：不走本 request 实例的拦截器，避免 401 递归循环。
const refreshClient = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// refreshingPromise 用于合并并发 401：同一时间只发起一次 refresh 请求。
let refreshingPromise: Promise<string> | null = null

async function refreshAccessTokenOnce(): Promise<string> {
  const rtoken = localStorage.getItem('refresh_token')
  if (!rtoken) {
    throw new Error('缺少 refresh token')
  }

  const res = await refreshClient.post<ApiResponse<{ token: string; refresh_token: string }>>(
    '/auth/refresh',
    { refresh_token: rtoken },
  )

  const data = res.data?.data
  if (!data?.token || !data?.refresh_token) {
    throw new Error('刷新失败')
  }

  // 刷新成功：同时更新 access token 与 refresh token
  // 同步到 Pinia（让页面上的登录态立即一致），同时写入 localStorage
  const authStore = useAuthStore()
  authStore.setTokens(data.token, data.refresh_token)

  return data.token
}

// 请求拦截器：自动附加 Authorization header
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：统一处理业务错误码和 HTTP 异常
request.interceptors.response.use(
  (response) => {
    const data: ApiResponse = response.data
    // 业务失败：code 非 0 时抛出，由调用方处理
    if (data.code !== 0) {
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return response
  },
  async (error) => {
    const status = error.response?.status
    const serverMessage =
      (error.response?.data && typeof error.response.data === 'object'
        ? (error.response.data.message as string | undefined)
        : undefined)

    // 401：优先尝试 refresh（有 refresh_token 才尝试），成功后重放原请求
    if (status === 401) {
      const originalConfig = error.config as any

      // 刷新接口本身返回 401 时不再二次 refresh，直接判定登录失效
      if (originalConfig?.url?.includes('/auth/refresh')) {
        const authStore = useAuthStore()
        authStore.logoutLocal()
        router.push('/login')
        return Promise.reject(new Error(serverMessage || '登录已过期，请重新登录'))
      }

      if (!originalConfig?._retry) {
        originalConfig._retry = true

        try {
          // 合并并发 refresh 请求
          if (!refreshingPromise) {
            refreshingPromise = refreshAccessTokenOnce().finally(() => {
              refreshingPromise = null
            })
          }
          const newToken = await refreshingPromise

          // 用新 token 重放原请求
          originalConfig.headers = originalConfig.headers || {}
          originalConfig.headers.Authorization = `Bearer ${newToken}`
          return request(originalConfig)
        } catch {
          // refresh 失败：清理登录态并跳转登录
          const authStore = useAuthStore()
          authStore.logoutLocal()
          router.push('/login')
          return Promise.reject(new Error(serverMessage || '登录已过期，请重新登录'))
        }
      }
    }

    if (status === 403) {
      return Promise.reject(new Error(serverMessage || '无权限访问'))
    }

    // 其他 HTTP 异常：优先透传后端 message，避免前端只显示“Request failed with status code 500”
    if (serverMessage) {
      return Promise.reject(new Error(serverMessage))
    }

    return Promise.reject(error)
  },
)

export default request
