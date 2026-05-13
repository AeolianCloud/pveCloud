import { http, type ApiEnvelope } from '../utils/request'
import type { PaginatedData } from './admin-user'

export type TicketStatus = 'waiting_admin' | 'waiting_user' | 'closed'
export type TicketCategory = 'account' | 'order' | 'product' | 'technical' | 'billing' | 'other'
export type TicketPriority = 'low' | 'normal' | 'high' | 'urgent'
export type TicketSenderType = 'user' | 'admin'

export interface TicketUserSummary {
  id: number
  username: string
  email: string
  display_name: string | null
}

export interface AdminTicketAttachment {
  file_id: number
  original_name: string
  mime_type: string
  extension: string
  size: number
  download_url: string
}

export interface AdminTicketMessage {
  id: number
  sender_type: TicketSenderType
  sender_name: string
  content: string
  attachments: AdminTicketAttachment[]
  created_at: string
}

export interface AdminTicketItem {
  ticket_no: string
  user: TicketUserSummary
  title: string
  category: TicketCategory
  priority: TicketPriority
  status: TicketStatus
  order_no: string | null
  last_message_at: string
  created_at: string
  closed_at: string | null
}

export interface AdminTicketDetail extends AdminTicketItem {
  close_reason: string | null
  messages: AdminTicketMessage[]
}

export interface AdminTicketListQuery {
  page?: number
  per_page?: number
  status?: string
  category?: string
  priority?: string
  ticket_no?: string
  order_no?: string
  user_keyword?: string
  date_from?: string
  date_to?: string
}

export async function getTickets(params?: AdminTicketListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<AdminTicketItem>>>('/tickets', { params })
  return response.data.data
}

export async function getTicketDetail(ticketNo: string) {
  const response = await http.get<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}`)
  return response.data.data
}

export async function replyTicket(ticketNo: string, content: string, files: File[]) {
  const formData = new FormData()
  formData.append('content', content)
  files.forEach((file) => formData.append('attachments', file))
  const response = await http.post<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/messages`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return response.data.data
}

export async function closeTicket(ticketNo: string, reason?: string) {
  const response = await http.post<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/close`, { reason })
  return response.data.data
}

export async function downloadTicketAttachment(ticketNo: string, fileId: number) {
  const response = await http.get<Blob>(`/tickets/${encodeURIComponent(ticketNo)}/attachments/${fileId}/download`, {
    responseType: 'blob',
  })
  return response.data
}
