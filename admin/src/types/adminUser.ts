import type { PageResponse } from './audit'
import type { AdminSessionSummary } from './auth'

export type AdminUserStatus = 'active' | 'disabled'

export interface AdminRoleSummary {
  id: number
  code: string
  name: string
}

export interface AdminUserItem {
  id: number
  username: string
  email: string | null
  display_name: string
  status: AdminUserStatus
  role_ids: number[]
  roles: AdminRoleSummary[]
  last_login_at: string | null
  last_login_ip: string | null
  created_at: string
  updated_at: string
}

export interface AdminUserDetail extends AdminUserItem {
  permission_codes: string[]
  sessions: AdminSessionSummary[]
}

export interface AdminUserQuery {
  page?: number
  per_page?: number
  keyword?: string
  status?: AdminUserStatus | ''
  role_id?: number
}

export interface AdminUserCreateRequest {
  username: string
  email?: string | null
  display_name: string
  password: string
  status: AdminUserStatus
  role_ids: number[]
}

export interface AdminUserUpdateRequest {
  email?: string | null
  display_name?: string
  status?: AdminUserStatus
  role_ids?: number[]
}

export interface AdminUserPasswordRequest {
  password: string
}

export type AdminUserPageResponse = PageResponse<AdminUserItem>
