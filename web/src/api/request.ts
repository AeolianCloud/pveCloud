import axios from 'axios'

export const request = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

request.interceptors.response.use(
  (response) => {
    if (typeof response.data?.code === 'number' && response.data.code !== 0) {
      return Promise.reject({ response })
    }
    return response
  },
  (error) => {
    return Promise.reject(error)
  },
)

export type WebApiEnvelope<T> = {
  code: number
  message: string
  data: T
}
