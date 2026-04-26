import { http } from './http'
import type { ApiEnvelope } from '../types/auth'
import type {
  AdminUserCreateRequest,
  AdminUserDetail,
  AdminUserItem,
  AdminUserPageResponse,
  AdminUserPasswordRequest,
  AdminUserQuery,
  AdminUserUpdateRequest,
} from '../types/adminUser'

export async function getAdminUsers(params: AdminUserQuery): Promise<AdminUserPageResponse> {
  const response = await http.get<ApiEnvelope<AdminUserPageResponse>>('/admin-users', { params })
  return response.data.data
}

export async function createAdminUser(payload: AdminUserCreateRequest): Promise<AdminUserItem> {
  const response = await http.post<ApiEnvelope<AdminUserItem>>('/admin-users', payload)
  return response.data.data
}

export async function getAdminUser(id: number): Promise<AdminUserDetail> {
  const response = await http.get<ApiEnvelope<AdminUserDetail>>(`/admin-users/${id}`)
  return response.data.data
}

export async function updateAdminUser(id: number, payload: AdminUserUpdateRequest): Promise<AdminUserItem> {
  const response = await http.patch<ApiEnvelope<AdminUserItem>>(`/admin-users/${id}`, payload)
  return response.data.data
}

export async function resetAdminUserPassword(id: number, payload: AdminUserPasswordRequest): Promise<void> {
  await http.post<ApiEnvelope<Record<string, never>>>(`/admin-users/${id}/password`, payload)
}
