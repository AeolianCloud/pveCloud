import { request } from '../lib/http'

export interface LoginPayload {
  token: string
  subject_id: number
  subject_type: string
}

export interface RegisterPayload {
  token: string
  user_id: number
  user_no: string
  subject_type: string
}

export function login(phone: string, password: string) {
  return request<LoginPayload>('/auth/login', {
    method: 'POST',
    bodyJson: { phone, password },
  })
}

export function register(phone: string, email: string, password: string) {
  return request<RegisterPayload>('/auth/register', {
    method: 'POST',
    bodyJson: { phone, email, password },
  })
}
