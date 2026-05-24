import { http, type ApiEnvelope } from '../utils/request'
import type { PaginatedData } from './admin-user'
import type { OrderUserSummary } from './order'

export type PaymentProvider = 'alipay' | 'wechat'
export type PaymentMethod = 'alipay_page' | 'alipay_wap' | 'wechat_native' | 'wechat_h5'
export type PaymentStatus = 'pending' | 'paid' | 'closed' | 'failed' | 'refunded'
export type RefundStatus = 'pending' | 'succeeded' | 'failed'

export interface AdminPaymentItem {
  payment_no: string
  order_no: string
  user: OrderUserSummary
  provider: PaymentProvider
  method: PaymentMethod
  status: PaymentStatus
  amount_cents: number
  currency: string
  expires_at: string
  paid_at: string | null
  created_at: string
  order_status: string
  order_type: 'purchase' | 'renewal'
}

export interface AdminPaymentDetail extends AdminPaymentItem {
  upstream_trade_no: string | null
  last_error_message: string | null
  refund: AdminRefundItem | null
}

export interface AdminRefundItem {
  refund_no: string
  payment_no: string
  order_no: string
  user: OrderUserSummary
  provider: PaymentProvider
  status: RefundStatus
  amount_cents: number
  currency: string
  reason: string
  created_at: string
  completed_at: string | null
  failed_at: string | null
}

export async function getPayments(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<AdminPaymentItem>>>('/payments', { params })
  return response.data.data
}

export async function getPaymentDetail(paymentNo: string) {
  const response = await http.get<ApiEnvelope<AdminPaymentDetail>>(`/payments/${paymentNo}`)
  return response.data.data
}

export async function syncPayment(paymentNo: string) {
  const response = await http.post<ApiEnvelope<AdminPaymentDetail>>(`/payments/${paymentNo}/sync`)
  return response.data.data
}

export async function createPaymentRefund(paymentNo: string, reason: string) {
  const response = await http.post<ApiEnvelope<AdminRefundItem>>(`/payments/${paymentNo}/refunds`, { reason })
  return response.data.data
}

export async function retryPaymentProvision(paymentNo: string) {
  const response = await http.post<ApiEnvelope<AdminPaymentDetail>>(`/payments/${paymentNo}/retry-provision`)
  return response.data.data
}

export async function getRefunds(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<AdminRefundItem>>>('/refunds', { params })
  return response.data.data
}
