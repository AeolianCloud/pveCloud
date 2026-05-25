import type { AdminAuthStateResponse } from './auth'
import { http, type ApiEnvelope } from '../utils/request'

export interface DashboardMetric {
  key: string
  title: string
  value: number
  unit: string | null
}

export interface DashboardBusinessMetric {
  key: string
  title: string
  value: number
  unit: string | null
  description: string
  target_path: string | null
  target_permission: string | null
  severity: 'info' | 'warning' | 'error'
}

export interface DashboardResponse extends AdminAuthStateResponse {
  metrics: DashboardMetric[]
  business_metrics: DashboardBusinessMetric[]
}

export async function getAdminDashboard() {
  const response = await http.get<ApiEnvelope<DashboardResponse>>('/dashboard')
  return response.data.data
}
