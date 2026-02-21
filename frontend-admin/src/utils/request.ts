import axios from 'axios'
import { useAuthStore } from '@/store/auth'
import { router } from '@/router'
import type { ApiResponse } from '@/types'

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

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
  (error) => {
    const status = error.response?.status
    if (status === 401) {
      // Token 失效，清除登录态并跳转登录页
      const authStore = useAuthStore()
      authStore.logout()
      router.push('/login')
      return Promise.reject(new Error('登录已过期，请重新登录'))
    }
    if (status === 403) {
      return Promise.reject(new Error('无权限访问'))
    }
    return Promise.reject(error)
  },
)

export default request
