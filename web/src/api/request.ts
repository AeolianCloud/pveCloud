import axios from 'axios'

import { clearStoredWebToken, getStoredWebToken, notifyWebUnauthorized } from '../utils/web-auth'

function isUnauthorizedCode(code: unknown) {
  return typeof code === 'number' && code >= 40100 && code < 40200
}

function clearTokenOnUnauthorized(status: number | undefined, code: unknown) {
  if (status === 401 || isUnauthorizedCode(code)) {
    clearStoredWebToken()
    notifyWebUnauthorized()
  }
}

export const request = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

request.interceptors.request.use((config) => {
  const token = getStoredWebToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    clearTokenOnUnauthorized(response.status, response.data?.code)
    if (typeof response.data?.code === 'number' && response.data.code !== 0) {
      return Promise.reject({ response })
    }
    return response
  },
  (error) => {
    clearTokenOnUnauthorized(error.response?.status, error.response?.data?.code)
    return Promise.reject(error)
  },
)

export type WebApiEnvelope<T> = {
  code: number
  message: string
  data: T
}
