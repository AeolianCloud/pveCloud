import request from '@/utils/request'
import type { ApiResponse } from '@/types'

// 按 group 分组的权限结构，与后端 GroupedPermissions 对应
export interface Permission {
  id: number
  name: string
  label: string
  group: string
}

export interface GroupedPermissions {
  group: string
  permissions: Permission[]
}

// 获取全部权限（按 group 分组）
export function listPermissions() {
  return request.get<ApiResponse<GroupedPermissions[]>>('/permissions')
}
