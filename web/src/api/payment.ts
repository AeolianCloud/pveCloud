import { request } from '../lib/http'

export interface PaymentOrder {
  id: number
  payment_order_no: string
  order_id: number
  pay_status: string
  payable_amount: number
  paid_at?: string
}

export function getPaymentStatus(paymentOrderNo: string) {
  return request<PaymentOrder>(`/payments/${paymentOrderNo}`)
}
