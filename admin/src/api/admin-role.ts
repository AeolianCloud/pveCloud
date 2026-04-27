import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface AdminRoleItem {
  id: number
  code: string
  name: string
  description: string | null
  status: string
  permission_codes: string[]
  created_at: string
  updated_at: string
}

export interface AdminRoleListQuery {
  page?: number
  per_page?: number
  keyword?: string
  status?: string
}

export interface AdminRoleCreateRequest {
  code: string
  name: string
  description?: string | null
  status: string
  permission_codes?: string[]
}

export interface AdminRoleUpdateRequest {
  name?: string
  description?: string | null
  status?: string
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

export async function getAdminRoles(query?: AdminRoleListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<AdminRoleItem>>>('/admin-roles', { params: query })
  return response.data.data
}

export async function createAdminRole(payload: AdminRoleCreateRequest) {
  const response = await http.post<ApiEnvelope<AdminRoleItem>>('/admin-roles', payload)
  return response.data.data
}

export async function updateAdminRole(id: number, payload: AdminRoleUpdateRequest) {
  const response = await http.patch<ApiEnvelope<AdminRoleItem>>(`/admin-roles/${id}`, payload)
  return response.data.data
}

export async function getAdminPermissions(groupName?: string) {
  const params: Record<string, string> = {}
  if (groupName) {
    params.group_name = groupName
  }
  const response = await http.get<ApiEnvelope<AdminPermissionGroup[]>>('/admin-permissions', { params })
  return response.data.data
}
