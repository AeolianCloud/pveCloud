import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface RealNameUserSummary {
  id: number
  username: string
  email: string
  display_name: string | null
  status: string
}

export interface RealNameFileSummary {
  id: number
  original_name: string
  mime_type: string
  size: number
  created_at: string
}

export interface RealNameApplicationItem {
  id: number
  application_no: string
  user: RealNameUserSummary
  real_name: string
  id_type: string
  id_number_masked: string
  status: string
  submit_attempt: number
  review_admin: RealNameUserSummary | null
  reviewed_at: string | null
  reject_reason: string | null
  id_card_front_file?: RealNameFileSummary | null
  id_card_back_file?: RealNameFileSummary | null
  hold_card_file?: RealNameFileSummary | null
  created_at: string
  updated_at: string
}

export interface RealNameListQuery {
  page?: number
  per_page?: number
  keyword?: string
  status?: string
  id_type?: string
  date_from?: string
  date_to?: string
}

export async function getRealNameApplications(query?: RealNameListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<RealNameApplicationItem>>>('/real-name-applications', { params: query })
  return response.data.data
}

export async function getRealNameApplication(id: number) {
  const response = await http.get<ApiEnvelope<RealNameApplicationItem>>(`/real-name-applications/${id}`)
  return response.data.data
}

export async function reviewRealNameApplication(id: number, status: 'approved' | 'rejected', rejectReason?: string) {
  const response = await http.post<ApiEnvelope<RealNameApplicationItem>>(`/real-name-applications/${id}/review`, {
    status,
    reject_reason: rejectReason || undefined,
  })
  return response.data.data
}
