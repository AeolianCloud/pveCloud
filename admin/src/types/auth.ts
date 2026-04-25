import type { AdminMenuItem } from './dashboard'

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

export interface AdminSummary {
  id: number
  username: string
  email: string | null
  display_name: string
  status: string
}

export interface AdminLoginRequest {
  username: string
  password: string
}

export interface AdminSessionSummary {
  session_id: string
  issued_at: string
  expires_at: string
}

export interface AdminAuthStateResponse {
  admin: AdminSummary
  role_ids: number[]
  permission_codes: string[]
  menus: AdminMenuItem[]
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
