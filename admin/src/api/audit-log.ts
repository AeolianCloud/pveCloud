import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface AuditAdminSummary {
  id: number
  username: string
  display_name: string
  email: string | null
}

export interface AuditLogItem {
  id: number
  admin: AuditAdminSummary | null
  session_id: string | null
  request_id: string | null
  request_method: string | null
  request_path: string | null
  action: string
  object_type: string
  object_id: string | null
  before_data: string | null
  after_data: string | null
  ip: string | null
  user_agent: string | null
  remark: string | null
  created_at: string
}

export interface AuditLogListQuery {
  page?: number
  per_page?: number
  admin_id?: number
  action?: string
  object_type?: string
  object_id?: string
  date_from?: string
  date_to?: string
}

export async function getAuditLogs(query?: AuditLogListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<AuditLogItem>>>('/audit-logs', { params: query })
  return response.data.data
}
