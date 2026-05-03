import axios from 'axios'

const tokenKey = 'pve-web-token'

function isUnauthorizedCode(code: unknown) {
  return typeof code === 'number' && code >= 40100 && code < 40200
}

function clearTokenOnUnauthorized(status: number | undefined, code: unknown) {
  if (status === 401 || isUnauthorizedCode(code)) {
    localStorage.removeItem(tokenKey)
  }
}

export const request = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem(tokenKey)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (response) => {
    clearTokenOnUnauthorized(response.status, response.data?.code)
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
