import request from '@/utils/request'
import type { ApiResponse, PageResult, AdminRole } from '@/types'

export interface RoleListParams {
  page_num?: number
  page_size?: number
  keyword?: string
}

export interface CreateRoleReq {
  name: string
  label: string
  description?: string
  sort?: number
}

export interface UpdateRoleReq {
  label: string
  description?: string
  sort?: number
}

// 获取角色列表
export function listRoles(params?: RoleListParams) {
  return request.get<ApiResponse<PageResult<AdminRole>>>('/roles', { params })
}

// 获取角色详情（含 permissions）
export function getRoleByID(id: number) {
  return request.get<ApiResponse<AdminRole>>(`/roles/${id}`)
}

// 新建角色
export function createRole(data: CreateRoleReq) {
  return request.post<ApiResponse<AdminRole>>('/roles', data)
}

// 更新角色
export function updateRole(id: number, data: UpdateRoleReq) {
  return request.put<ApiResponse<AdminRole>>(`/roles/${id}`, data)
}

// 删除角色
export function deleteRole(id: number) {
  return request.delete<ApiResponse<null>>(`/roles/${id}`)
}

// 替换角色的权限列表（传空数组则清空）
export function assignPermissions(id: number, permissionIds: number[]) {
  return request.put<ApiResponse<null>>(`/roles/${id}/permissions`, { permission_ids: permissionIds })
}
