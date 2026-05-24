import { request, type WebApiEnvelope } from './request'

export interface WalletSummary {
  wallet_no: string
  currency: string
  status: 'active' | 'disabled'
  available_balance_cents: number
  total_recharged_cents: number
  total_spent_cents: number
  total_refunded_cents: number
  created_at: string
}

export interface WalletLedgerItem {
  entry_no: string
  wallet_no: string
  direction: 'credit' | 'debit'
  entry_type: 'recharge' | 'payment' | 'refund'
  amount_cents: number
  balance_before_cents: number
  balance_after_cents: number
  currency: string
  related_type: string
  related_no: string
  summary: string | null
  created_at: string
}

export interface WalletRechargeStatus {
  recharge_no: string
  wallet_no: string
  provider: 'alipay' | 'wechat'
  method: 'alipay_page' | 'alipay_wap' | 'wechat_native' | 'wechat_h5'
  status: 'pending' | 'paid' | 'closed' | 'failed'
  amount_cents: number
  currency: string
  expires_at: string
  paid_at: string | null
  closed_at: string | null
  failed_at: string | null
  redirect_url: string | null
  qr_code_url: string | null
  last_error_message: string | null
  created_at: string
}

export interface PageData<T> {
  list: T[]
  total: number
  page: number
  per_page: number
  last_page: number
}

export interface CreateWalletRechargePayload {
  provider: 'alipay' | 'wechat'
  method: 'alipay_page' | 'alipay_wap' | 'wechat_native' | 'wechat_h5'
  amount_cents: number
  client_token: string
}

export async function getWallet() {
  const response = await request.get<WebApiEnvelope<WalletSummary>>('/wallet')
  return response.data.data
}

export async function getWalletLedger(params?: Record<string, unknown>) {
  const response = await request.get<WebApiEnvelope<PageData<WalletLedgerItem>>>('/wallet/ledger', { params })
  return response.data.data
}

export async function createWalletRecharge(payload: CreateWalletRechargePayload) {
  const response = await request.post<WebApiEnvelope<WalletRechargeStatus>>('/wallet/recharges', payload)
  return response.data.data
}

export async function getWalletRecharge(rechargeNo: string) {
  const response = await request.get<WebApiEnvelope<WalletRechargeStatus>>(`/wallet/recharges/${rechargeNo}`)
  return response.data.data
}
