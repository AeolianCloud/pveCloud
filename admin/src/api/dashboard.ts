import type { AdminAuthStateResponse } from './auth'
import { http, type ApiEnvelope } from '../utils/request'

export interface DashboardMetric {
  key: string
  title: string
  value: number
  unit: string | null
}

export interface DashboardResponse extends AdminAuthStateResponse {
  metrics: DashboardMetric[]
}

export async function getAdminDashboard() {
  const response = await http.get<ApiEnvelope<DashboardResponse>>('/dashboard')
  return response.data.data
}
