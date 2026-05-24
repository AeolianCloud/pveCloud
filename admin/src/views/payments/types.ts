export interface PaymentQueryState {
  page: number
  per_page: number
  provider: string
  method: string
  status: string
  order_no: string
  payment_no: string
  user_keyword: string
}

export interface RefundQueryState {
  page: number
  per_page: number
  provider: string
  status: string
  order_no: string
  payment_no: string
  refund_no: string
}
