import request from '@/utils/request'
import type { ApiResponse, LoginResult, AdminUser } from '@/types'

// 登录
export function login(username: string, password: string) {
  return request.post<ApiResponse<LoginResult>>('/auth/login', { username, password })
}

// 获取当前登录用户信息
export function getProfile() {
  return request.get<ApiResponse<AdminUser>>('/profile')
}
