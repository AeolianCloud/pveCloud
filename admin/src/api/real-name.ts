import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface RealNameUserSummary {
  id: number
  username: string
  email: string
  display_name: string | null
  status: string
}

export interface RealNameApplicationItem {
  id: number
  application_no: string
  user: RealNameUserSummary
  real_name: string
  id_type: string
  id_number_masked: string
  verification_provider: string | null
  provider_application_id: string | null
  provider_status: string | null
  provider_result_code: string | null
  provider_result_message: string | null
  provider_trace_id: string | null
  status: string
  submit_attempt: number
  failure_reason: string | null
  provider_started_at: string | null
  provider_finished_at: string | null
  created_at: string
  updated_at: string
}

export interface RealNameListQuery {
  page?: number
  per_page?: number
  keyword?: string
  status?: string
  id_type?: string
  provider?: string
  provider_status?: string
  date_from?: string
  date_to?: string
}

export interface RealNameReviewRequest {
  status: 'approved' | 'rejected'
  reason?: string
}

export async function getRealNameApplications(query?: RealNameListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<RealNameApplicationItem>>>('/real-name-applications', { params: query })
  return response.data.data
}

export async function getRealNameApplication(id: number) {
  const response = await http.get<ApiEnvelope<RealNameApplicationItem>>(`/real-name-applications/${id}`)
  return response.data.data
}

export async function syncRealNameApplication(id: number) {
  const response = await http.post<ApiEnvelope<RealNameApplicationItem>>(`/real-name-applications/${id}/sync`)
  return response.data.data
}

export async function reviewRealNameApplication(id: number, payload: RealNameReviewRequest) {
  const response = await http.post<ApiEnvelope<RealNameApplicationItem>>(`/real-name-applications/${id}/review`, payload)
  return response.data.data
}
