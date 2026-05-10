import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface ProductItem {
  id: number
  product_no: string
  type: string
  slug: string
  name: string
  summary: string | null
  description: string | null
  status: string
  visible: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface ProductPayload {
  product_no?: string
  type: 'server'
  slug: string
  name: string
  summary?: string | null
  description?: string | null
  status: 'draft' | 'active' | 'inactive'
  visible: boolean
  sort_order: number
}

export interface ProductPlanItem {
  id: number
  plan_no: string
  product_id: number
  code: string
  name: string
  summary: string | null
  cpu_cores: number
  memory_mb: number
  system_disk_gb: number
  data_disk_gb: number
  bandwidth_mbps: number
  traffic_gb: number | null
  public_ip_count: number
  virtualization: string
  architecture: string
  is_featured: boolean
  status: string
  visible: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface ProductPlanPayload {
  plan_no?: string
  product_id: number
  code: string
  name: string
  summary?: string | null
  cpu_cores: number
  memory_mb: number
  system_disk_gb: number
  data_disk_gb: number
  bandwidth_mbps: number
  traffic_gb?: number | null
  public_ip_count: number
  virtualization: 'kvm'
  architecture: 'x86_64'
  is_featured: boolean
  status: 'draft' | 'active' | 'inactive' | 'sold_out'
  visible: boolean
  sort_order: number
}

export interface PlanPriceItem {
  id: number
  plan_id: number
  billing_cycle: string
  price_cents: number
  original_price_cents: number | null
  currency: string
  status: string
  sort_order: number
}

export interface PlanPricePayload {
  billing_cycle: 'monthly' | 'quarterly' | 'semi_yearly' | 'yearly'
  price_cents: number
  original_price_cents?: number | null
  currency: 'CNY'
  status: 'active' | 'inactive'
  sort_order: number
}

export interface SalesRegionItem {
  id: number
  region_no: string
  code: string
  name: string
  country: string | null
  city: string | null
  summary: string | null
  status: string
  visible: boolean
  sort_order: number
}

export interface SalesRegionPayload {
  region_no?: string
  code: string
  name: string
  country?: string | null
  city?: string | null
  summary?: string | null
  status: 'active' | 'inactive'
  visible: boolean
  sort_order: number
}

export interface ServerOsTemplateItem {
  id: number
  template_no: string
  code: string
  name: string
  os_family: string
  distribution: string
  version: string
  architecture: string
  summary: string | null
  status: string
  visible: boolean
  sort_order: number
}

export interface ServerOsTemplatePayload {
  template_no?: string
  code: string
  name: string
  os_family: 'linux' | 'windows' | 'bsd'
  distribution: string
  version: string
  architecture: 'x86_64'
  summary?: string | null
  status: 'active' | 'inactive'
  visible: boolean
  sort_order: number
}

export interface NetworkTypeItem {
  id: number
  network_type_no: string
  code: string
  name: string
  summary: string | null
  status: string
  visible: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface NetworkTypePayload {
  network_type_no?: string
  code: string
  name: string
  summary?: string | null
  status: 'active' | 'inactive'
  visible: boolean
  sort_order: number
}

export async function getProducts(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<ProductItem>>>('/products', { params })
  return response.data.data
}

export async function createProduct(payload: ProductPayload) {
  const response = await http.post<ApiEnvelope<ProductItem>>('/products', payload)
  return response.data.data
}

export async function updateProduct(id: number, payload: ProductPayload) {
  const response = await http.put<ApiEnvelope<ProductItem>>(`/products/${id}`, payload)
  return response.data.data
}

export async function updateProductStatus(id: number, status: string) {
  const response = await http.patch<ApiEnvelope<ProductItem>>(`/products/${id}/status`, { status })
  return response.data.data
}

export async function deleteProduct(id: number) {
  await http.delete<ApiEnvelope<null>>(`/products/${id}`)
}

export async function getProductPlans(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<ProductPlanItem>>>('/product-plans', { params })
  return response.data.data
}

export async function createProductPlan(payload: ProductPlanPayload) {
  const response = await http.post<ApiEnvelope<ProductPlanItem>>('/product-plans', payload)
  return response.data.data
}

export async function updateProductPlan(id: number, payload: ProductPlanPayload) {
  const response = await http.put<ApiEnvelope<ProductPlanItem>>(`/product-plans/${id}`, payload)
  return response.data.data
}

export async function updateProductPlanStatus(id: number, status: string) {
  const response = await http.patch<ApiEnvelope<ProductPlanItem>>(`/product-plans/${id}/status`, { status })
  return response.data.data
}

export async function deleteProductPlan(id: number) {
  await http.delete<ApiEnvelope<null>>(`/product-plans/${id}`)
}

export async function updatePlanPrices(id: number, prices: PlanPricePayload[]) {
  const response = await http.put<ApiEnvelope<PlanPriceItem[]>>(`/product-plans/${id}/prices`, { prices })
  return response.data.data
}

export async function getPlanPrices(id: number) {
  const response = await http.get<ApiEnvelope<PlanPriceItem[]>>(`/product-plans/${id}/prices`)
  return response.data.data
}

export async function updatePlanRegions(id: number, ids: number[]) {
  const response = await http.put<ApiEnvelope<{ plan_id: number; related_ids: number[] }>>(`/product-plans/${id}/regions`, { ids })
  return response.data.data
}

export async function getPlanRegions(id: number) {
  const response = await http.get<ApiEnvelope<SalesRegionItem[]>>(`/product-plans/${id}/regions`)
  return response.data.data
}

export async function updatePlanOsTemplates(id: number, ids: number[]) {
  const response = await http.put<ApiEnvelope<{ plan_id: number; related_ids: number[] }>>(`/product-plans/${id}/os-templates`, { ids })
  return response.data.data
}

export async function updatePlanNetworkTypes(id: number, ids: number[]) {
  const response = await http.put<ApiEnvelope<{ plan_id: number; related_ids: number[] }>>(`/product-plans/${id}/network-types`, { ids })
  return response.data.data
}

export async function getPlanNetworkTypes(id: number) {
  const response = await http.get<ApiEnvelope<NetworkTypeItem[]>>(`/product-plans/${id}/network-types`)
  return response.data.data
}

export async function getPlanOsTemplates(id: number) {
  const response = await http.get<ApiEnvelope<ServerOsTemplateItem[]>>(`/product-plans/${id}/os-templates`)
  return response.data.data
}

export async function getSalesRegions(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<SalesRegionItem[]>>('/sales-regions', { params })
  return response.data.data
}

export async function createSalesRegion(payload: SalesRegionPayload) {
  const response = await http.post<ApiEnvelope<SalesRegionItem>>('/sales-regions', payload)
  return response.data.data
}

export async function updateSalesRegion(id: number, payload: SalesRegionPayload) {
  const response = await http.put<ApiEnvelope<SalesRegionItem>>(`/sales-regions/${id}`, payload)
  return response.data.data
}

export async function deleteSalesRegion(id: number) {
  await http.delete<ApiEnvelope<null>>(`/sales-regions/${id}`)
}

export async function getServerOsTemplates(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<ServerOsTemplateItem[]>>('/server-os-templates', { params })
  return response.data.data
}

export async function createServerOsTemplate(payload: ServerOsTemplatePayload) {
  const response = await http.post<ApiEnvelope<ServerOsTemplateItem>>('/server-os-templates', payload)
  return response.data.data
}

export async function updateServerOsTemplate(id: number, payload: ServerOsTemplatePayload) {
  const response = await http.put<ApiEnvelope<ServerOsTemplateItem>>(`/server-os-templates/${id}`, payload)
  return response.data.data
}

export async function deleteServerOsTemplate(id: number) {
  await http.delete<ApiEnvelope<null>>(`/server-os-templates/${id}`)
}

export async function getNetworkTypes(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<NetworkTypeItem[]>>('/network-types', { params })
  return response.data.data
}

export async function createNetworkType(payload: NetworkTypePayload) {
  const response = await http.post<ApiEnvelope<NetworkTypeItem>>('/network-types', payload)
  return response.data.data
}

export async function updateNetworkType(id: number, payload: NetworkTypePayload) {
  const response = await http.put<ApiEnvelope<NetworkTypeItem>>(`/network-types/${id}`, payload)
  return response.data.data
}

export async function deleteNetworkType(id: number) {
  await http.delete<ApiEnvelope<null>>(`/network-types/${id}`)
}
