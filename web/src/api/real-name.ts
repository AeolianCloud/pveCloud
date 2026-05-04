import { request, type WebApiEnvelope } from './request'
import type { RealNameConfig } from './site-config'

export interface RealNameApplicationSummary {
  application_no: string
  real_name: string
  id_type: string
  id_number_masked: string
  status: string
  reject_reason: string | null
  submit_attempt: number
  created_at: string
  reviewed_at: string | null
}

export interface RealNameStatusResponse {
  status: 'unverified' | 'pending' | 'approved' | 'rejected'
  application: RealNameApplicationSummary | null
  config: RealNameConfig
}

export interface RealNameSubmitRequest {
  real_name: string
  id_type: 'id_card'
  id_number: string
  id_card_front_file_id?: number
  id_card_back_file_id?: number
  hold_card_file_id?: number
}

export interface RealNameFileUploadResponse {
  id: number
  original_name: string
  mime_type: string
  size: number
  created_at: string
}

export async function getRealNameStatus() {
  const response = await request.get<WebApiEnvelope<RealNameStatusResponse>>('/user/real-name')
  return response.data.data
}

export async function uploadRealNameFile(file: File) {
  const form = new FormData()
  form.append('file', file)
  const response = await request.post<WebApiEnvelope<RealNameFileUploadResponse>>('/user/real-name/files', form)
  return response.data.data
}

export async function submitRealName(payload: RealNameSubmitRequest) {
  const response = await request.post<WebApiEnvelope<RealNameApplicationSummary>>('/user/real-name', payload)
  return response.data.data
}
