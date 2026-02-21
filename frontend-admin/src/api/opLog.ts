import request from '@/utils/request'
import type { ApiResponse, PageResult } from '@/types'

// 操作日志条目
export interface OpLog {
  id: number
  admin_user_id: number
  username: string
  module: string
  action: string
  target_id: number
  target_label: string
  status: number
  ip: string
  created_at: string
}

export interface OpLogListParams {
  page_num?: number
  page_size?: number
  username?: string
  module?: string
  action?: string
}

// 获取操作日志列表
export function listOpLogs(params?: OpLogListParams) {
  return request.get<ApiResponse<PageResult<OpLog>>>('/op-logs', { params })
}
