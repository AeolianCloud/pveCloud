import { request, type WebApiEnvelope } from './request'
import type { RealNameConfig } from './site-config'

export interface RealNameApplicationSummary {
  application_no: string
  real_name: string
  id_type: string
  id_number_masked: string
  verification_provider: string | null
  provider_status: string | null
  status: string
  failure_reason: string | null
  submit_attempt: number
  created_at: string
  verified_at: string | null
}

export interface RealNameProviderAction {
  provider: string
  action_type: string
  redirect_url: string
  expires_at: string | null
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
  provider?: 'alipay' | 'wechat' | 'manual'
}

export interface RealNameSubmitResponse {
  application: RealNameApplicationSummary
  provider_action: RealNameProviderAction
}

export interface RealNameSyncRequest {
  application_no?: string
}

export async function getRealNameStatus() {
  const response = await request.get<WebApiEnvelope<RealNameStatusResponse>>('/user/real-name')
  return response.data.data
}

export async function submitRealName(payload: RealNameSubmitRequest) {
  const response = await request.post<WebApiEnvelope<RealNameSubmitResponse>>('/user/real-name', payload)
  return response.data.data
}

export async function syncRealName(payload: RealNameSyncRequest = {}) {
  const response = await request.post<WebApiEnvelope<RealNameStatusResponse>>('/user/real-name/sync', payload)
  return response.data.data
}
