import request from '@/utils/request'
import type { ApiResponse } from '@/types'
import type { AdminMenuNode } from '@/types'

// 获取当前用户可见菜单树（后端已按权限/超管可见性裁剪）
export function getMyMenus() {
  return request.get<ApiResponse<AdminMenuNode[]>>('/menus/my')
}

// 获取完整菜单树（仅 super_admin）
export function listMenus() {
  return request.get<ApiResponse<AdminMenuNode[]>>('/menus')
}

// 新建菜单请求体（path 为空表示目录节点）
export interface CreateMenuReq {
  parent_id: number
  title: string
  path?: string
  permission?: string
  super_admin_only: 0 | 1
  icon?: string
  sort: number
  visible: 0 | 1
}

// 新建菜单（仅 super_admin）
export function createMenu(data: CreateMenuReq) {
  return request.post<ApiResponse<AdminMenuNode>>('/menus', data)
}

// 更新菜单请求体
export interface UpdateMenuReq extends CreateMenuReq {}

// 更新菜单（仅 super_admin）
export function updateMenu(id: number, data: UpdateMenuReq) {
  return request.put<ApiResponse<AdminMenuNode>>(`/menus/${id}`, data)
}

// 删除菜单（软删除，仅 super_admin）
export function deleteMenu(id: number) {
  return request.delete<ApiResponse<null>>(`/menus/${id}`)
}

