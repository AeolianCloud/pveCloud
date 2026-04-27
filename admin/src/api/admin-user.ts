import { http, type ApiEnvelope } from '../utils/request'

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
  status: string
  role_ids: number[]
  roles: AdminRoleSummary[]
  last_login_at: string | null
  last_login_ip: string | null
  created_at: string
  updated_at: string
}

export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  per_page: number
  last_page: number
}

export interface AdminUserListQuery {
  page?: number
  per_page?: number
  keyword?: string
  status?: string
  role_id?: number
}

export interface AdminUserCreateRequest {
  username: string
  email?: string | null
  display_name: string
  password: string
  status: string
  role_ids?: number[]
}

export interface AdminUserUpdateRequest {
  email?: string | null
  display_name?: string
  status?: string
  role_ids?: number[]
}

export async function getAdminUsers(query?: AdminUserListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<AdminUserItem>>>('/admin-users', { params: query })
  return response.data.data
}

export async function createAdminUser(payload: AdminUserCreateRequest) {
  const response = await http.post<ApiEnvelope<AdminUserItem>>('/admin-users', payload)
  return response.data.data
}

export async function getAdminUser(id: number) {
  const response = await http.get<ApiEnvelope<AdminUserItem>>(`/admin-users/${id}`)
  return response.data.data
}

export async function updateAdminUser(id: number, payload: AdminUserUpdateRequest) {
  const response = await http.patch<ApiEnvelope<AdminUserItem>>(`/admin-users/${id}`, payload)
  return response.data.data
}

export async function resetAdminUserPassword(id: number, password: string) {
  await http.post<ApiEnvelope<null>>(`/admin-users/${id}/password`, { password })
}
