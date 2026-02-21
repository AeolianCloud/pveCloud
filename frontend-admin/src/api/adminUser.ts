import request from '@/utils/request'
import type { ApiResponse, PageResult } from '@/types'
import type { AdminUser } from '@/types'

// 列表查询参数
export interface AdminUserListParams {
  page_num?: number
  page_size?: number
  keyword?: string
}

// 新建请求体
export interface CreateAdminUserReq {
  username: string
  password: string
  nickname?: string
  email?: string
  role_ids: number[]
}

// 更新请求体
export interface UpdateAdminUserReq {
  nickname?: string
  email?: string
  role_ids?: number[]
}

// 获取管理员列表
export function listAdminUsers(params: AdminUserListParams) {
  return request.get<ApiResponse<PageResult<AdminUser>>>('/admin-users', { params })
}

// 新建管理员
export function createAdminUser(data: CreateAdminUserReq) {
  return request.post<ApiResponse<AdminUser>>('/admin-users', data)
}

// 更新管理员
export function updateAdminUser(id: number, data: UpdateAdminUserReq) {
  return request.put<ApiResponse<AdminUser>>(`/admin-users/${id}`, data)
}

// 切换启用/禁用状态
export function setAdminUserStatus(id: number, status: 0 | 1) {
  return request.patch<ApiResponse<null>>(`/admin-users/${id}/status`, { status })
}

// 删除管理员
export function deleteAdminUser(id: number) {
  return request.delete<ApiResponse<null>>(`/admin-users/${id}`)
}
