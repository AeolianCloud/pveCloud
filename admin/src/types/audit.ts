export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  per_page: number
  last_page: number
}

export interface AuditAdminSummary {
  id: number
  username: string
  display_name: string
  email: string | null
}

export interface AuditLogItem {
  id: number
  admin: AuditAdminSummary | null
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

export interface RiskLogItem extends AuditLogItem {
  audit_log_id: number | null
  risk_level: 'medium' | 'high' | 'critical'
  risk_reason: string
}

export interface AuditLogQuery {
  page?: number
  per_page?: number
  admin_id?: number
  action?: string
  object_type?: string
  object_id?: string
  date_from?: string
  date_to?: string
}

export interface RiskLogQuery extends AuditLogQuery {
  risk_level?: 'medium' | 'high' | 'critical' | ''
}
