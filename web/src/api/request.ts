import axios from 'axios'

export const request = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('pve-web-token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export type WebApiEnvelope<T> = {
  code: string
  message: string
  data: T
}
