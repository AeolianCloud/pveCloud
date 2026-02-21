import request from '@/utils/request'
import type { ApiResponse, PageResult } from '@/types'

export interface LoginLog {
  id: number
  admin_user_id: number
  username: string
  ip: string
  user_agent: string
  status: number  // 1 成功  0 失败
  remark: string
  created_at: string
}

export interface LoginLogListParams {
  page_num?: number
  page_size?: number
  username?: string
  status?: number | ''  // 1 | 0 | '' 不过滤
}

// 获取登录日志列表
export function listLoginLogs(params: LoginLogListParams) {
  return request.get<ApiResponse<PageResult<LoginLog>>>('/login-logs', { params })
}
