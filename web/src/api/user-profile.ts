import { request, type WebApiEnvelope } from './request'
import type { AuthStateResponse } from './auth'

export interface UpdateProfileRequest {
  email: string
  display_name?: string | null
}

export interface ChangePasswordRequest {
  current_password: string
  password: string
}

export async function updateProfile(payload: UpdateProfileRequest) {
  const response = await request.patch<WebApiEnvelope<AuthStateResponse>>('/user/profile', payload)
  return response.data.data
}

export async function changePassword(payload: ChangePasswordRequest) {
  const response = await request.post<WebApiEnvelope<Record<string, never>>>('/user/password', payload)
  return response.data.data
}
