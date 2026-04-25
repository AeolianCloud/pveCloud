import { http } from './http'
import type { ApiEnvelope } from '../types/auth'
import type { DashboardResponse } from '../types/dashboard'

export async function getAdminDashboard(): Promise<DashboardResponse> {
  const response = await http.get<ApiEnvelope<DashboardResponse>>('/dashboard')
  return response.data.data
}
