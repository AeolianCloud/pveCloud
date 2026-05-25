import { http, type ApiEnvelope } from '../utils/request'
import type { PaginatedData } from './admin-user'
import type { OrderUserSummary } from './order'

export type InvoiceStatus = 'pending' | 'processing' | 'issued' | 'rejected' | 'cancelled'
export type InvoiceTitleType = 'personal' | 'company'

export interface InvoiceOrderItem {
  order_no: string
  order_type: 'purchase' | 'renewal'
  order_amount_cents: number
  currency: string
  payment_status: string
  paid_at: string | null
  product_name: string | null
  plan_name: string | null
}

export interface InvoiceFileSummary {
  id: number
  original_name: string
  mime_type: string
  size: number
  download_url: string
}

export interface AdminInvoiceItem {
  invoice_no: string
  invoice_type: 'electronic_normal'
  user: OrderUserSummary
  title_type: InvoiceTitleType
  title: string
  amount_cents: number
  currency: string
  status: InvoiceStatus
  order_count: number
  invoice_number: string | null
  created_at: string
  accepted_at: string | null
  issued_at: string | null
}

export interface AdminInvoiceDetail extends AdminInvoiceItem {
  tax_no: string | null
  email: string | null
  remark: string | null
  admin_note: string | null
  reject_reason: string | null
  cancel_reason: string | null
  invoice_code: string | null
  rejected_at: string | null
  cancelled_at: string | null
  orders: InvoiceOrderItem[]
  file: InvoiceFileSummary | null
}

export async function getInvoices(params?: Record<string, unknown>) {
  const response = await http.get<ApiEnvelope<PaginatedData<AdminInvoiceItem>>>('/invoices', { params })
  return response.data.data
}

export async function getInvoiceDetail(invoiceNo: string) {
  const response = await http.get<ApiEnvelope<AdminInvoiceDetail>>(`/invoices/${invoiceNo}`)
  return response.data.data
}

export async function acceptInvoice(invoiceNo: string) {
  const response = await http.post<ApiEnvelope<AdminInvoiceDetail>>(`/invoices/${invoiceNo}/accept`)
  return response.data.data
}

export async function rejectInvoice(invoiceNo: string, reason: string) {
  const response = await http.post<ApiEnvelope<AdminInvoiceDetail>>(`/invoices/${invoiceNo}/reject`, { reason })
  return response.data.data
}

export async function issueInvoice(invoiceNo: string, payload: { invoice_code?: string | null; invoice_number: string; issued_at: string; file_id: number }) {
  const response = await http.post<ApiEnvelope<AdminInvoiceDetail>>(`/invoices/${invoiceNo}/issue`, payload)
  return response.data.data
}

export async function updateInvoiceAdminNote(invoiceNo: string, adminNote: string | null) {
  const response = await http.patch<ApiEnvelope<AdminInvoiceDetail>>(`/invoices/${invoiceNo}/admin-note`, { admin_note: adminNote })
  return response.data.data
}

export async function downloadInvoice(invoiceNo: string) {
  const response = await http.get<Blob>(`/invoices/${invoiceNo}/download`, { responseType: 'blob' })
  return response.data
}
