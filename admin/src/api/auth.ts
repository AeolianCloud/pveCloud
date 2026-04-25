import { http } from './http'
import type {
  AdminAuthStateResponse,
  AdminLoginCaptchaResponse,
  AdminLoginRequest,
  AdminLoginResponse,
  ApiEnvelope,
} from '../types/auth'

export async function getAdminLoginCaptcha(): Promise<AdminLoginCaptchaResponse> {
  const response = await http.get<ApiEnvelope<AdminLoginCaptchaResponse>>('/auth/captcha')
  return response.data.data
}

export async function loginAdmin(payload: AdminLoginRequest): Promise<AdminLoginResponse> {
  const response = await http.post<ApiEnvelope<AdminLoginResponse>>('/auth/login', payload)
  return response.data.data
}

export async function getCurrentAdmin(): Promise<AdminAuthStateResponse> {
  const response = await http.get<ApiEnvelope<AdminAuthStateResponse>>('/auth/me')
  return response.data.data
}

export async function logoutAdmin(): Promise<void> {
  await http.post<ApiEnvelope<Record<string, never>>>('/auth/logout')
}

export async function refreshAdminToken(): Promise<AdminLoginResponse> {
  const response = await http.post<ApiEnvelope<AdminLoginResponse>>('/auth/refresh')
  return response.data.data
}
