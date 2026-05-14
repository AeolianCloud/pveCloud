import { request, type WebApiEnvelope } from './request'
import type { PageResponse } from './order'

export type TicketStatus = 'waiting_admin' | 'waiting_user' | 'closed'
export type TicketCategory = 'account' | 'order' | 'product' | 'technical' | 'billing' | 'other'
export type TicketPriority = 'low' | 'normal' | 'high' | 'urgent'
export type TicketSenderType = 'user' | 'admin'

export interface TicketTagItem {
  id: number
  name: string
  color: string | null
  visibility: 'public'
}

export interface TicketAttachment {
  file_id: number
  original_name: string
  mime_type: string
  extension: string
  size: number
  download_url: string
}

export interface TicketMessage {
  id: number
  sender_type: TicketSenderType
  sender_name: string
  content: string
  attachments: TicketAttachment[]
  created_at: string
}

export interface TicketItem {
  ticket_no: string
  title: string
  category: TicketCategory
  priority: TicketPriority
  status: TicketStatus
  tags: TicketTagItem[]
  order_no: string | null
  last_message_at: string
  created_at: string
  closed_at: string | null
}

export interface TicketDetail extends TicketItem {
  close_reason: string | null
  messages: TicketMessage[]
}

export interface TicketListQuery {
  page?: number
  per_page?: number
  status?: string
  category?: string
  priority?: string
  order_no?: string
}

export interface CreateTicketPayload {
  title: string
  category: string
  priority?: string
  content: string
  order_no?: string
  files?: File[]
}

export async function getTickets(params?: TicketListQuery) {
  const response = await request.get<WebApiEnvelope<PageResponse<TicketItem>>>('/tickets', { params })
  return response.data.data
}

export async function createTicket(payload: CreateTicketPayload) {
  const formData = new FormData()
  formData.append('title', payload.title)
  formData.append('category', payload.category)
  formData.append('priority', payload.priority || 'normal')
  formData.append('content', payload.content)
  if (payload.order_no) formData.append('order_no', payload.order_no)
  payload.files?.forEach((file) => formData.append('attachments', file))
  const response = await request.post<WebApiEnvelope<TicketDetail>>('/tickets', formData)
  return response.data.data
}

export async function getTicketDetail(ticketNo: string) {
  const response = await request.get<WebApiEnvelope<TicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}`)
  return response.data.data
}

export async function replyTicket(ticketNo: string, content: string, files: File[]) {
  const formData = new FormData()
  formData.append('content', content)
  files.forEach((file) => formData.append('attachments', file))
  const response = await request.post<WebApiEnvelope<TicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/messages`, formData)
  return response.data.data
}

export async function closeTicket(ticketNo: string, reason?: string) {
  const response = await request.post<WebApiEnvelope<TicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/close`, { reason })
  return response.data.data
}

export async function downloadTicketAttachment(ticketNo: string, fileId: number) {
  const response = await request.get<Blob>(`/tickets/${encodeURIComponent(ticketNo)}/attachments/${fileId}/download`, {
    responseType: 'blob',
  })
  return response.data
}
