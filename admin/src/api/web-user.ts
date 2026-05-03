import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface WebUserItem {
  id: number
  username: string
  email: string
  display_name: string | null
  status: string
  created_at: string
  updated_at: string
}

export interface WebUserListQuery {
  page?: number
  per_page?: number
  keyword?: string
  status?: string
}

export interface WebUserCreateRequest {
  username: string
  email: string
  display_name?: string | null
  password: string
  status: string
}

export interface WebUserUpdateRequest {
  email?: string
  display_name?: string | null
  status?: string
}

export interface WebUserSessionItem {
  session_id: string
  user: WebUserItem
  status: string
  issued_at: string
  expires_at: string
  last_seen_at: string | null
  last_seen_ip: string | null
  user_agent: string | null
  revoked_at: string | null
  revoke_reason: string | null
  created_at: string
}

export interface WebUserSessionListQuery {
  page?: number
  per_page?: number
  user_id?: number
  status?: string
  date_from?: string
  date_to?: string
}

export async function getWebUsers(query?: WebUserListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<WebUserItem>>>('/users', { params: query })
  return response.data.data
}

export async function createWebUser(payload: WebUserCreateRequest) {
  const response = await http.post<ApiEnvelope<WebUserItem>>('/users', payload)
  return response.data.data
}

export async function updateWebUser(id: number, payload: WebUserUpdateRequest) {
  const response = await http.patch<ApiEnvelope<WebUserItem>>(`/users/${id}`, payload)
  return response.data.data
}

export async function resetWebUserPassword(id: number, password: string) {
  await http.post<ApiEnvelope<Record<string, never>>>(`/users/${id}/password`, { password })
}

export async function getWebUserSessions(query?: WebUserSessionListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<WebUserSessionItem>>>('/user-sessions', { params: query })
  return response.data.data
}

export async function revokeWebUserSession(sessionId: string) {
  await http.patch<ApiEnvelope<Record<string, never>>>(`/user-sessions/${encodeURIComponent(sessionId)}`, {
    status: 'revoked',
  })
}
