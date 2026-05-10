import { request, type WebApiEnvelope } from './request'

export interface ServerCatalogResponse {
  products: ServerCatalogProduct[]
}

export interface ServerCatalogProduct {
  product_no: string
  slug: string
  name: string
  summary: string | null
  description: string | null
  plans: ServerCatalogPlan[]
}

export interface ServerCatalogPlan {
  plan_no: string
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
  prices: ServerCatalogPlanPrice[]
  regions: ServerCatalogRegion[]
  os_templates: ServerCatalogOSTemplate[]
  network_types: ServerCatalogNetworkType[]
}

export interface ServerCatalogPlanPrice {
  billing_cycle: 'monthly' | 'quarterly' | 'semi_yearly' | 'yearly'
  price_cents: number
  original_price_cents: number | null
  currency: string
}

export interface ServerCatalogRegion {
  region_no: string
  code: string
  name: string
  country: string | null
  city: string | null
  summary: string | null
}

export interface ServerCatalogOSTemplate {
  template_no: string
  code: string
  name: string
  os_family: string
  distribution: string
  version: string
  architecture: string
  summary: string | null
}

export interface ServerCatalogNetworkType {
  network_type_no: string
  code: string
  name: string
  summary: string | null
}

export async function getServerCatalog() {
  const response = await request.get<WebApiEnvelope<ServerCatalogResponse>>('/server-catalog')
  return response.data.data
}
