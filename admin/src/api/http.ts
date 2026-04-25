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
    if (body.code !== 0) {
      throw new Error(body.message || '请求失败')
    }
    return response
  },
  (error: AxiosError<ApiEnvelope<null>>) => {
    const message = error.response?.data?.message || error.message || '网络请求失败'
    throw new Error(message)
  },
)
