import { http, type ApiEnvelope } from '../utils/request'
import type { PaginatedData } from './admin-user'

export type TicketStatus = 'waiting_admin' | 'waiting_user' | 'closed'
export type TicketCategory = 'account' | 'order' | 'product' | 'technical' | 'billing' | 'other'
export type TicketPriority = 'low' | 'normal' | 'high' | 'urgent'
export type TicketSenderType = 'user' | 'admin'
export type TicketTagVisibility = 'public' | 'internal'
export type TicketTagStatus = 'active' | 'disabled'
export type TicketSlaStatus = 'normal' | 'first_response_overdue' | 'resolution_overdue'

export interface TicketUserSummary {
  id: number
  username: string
  email: string
  display_name: string | null
}

export interface TicketAdminSummary {
  id: number
  username: string
  email: string | null
  display_name: string
  status?: string
}

export interface TicketTagItem {
  id: number
  name: string
  color: string | null
  visibility: TicketTagVisibility
  status: TicketTagStatus
  sort_order: number
  created_at: string
  updated_at: string
}

export interface TicketSlaInfo {
  first_response_due_at: string | null
  first_responded_at: string | null
  resolution_due_at: string | null
  resolved_at: string | null
  status: TicketSlaStatus
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
  assignee: TicketAdminSummary | null
  tags: TicketTagItem[]
  sla: TicketSlaInfo
  order_no: string | null
  instance_no: string | null
  last_message_at: string
  created_at: string
  closed_at: string | null
}

export interface AdminTicketDetail extends AdminTicketItem {
  close_reason: string | null
  messages: AdminTicketMessage[]
  collaborators: TicketAdminSummary[]
  internal_notes: TicketInternalNote[]
  events: TicketEvent[]
}

export interface TicketInternalNote {
  id: number
  admin: TicketAdminSummary
  content: string
  created_at: string
}

export interface TicketEvent {
  id: number
  event_type: string
  actor: { type: string; id: number; username: string; display_name: string | null } | null
  before_data: string | null
  after_data: string | null
  remark: string | null
  created_at: string
}

export interface AdminTicketListQuery {
  page?: number
  per_page?: number
  status?: string
  category?: string
  priority?: string
  ticket_no?: string
  order_no?: string
  instance_no?: string
  user_keyword?: string
  date_from?: string
  date_to?: string
  assignee_admin_id?: number | string
  tag_id?: number | string
  sla_status?: string
}

export interface TicketTagListQuery {
  page?: number
  per_page?: number
  keyword?: string
  visibility?: string
  status?: string
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

export async function getAssigneeCandidates(params?: { page?: number; per_page?: number; keyword?: string }) {
  const response = await http.get<ApiEnvelope<PaginatedData<TicketAdminSummary>>>('/tickets/assignee-candidates', { params })
  return response.data.data
}

export async function assignTicket(ticketNo: string, assigneeAdminId: number, reason?: string) {
  const response = await http.post<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/assign`, { assignee_admin_id: assigneeAdminId, reason })
  return response.data.data
}

export async function addTicketCollaborator(ticketNo: string, adminId: number) {
  const response = await http.post<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/collaborators`, { admin_id: adminId })
  return response.data.data
}

export async function removeTicketCollaborator(ticketNo: string, adminId: number) {
  const response = await http.delete<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/collaborators/${adminId}`)
  return response.data.data
}

export async function addTicketInternalNote(ticketNo: string, content: string) {
  const response = await http.post<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/internal-notes`, { content })
  return response.data.data
}

export async function upgradeTicketPriority(ticketNo: string, priority: TicketPriority, reason: string) {
  const response = await http.post<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/priority`, { priority, reason })
  return response.data.data
}

export async function replaceTicketTags(ticketNo: string, tagIds: number[]) {
  const response = await http.put<ApiEnvelope<AdminTicketDetail>>(`/tickets/${encodeURIComponent(ticketNo)}/tags`, { tag_ids: tagIds })
  return response.data.data
}

export async function getTicketTags(params?: TicketTagListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<TicketTagItem>>>('/ticket-tags', { params })
  return response.data.data
}

export async function createTicketTag(payload: { name: string; color?: string | null; visibility: TicketTagVisibility; status: TicketTagStatus; sort_order?: number }) {
  const response = await http.post<ApiEnvelope<TicketTagItem>>('/ticket-tags', payload)
  return response.data.data
}

export async function updateTicketTag(id: number, payload: Partial<{ name: string; color: string | null; visibility: TicketTagVisibility; status: TicketTagStatus; sort_order: number }>) {
  const response = await http.patch<ApiEnvelope<TicketTagItem>>(`/ticket-tags/${id}`, payload)
  return response.data.data
}

export async function downloadTicketAttachment(ticketNo: string, fileId: number) {
  const response = await http.get<Blob>(`/tickets/${encodeURIComponent(ticketNo)}/attachments/${fileId}/download`, {
    responseType: 'blob',
  })
  return response.data
}
