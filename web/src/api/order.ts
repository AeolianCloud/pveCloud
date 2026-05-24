import { request, type WebApiEnvelope } from './request'

export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  per_page: number
  last_page: number
}

export interface OrderItem {
  order_no: string
  order_type: 'purchase' | 'renewal'
  payment_status: 'unpaid' | 'paid' | 'manual_confirmed' | 'refunded'
  status: 'pending' | 'provisioning' | 'fulfilled' | 'error' | 'cancelled' | 'closed'
  related_instance_no: string | null
  product_name: string
  plan_name: string
  billing_cycle: string
  network_type_name: string
  total_amount_cents: number
  currency: string
  created_at: string
  paid_at: string | null
  cancelled_at: string | null
  closed_at: string | null
}

export interface OrderDetail extends OrderItem {
  user_note: string | null
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
  network_type_no: string
  network_type_code: string
  network_type_name: string
  template_no: string
  template_code: string
  template_name: string
  os_family: string
  os_distribution: string
  os_version: string
  os_architecture: string
}

export interface CreateOrderPayload {
  plan_no: string
  billing_cycle: string
  region_no: string
  template_no: string
  network_type_no: string
  quantity: 1
  client_token: string
  user_note?: string | null
}

export async function createOrder(payload: CreateOrderPayload) {
  const response = await request.post<WebApiEnvelope<OrderDetail>>('/orders', payload)
  return response.data.data
}

export async function getOrders(params?: Record<string, unknown>) {
  const response = await request.get<WebApiEnvelope<PageResponse<OrderItem>>>('/orders', { params })
  return response.data.data
}

export async function getOrderDetail(orderNo: string) {
  const response = await request.get<WebApiEnvelope<OrderDetail>>(`/orders/${orderNo}`)
  return response.data.data
}

export async function cancelOrder(orderNo: string, reason?: string) {
  const response = await request.post<WebApiEnvelope<OrderDetail>>(`/orders/${orderNo}/cancel`, { reason })
  return response.data.data
}
