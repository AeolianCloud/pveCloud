import { http } from './http'
import type { ApiEnvelope } from '../types/auth'
import type { AdminSessionPageResponse, AdminSessionQuery } from '../types/adminSession'

export async function getAdminSessions(params: AdminSessionQuery): Promise<AdminSessionPageResponse> {
  const response = await http.get<ApiEnvelope<AdminSessionPageResponse>>('/admin-sessions', { params })
  return response.data.data
}

export async function revokeAdminSession(id: number): Promise<void> {
  await http.post<ApiEnvelope<Record<string, never>>>(`/admin-sessions/${id}/revoke`)
}
