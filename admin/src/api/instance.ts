import { http, type ApiEnvelope } from '../utils/request'
import type { PaginatedData } from './admin-user'
import type { OrderUserSummary } from './order'

export type InstanceStatus = 'creating' | 'running' | 'stopped' | 'error' | 'releasing' | 'released'
export type InstanceOperationAction = 'provision' | 'start' | 'stop' | 'release' | 'sync'
export type InstanceOperationStatus = 'running' | 'succeeded' | 'failed'
export type MappingStatus = 'active' | 'inactive'

export interface InstanceMappingItem {
  id: number
  mapping_no: string
  product_no: string | null
  plan_no: string
  region_no: string
  template_no: string
  network_type_no: string
  node: string
  storage: string
  disk_source: string
  disk_format: string | null
  disk_interface: string | null
  snippets_storage: string | null
  ci_user: string | null
  ssh_keys: string | null
  ip_config0: string | null
  nameserver: string | null
  search_domain: string | null
  ci_packages: string | null
  apt_mirror: string | null
  vmid_start: number
  vmid_end: number
  next_vmid: number
  status: MappingStatus
  remark: string | null
  created_at: string
  updated_at: string
}

export interface InstanceMappingPayload {
  mapping_no?: string
  product_no?: string | null
  plan_no: string
  region_no: string
  template_no: string
  network_type_no?: string
  node: string
  storage: string
  disk_source: string
  disk_format?: string | null
  disk_interface?: string | null
  snippets_storage?: string | null
  ci_user?: string | null
  ssh_keys?: string | null
  ip_config0?: string | null
  nameserver?: string | null
  search_domain?: string | null
  ci_packages?: string | null
  apt_mirror?: string | null
  vmid_start: number
  vmid_end: number
  next_vmid: number
  status: MappingStatus
  remark?: string | null
}

export interface InstanceItem {
  instance_no: string
  order_no: string
  user: OrderUserSummary
  status: InstanceStatus
  product_name: string
  plan_name: string
  region_name: string
  network_type_name: string | null
  template_name: string
  external_node: string
  external_vmid: number
  created_at: string
  released_at: string | null
}

export interface InstanceOperation {
  operation_no: string
  action: InstanceOperationAction
  status: InstanceOperationStatus
  external_operation_id: string | null
  operation_location: string | null
  resource_location: string | null
  error_code: string | null
  error_message: string | null
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
  external_resource_location: string | null
  last_error_code: string | null
  last_error_message: string | null
  operations: InstanceOperation[]
}

export interface ProvisionResponse {
  instance: InstanceDetail
  operation: InstanceOperation
}

export interface PveNode {
  node: string
  name: string
  status: string
}

export interface PveVM {
  vmid: number
  name: string
  status: string
  cpus: number
  mem: number
  maxmem: number
}

export interface PveStorage {
  storage: string
  name: string
  type: string
  status: string
}

export async function getInstanceMappings(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<InstanceMappingItem>>>('/instance-provision-mappings', { params })
  return response.data.data
}

export async function createInstanceMapping(payload: InstanceMappingPayload) {
  const response = await http.post<ApiEnvelope<InstanceMappingItem>>('/instance-provision-mappings', payload)
  return response.data.data
}

export async function updateInstanceMapping(id: number, payload: InstanceMappingPayload) {
  const response = await http.patch<ApiEnvelope<InstanceMappingItem>>(`/instance-provision-mappings/${id}`, payload)
  return response.data.data
}

export async function getInstances(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<InstanceItem>>>('/instances', { params })
  return response.data.data
}

export async function getInstanceDetail(instanceNo: string) {
  const response = await http.get<ApiEnvelope<InstanceDetail>>(`/instances/${instanceNo}`)
  return response.data.data
}

export async function startInstance(instanceNo: string) {
  const response = await http.post<ApiEnvelope<InstanceDetail>>(`/instances/${instanceNo}/start`)
  return response.data.data
}

export async function stopInstance(instanceNo: string) {
  const response = await http.post<ApiEnvelope<InstanceDetail>>(`/instances/${instanceNo}/stop`)
  return response.data.data
}

export async function releaseInstance(instanceNo: string) {
  const response = await http.post<ApiEnvelope<InstanceDetail>>(`/instances/${instanceNo}/release`)
  return response.data.data
}

export async function syncInstance(instanceNo: string) {
  const response = await http.post<ApiEnvelope<InstanceDetail>>(`/instances/${instanceNo}/sync`)
  return response.data.data
}

export async function provisionOrder(orderNo: string) {
  const response = await http.post<ApiEnvelope<ProvisionResponse>>(`/orders/${orderNo}/provision`)
  return response.data.data
}

export async function getPveNodes() {
  const response = await http.get<ApiEnvelope<PveNode[]>>('/mcp-pve/nodes')
  return response.data.data
}

export async function getPveNode(node: string) {
  const response = await http.get<ApiEnvelope<PveNode>>(`/mcp-pve/nodes/${node}`)
  return response.data.data
}

export async function getPveNodeVMs(node: string) {
  const response = await http.get<ApiEnvelope<PveVM[]>>(`/mcp-pve/nodes/${node}/vms`)
  return response.data.data
}

export async function getPveStorage() {
  const response = await http.get<ApiEnvelope<PveStorage[]>>('/mcp-pve/storage')
  return response.data.data
}
