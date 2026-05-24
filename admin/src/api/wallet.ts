import { http, type ApiEnvelope } from '../utils/request'
import type { PaginatedData } from './admin-user'

export interface WalletUserSummary {
  id: number
  username: string
  email: string
  display_name: string | null
}

export interface WalletItem {
  wallet_no: string
  user: WalletUserSummary
  currency: string
  status: 'active' | 'disabled'
  available_balance_cents: number
  total_recharged_cents: number
  total_spent_cents: number
  total_refunded_cents: number
  created_at: string
  updated_at: string
}

export interface WalletLedgerItem {
  entry_no: string
  wallet_no: string
  user: WalletUserSummary
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

export interface WalletRechargeItem {
  recharge_no: string
  wallet_no: string
  user: WalletUserSummary
  provider: 'alipay' | 'wechat'
  method: 'alipay_page' | 'alipay_wap' | 'wechat_native' | 'wechat_h5'
  status: 'pending' | 'paid' | 'closed' | 'failed'
  amount_cents: number
  currency: string
  expires_at: string
  paid_at: string | null
  closed_at: string | null
  failed_at: string | null
  created_at: string
}

export interface WalletDetail extends WalletItem {
  recent_ledger: WalletLedgerItem[]
  recent_recharges: WalletRechargeItem[]
}

export async function getWallets(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<WalletItem>>>('/wallets', { params })
  return response.data.data
}

export async function getWalletDetail(walletNo: string) {
  const response = await http.get<ApiEnvelope<WalletDetail>>(`/wallets/${walletNo}`)
  return response.data.data
}

export async function getWalletLedger(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<WalletLedgerItem>>>('/wallet-ledger', { params })
  return response.data.data
}

export async function getWalletRecharges(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<WalletRechargeItem>>>('/wallet-recharges', { params })
  return response.data.data
}
