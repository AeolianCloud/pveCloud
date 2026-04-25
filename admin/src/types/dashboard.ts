import type { AdminSummary } from './auth'

export interface DashboardMetric {
  key: string
  title: string
  value: number
  unit: string | null
}

export interface AdminMenuItem {
  key: string
  title: string
  path: string
  icon: string | null
  permission_code: string | null
  children?: AdminMenuItem[]
}

export interface DashboardResponse {
  admin: AdminSummary
  role_ids: number[]
  permission_codes: string[]
  menus: AdminMenuItem[]
  metrics: DashboardMetric[]
}
