import { request } from '../lib/http'

export interface Order {
  id: number
  order_no: string
  user_id: number
  sku_id: number
  region_id: number
  status: string
  cycle: string
  payable_amount: number
}

export function listOrders() {
  return request<Order[]>('/orders')
}
