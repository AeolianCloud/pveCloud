import { request } from '../lib/http'

export interface Order {
  id: number
  order_no: string
  user_id: number
  sku_id: number
  region_id: number
  reservation_id: number
  status: string
  cycle: string
  original_amount: number
  discount_amount: number
  payable_amount: number
}

export interface PaymentOrder {
  id: number
  payment_order_no: string
  order_id: number
  pay_status: string
  payable_amount: number
  paid_at?: string
}

export interface CreateOrderPayload {
  order: Order
  payment_order: PaymentOrder
}

export function createOrder(skuID: number, regionID: number, cycle: string) {
  return request<CreateOrderPayload>('/orders', {
    method: 'POST',
    bodyJson: {
      sku_id: skuID,
      region_id: regionID,
      cycle,
    },
  })
}

export function listOrders() {
  return request<Order[]>('/orders')
}
