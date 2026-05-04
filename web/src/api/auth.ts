import { request, type WebApiEnvelope } from './request'

export interface UserSummary {
  id: number
  username: string
  email: string
  display_name: string | null
  status: string
}

export interface SessionSummary {
  session_id: string
  issued_at: string
  expires_at: string
}

export interface AuthStateResponse {
  user: UserSummary
  session: SessionSummary
}

export interface LoginRequest {
  account: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  display_name?: string | null
}

export interface PasswordResetRequest {
  email: string
}

export interface PasswordResetConfirmRequest {
  token: string
  password: string
}

export interface LoginResponse extends AuthStateResponse {
  access_token: string
  token_type: string
  expires_in: number
}

export async function login(payload: LoginRequest) {
  const response = await request.post<WebApiEnvelope<LoginResponse>>('/auth/login', payload)
  return response.data.data
}

export async function register(payload: RegisterRequest) {
  const response = await request.post<WebApiEnvelope<LoginResponse>>('/auth/register', payload)
  return response.data.data
}

export async function getCurrentUser() {
  const response = await request.get<WebApiEnvelope<AuthStateResponse>>('/auth/me')
  return response.data.data
}

export async function refreshToken() {
  const response = await request.post<WebApiEnvelope<LoginResponse>>('/auth/refresh')
  return response.data.data
}

export async function logout() {
  const response = await request.post<WebApiEnvelope<Record<string, never>>>('/auth/logout')
  return response.data.data
}

export async function requestPasswordReset(payload: PasswordResetRequest) {
  const response = await request.post<WebApiEnvelope<Record<string, never>>>('/auth/password-reset/request', payload)
  return response.data.data
}

export async function confirmPasswordReset(payload: PasswordResetConfirmRequest) {
  const response = await request.post<WebApiEnvelope<Record<string, never>>>('/auth/password-reset/confirm', payload)
  return response.data.data
}
