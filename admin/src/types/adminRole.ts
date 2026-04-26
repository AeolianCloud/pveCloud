import type { PageResponse } from './audit'

export type AdminRoleStatus = 'active' | 'disabled'

export interface AdminRoleItem {
  id: number
  code: string
  name: string
  description: string | null
  status: AdminRoleStatus
  permission_codes: string[]
  created_at: string
  updated_at: string
}

export interface AdminRoleQuery {
  page?: number
  per_page?: number
  keyword?: string
  status?: AdminRoleStatus | ''
}

export interface AdminRoleCreateRequest {
  code: string
  name: string
  description?: string | null
  status: AdminRoleStatus
  permission_codes: string[]
}

export interface AdminRoleUpdateRequest {
  name?: string
  description?: string | null
  status?: AdminRoleStatus
  permission_codes?: string[]
}

export interface AdminPermissionItem {
  id: number
  code: string
  name: string
  group_name: string
  description: string | null
}

export interface AdminPermissionGroup {
  group_name: string
  permissions: AdminPermissionItem[]
}

export type AdminRolePageResponse = PageResponse<AdminRoleItem>
