import request from '@/utils/request'
import type { ApiResponse, LoginResult, RefreshResult, AdminUser } from '@/types'

// 登录
export function login(username: string, password: string) {
  return request.post<ApiResponse<LoginResult>>('/auth/login', { username, password })
}

// 获取当前登录用户信息
export function getProfile() {
  return request.get<ApiResponse<AdminUser>>('/profile')
}

// 刷新 Access Token（同时旋转 Refresh Token）
export function refreshToken(refresh_token: string) {
  return request.post<ApiResponse<RefreshResult>>('/auth/refresh', { refresh_token })
}

// 退出登录（服务端撤销会话，使 token 立即失效）
export function logout() {
  return request.post<ApiResponse<null>>('/auth/logout')
}
