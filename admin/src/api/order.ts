import { http, type ApiEnvelope } from '../utils/request'
import type { PaginatedData } from './admin-user'

export interface OrderUserSummary {
  id: number
  username: string
  email: string
  display_name: string | null
}

export interface AdminOrderItem {
  order_no: string
  user: OrderUserSummary
  status: 'pending' | 'cancelled' | 'closed'
  product_name: string
  plan_name: string
  billing_cycle: string
  total_amount_cents: number
  currency: string
  admin_note: string | null
  created_at: string
  cancelled_at: string | null
  closed_at: string | null
}

export interface AdminOrderDetail extends AdminOrderItem {
  user_note: string | null
  cancel_reason: string | null
  closed_reason: string | null
  product_no: string
  product_type: string
  product_summary: string | null
  plan_no: string
  plan_code: string
  plan_summary: string | null
  cpu_cores: number
  memory_mb: number
  system_disk_gb: number
  data_disk_gb: number
  bandwidth_mbps: number
  traffic_gb: number | null
  public_ip_count: number
  virtualization: string
  architecture: string
  price_cents: number
  original_price_cents: number | null
  quantity: number
  region_no: string
  region_code: string
  region_name: string
  template_no: string
  template_code: string
  template_name: string
  os_family: string
  os_distribution: string
  os_version: string
  os_architecture: string
}

export async function getOrders(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<AdminOrderItem>>>('/orders', { params })
  return response.data.data
}

export async function getOrderDetail(orderNo: string) {
  const response = await http.get<ApiEnvelope<AdminOrderDetail>>(`/orders/${orderNo}`)
  return response.data.data
}

export async function updateOrderAdminNote(orderNo: string, adminNote: string | null) {
  const response = await http.patch<ApiEnvelope<AdminOrderDetail>>(`/orders/${orderNo}/admin-note`, { admin_note: adminNote })
  return response.data.data
}

export async function cancelOrder(orderNo: string, reason?: string) {
  const response = await http.post<ApiEnvelope<AdminOrderDetail>>(`/orders/${orderNo}/cancel`, { reason })
  return response.data.data
}

export async function closeOrder(orderNo: string, reason?: string) {
  const response = await http.post<ApiEnvelope<AdminOrderDetail>>(`/orders/${orderNo}/close`, { reason })
  return response.data.data
}
