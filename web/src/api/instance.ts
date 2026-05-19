import { request, type WebApiEnvelope } from './request'
import type { OrderDetail, PageResponse } from './order'

export type InstanceStatus = 'creating' | 'running' | 'stopped' | 'error' | 'releasing' | 'released'
export type InstanceOperationAction = 'provision' | 'start' | 'stop' | 'release' | 'sync'
export type InstanceOperationStatus = 'running' | 'succeeded' | 'failed'

export interface InstanceItem {
  instance_no: string
  order_no: string
  status: InstanceStatus
  product_name: string
  plan_name: string
  region_name: string
  network_type_name: string | null
  template_name: string
  service_started_at: string | null
  expires_at: string | null
  expire_status: 'active' | 'expired' | 'released' | 'unknown'
  release_countdown_seconds: number | null
  latest_renewal_order: RenewalOrderSummary | null
  created_at: string
  released_at: string | null
}

export interface InstanceOperation {
  operation_no: string
  action: InstanceOperationAction
  status: InstanceOperationStatus
  created_at: string
  completed_at: string | null
}

export interface InstanceDetail extends InstanceItem {
  product_no: string
  plan_no: string
  cpu_cores: number
  memory_mb: number
  system_disk_gb: number
  data_disk_gb: number
  bandwidth_mbps: number
  region_no: string
  network_type_no: string | null
  template_no: string
  os_family: string
  os_distribution: string
  os_version: string
  expire_notice_sent_at: string | null
  expire_release_scheduled_at: string | null
  expire_released_at: string | null
  renewal_available: boolean
  operations: InstanceOperation[]
}

export interface RenewalOrderSummary {
  order_no: string
  status: string
  payment_status: string
  billing_cycle: string
  total_amount_cents: number
  currency: string
  paid_at: string | null
  created_at: string
}

export async function getInstances(params?: Record<string, unknown>) {
  const response = await request.get<WebApiEnvelope<PageResponse<InstanceItem>>>('/instances', { params })
  return response.data.data
}

export async function getInstanceDetail(instanceNo: string) {
  const response = await request.get<WebApiEnvelope<InstanceDetail>>(`/instances/${instanceNo}`)
  return response.data.data
}

export async function startInstance(instanceNo: string) {
  const response = await request.post<WebApiEnvelope<InstanceDetail>>(`/instances/${instanceNo}/start`)
  return response.data.data
}

export async function stopInstance(instanceNo: string) {
  const response = await request.post<WebApiEnvelope<InstanceDetail>>(`/instances/${instanceNo}/stop`)
  return response.data.data
}

export async function createRenewalOrder(instanceNo: string, billingCycle: string, clientToken: string) {
  const response = await request.post<WebApiEnvelope<OrderDetail>>(`/instances/${instanceNo}/renewal-orders`, {
    billing_cycle: billingCycle,
    client_token: clientToken,
  })
  return response.data.data
}
