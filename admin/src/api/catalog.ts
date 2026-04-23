import { request } from '../lib/http'

export interface Product {
  id: number
  product_no: string
  product_name: string
  product_type: string
  status: string
}

export interface SKU {
  id: number
  sku_no: string
  product_id: number
  sku_name: string
  cpu_cores: number
  memory_mb: number
  disk_gb: number
  bandwidth_mbps: number
  status: string
}

export interface SaleableProduct {
  product: Product
  skus: SKU[]
}

export function listProducts() {
  return request<SaleableProduct[]>('/products')
}
