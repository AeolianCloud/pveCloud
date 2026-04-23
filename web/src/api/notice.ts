import { request } from '../lib/http'

export interface Notice {
  id: number
  user_id: number
  title: string
  body: string
  type: string
  is_read: boolean
  created_at: string
}

export function listNotices(limit = 20): Promise<Notice[]> {
  return request<Notice[]>(`/notices?limit=${limit}`)
}

export function markNoticeRead(id: number): Promise<{ status: string }> {
  return request<{ status: string }>(`/notices/${id}/read`, { method: 'PUT' })
}
