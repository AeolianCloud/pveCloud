import { request } from '../lib/http'

export interface LoginPayload {
  token: string
  subject_id: number
  subject_type: string
}

export function login(username: string, password: string) {
  return request<LoginPayload>('/auth/login', {
    method: 'POST',
    bodyJson: { username, password },
  })
}
