import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface AdminSessionItem {
  session_id: string
  admin_id: number
  admin_username: string
  admin_display_name: string
  admin_email: string | null
  status: string
  issued_at: string
  expires_at: string
  last_seen_at: string | null
  last_seen_ip: string | null
  user_agent: string | null
  revoked_at: string | null
  revoke_reason: string | null
  is_current: boolean
}

export interface AdminSessionListQuery {
  page?: number
  per_page?: number
  keyword?: string
  status?: string
}

export async function getAdminSessions(query?: AdminSessionListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<AdminSessionItem>>>('/admin-sessions', { params: query })
  return response.data.data
}

export async function revokeAdminSession(sessionId: string) {
  await http.patch<ApiEnvelope<Record<string, never>>>(`/admin-sessions/${encodeURIComponent(sessionId)}`, {
    status: 'revoked',
  })
}
