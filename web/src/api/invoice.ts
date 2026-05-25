import { request, type WebApiEnvelope } from './request'
import type { PageResponse } from './order'

export type InvoiceStatus = 'pending' | 'processing' | 'issued' | 'rejected' | 'cancelled'
export type InvoiceTitleType = 'personal' | 'company'
export type InvoiceType = 'electronic_normal'

export interface InvoiceEligibleOrderItem {
  order_no: string
  order_type: 'purchase' | 'renewal'
  related_instance_no: string | null
  amount_cents: number
  currency: string
  payment_status: string
  paid_at: string | null
  product_name: string
  plan_name: string
  invoice_occupied: boolean
}

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

export interface InvoiceItem {
  invoice_no: string
  invoice_type: InvoiceType
  title_type: InvoiceTitleType
  title: string
  amount_cents: number
  currency: string
  status: InvoiceStatus
  order_count: number
  invoice_number: string | null
  issued_at: string | null
  created_at: string
  can_cancel: boolean
  can_download: boolean
  download_url: string | null
}

export interface InvoiceDetail extends InvoiceItem {
  tax_no: string | null
  email: string | null
  remark: string | null
  reject_reason: string | null
  cancel_reason: string | null
  invoice_code: string | null
  accepted_at: string | null
  rejected_at: string | null
  cancelled_at: string | null
  orders: InvoiceOrderItem[]
}

export interface CreateInvoicePayload {
  order_nos: string[]
  title_type: InvoiceTitleType
  title: string
  tax_no?: string | null
  email?: string | null
  remark?: string | null
  client_token: string
}

export interface InvoiceListQuery {
  page?: number
  per_page?: number
  status?: string
  date_from?: string
  date_to?: string
}

export interface InvoiceEligibleOrderQuery {
  page?: number
  per_page?: number
  keyword?: string
  date_from?: string
  date_to?: string
}

export async function getInvoiceEligibleOrders(params?: InvoiceEligibleOrderQuery) {
  const response = await request.get<WebApiEnvelope<PageResponse<InvoiceEligibleOrderItem>>>('/invoice-eligible-orders', { params })
  return response.data.data
}

export async function createInvoice(payload: CreateInvoicePayload) {
  const response = await request.post<WebApiEnvelope<InvoiceDetail>>('/invoices', payload)
  return response.data.data
}

export async function getInvoices(params?: InvoiceListQuery) {
  const response = await request.get<WebApiEnvelope<PageResponse<InvoiceItem>>>('/invoices', { params })
  return response.data.data
}

export async function getInvoiceDetail(invoiceNo: string) {
  const response = await request.get<WebApiEnvelope<InvoiceDetail>>(`/invoices/${encodeURIComponent(invoiceNo)}`)
  return response.data.data
}

export async function cancelInvoice(invoiceNo: string, reason?: string) {
  const response = await request.post<WebApiEnvelope<InvoiceDetail>>(`/invoices/${encodeURIComponent(invoiceNo)}/cancel`, { reason })
  return response.data.data
}

export async function downloadInvoice(invoiceNo: string) {
  const response = await request.get<Blob>(`/invoices/${encodeURIComponent(invoiceNo)}/download`, {
    responseType: 'blob',
  })
  return response.data
}
