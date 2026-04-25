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

export interface AdminLoginResponse {
  access_token: string
  token_type: 'Bearer'
  expires_in: number
  admin: AdminSummary
  role_ids: number[]
  permission_codes: string[]
}
