import type { PageResponse } from './audit'
import type { AuditAdminSummary } from './audit'

export type AdminSessionStatus = 'active' | 'revoked' | 'expired'

export interface AdminSessionItem {
  id: number
  session_id: string
  admin: AuditAdminSummary | null
  status: AdminSessionStatus
  issued_at: string
  expires_at: string
  last_seen_at: string | null
  last_seen_ip: string | null
  user_agent: string | null
  revoked_at: string | null
  revoke_reason: string | null
}

export interface AdminSessionQuery {
  page?: number
  per_page?: number
  admin_id?: number
  status?: AdminSessionStatus | ''
  keyword?: string
}

export type AdminSessionPageResponse = PageResponse<AdminSessionItem>
