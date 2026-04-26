import { http } from './http'
import type { ApiEnvelope } from '../types/auth'
import type {
  AdminPermissionGroup,
  AdminRoleCreateRequest,
  AdminRoleItem,
  AdminRolePageResponse,
  AdminRoleQuery,
  AdminRoleUpdateRequest,
} from '../types/adminRole'

export async function getAdminRoles(params: AdminRoleQuery): Promise<AdminRolePageResponse> {
  const response = await http.get<ApiEnvelope<AdminRolePageResponse>>('/admin-roles', { params })
  return response.data.data
}

export async function createAdminRole(payload: AdminRoleCreateRequest): Promise<AdminRoleItem> {
  const response = await http.post<ApiEnvelope<AdminRoleItem>>('/admin-roles', payload)
  return response.data.data
}

export async function updateAdminRole(id: number, payload: AdminRoleUpdateRequest): Promise<AdminRoleItem> {
  const response = await http.patch<ApiEnvelope<AdminRoleItem>>(`/admin-roles/${id}`, payload)
  return response.data.data
}

export async function getAdminPermissions(groupName?: string): Promise<AdminPermissionGroup[]> {
  const response = await http.get<ApiEnvelope<AdminPermissionGroup[]>>('/admin-permissions', {
    params: { group_name: groupName || undefined },
  })
  return response.data.data
}
