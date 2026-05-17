import { request, type WebApiEnvelope } from './request'
import type { PageResponse } from './order'

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
  operations: InstanceOperation[]
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
