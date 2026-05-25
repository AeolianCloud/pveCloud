export interface InvoiceQueryState {
  page: number
  per_page: number
  status: string
  invoice_no: string
  order_no: string
  user_keyword: string
  title_keyword: string
}

export interface RejectFormState {
  invoice_no: string
  reason: string
}

export interface IssueFormState {
  invoice_no: string
  invoice_code: string
  invoice_number: string
  issued_at: number | null
  file_id: number | null
  file_name: string
}
