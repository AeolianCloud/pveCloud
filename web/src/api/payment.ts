import { request, type WebApiEnvelope } from './request'

export type PaymentProvider = 'alipay' | 'wechat' | 'wallet'
export type PaymentMethod = 'alipay_page' | 'alipay_wap' | 'wechat_native' | 'wechat_h5' | 'wallet_balance'
export type PaymentStatusValue = 'pending' | 'paid' | 'closed' | 'failed' | 'refunded'

export interface PaymentStatus {
  payment_no: string
  order_no: string
  provider: PaymentProvider
  method: PaymentMethod
  amount_cents: number
  currency: string
  status: PaymentStatusValue
  expires_at: string
  redirect_url: string | null
  qr_code_url: string | null
  paid_at: string | null
  order_status: string
  order_payment_status: string
  related_instance_no: string | null
  last_error_message: string | null
}

export interface CreatePaymentPayload {
  provider: PaymentProvider
  method: PaymentMethod
  client_token: string
}

export async function createPayment(orderNo: string, payload: CreatePaymentPayload) {
  const response = await request.post<WebApiEnvelope<PaymentStatus>>(`/orders/${orderNo}/payments`, payload)
  return response.data.data
}

export async function getPayment(paymentNo: string) {
  const response = await request.get<WebApiEnvelope<PaymentStatus>>(`/payments/${paymentNo}`)
  return response.data.data
}
