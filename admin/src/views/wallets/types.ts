export interface WalletQueryState {
  page: number
  per_page: number
  wallet_no: string
  user_keyword: string
  status: string
}

export interface WalletLedgerQueryState {
  page: number
  per_page: number
  wallet_no: string
  user_keyword: string
  direction: string
  entry_type: string
  related_no: string
}

export interface WalletRechargeQueryState {
  page: number
  per_page: number
  wallet_no: string
  user_keyword: string
  provider: string
  method: string
  status: string
  recharge_no: string
}
