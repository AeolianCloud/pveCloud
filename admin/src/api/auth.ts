import { http, type ApiEnvelope } from '../utils/request'

export interface AdminSummary {
  id: number
  username: string
  email: string | null
  display_name: string
  status: string
}

export interface AdminMenuItem {
  key: string
  title: string
  path: string
  icon: string | null
  permission_code: string | null
  children?: AdminMenuItem[]
}

export interface AdminSessionSummary {
  session_id: string
  issued_at: string
  expires_at: string
}

export interface AdminPermissionSnapshot {
  role_ids: number[]
  permission_codes: string[]
  menus: AdminMenuItem[]
}

export interface AdminLoginRequest {
  username: string
  password: string
  captcha_id: string
  captcha_code: string
}

export interface AdminLoginCaptchaResponse {
  captcha_id: string
  image: string
  expires_in: number
}

export interface AdminAuthStateResponse extends AdminPermissionSnapshot {
  admin: AdminSummary
  session: AdminSessionSummary
}

export interface AdminLoginResponse {
  access_token: string
  token_type: 'Bearer'
  expires_in: number
  admin: AdminSummary
  role_ids: number[]
  permission_codes: string[]
  session: AdminSessionSummary
}

export async function getAdminLoginCaptcha() {
  const response = await http.get<ApiEnvelope<AdminLoginCaptchaResponse>>('/auth/captcha')
  return response.data.data
}

export async function loginAdmin(payload: AdminLoginRequest) {
  const response = await http.post<ApiEnvelope<AdminLoginResponse>>('/auth/login', payload)
  return response.data.data
}

export async function getCurrentAdmin() {
  const response = await http.get<ApiEnvelope<AdminAuthStateResponse>>('/auth/me')
  return response.data.data
}

export async function logoutAdmin() {
  await http.post<ApiEnvelope<Record<string, never>>>('/auth/logout')
}

export async function refreshAdminToken() {
  const response = await http.post<ApiEnvelope<AdminLoginResponse>>('/auth/refresh')
  return response.data.data
}
