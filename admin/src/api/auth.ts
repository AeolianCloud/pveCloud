import { http } from './http'
import type { AdminLoginRequest, AdminLoginResponse, ApiEnvelope } from '../types/auth'

export async function loginAdmin(payload: AdminLoginRequest): Promise<AdminLoginResponse> {
  const response = await http.post<ApiEnvelope<AdminLoginResponse>>('/auth/login', payload)
  return response.data.data
}
