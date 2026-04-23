import { request } from '../lib/http'

export interface DashboardStats {
  total_orders: number
  pending_orders: number
  total_instances: number
  running_instances: number
  total_users: number
  total_tasks: number
  pending_tasks: number
}

export function getDashboardStats(): Promise<DashboardStats> {
  return request<DashboardStats>('/dashboard')
}
