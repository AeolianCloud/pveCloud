import axios from 'axios'

const TOKEN_KEY = 'token'

export const request = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    if (typeof response.data?.code === 'number' && response.data.code !== 0) {
      if (response.data.code >= 40100 && response.data.code < 40200) {
        clearStoredAuth()
      }
      return Promise.reject({ response })
    }
    return response
  },
  (error) => {
    if (error?.response?.status === 401) {
      clearStoredAuth()
    }
    return Promise.reject(error)
  },
)

function clearStoredAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem('user')
  localStorage.removeItem('session')
  localStorage.removeItem('token_expires_at')
  if (window.location.pathname.startsWith('/user')) {
    const redirect = `${window.location.pathname}${window.location.search}`
    window.location.assign(`/login?redirect=${encodeURIComponent(redirect)}`)
  }
}

export function getApiErrorMessage(error: unknown, fallback: string) {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { message?: string } } }).response
    if (response?.data?.message) {
      return response.data.message
    }
  }
  return fallback
}

export type WebApiEnvelope<T> = {
  code: number
  message: string
  data: T
}
