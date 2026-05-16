import { http, type ApiEnvelope } from '../utils/request'

import type { PaginatedData } from './admin-user'

export interface LogUserSummary {
  id: number
  username: string
  email: string | null
  display_name: string | null
}

export interface LogListQuery {
  page?: number
  per_page?: number
  user_id?: number
  username?: string
  action?: string
  result?: string
  module?: string
  object_type?: string
  object_id?: string
  source_app?: string
  page_path?: string
  error_type?: string
  api_path?: string
  http_status?: number
  level?: string
  category?: string
  status?: number
  request_id?: string
  request_path?: string
  ip?: string
  date_from?: string
  date_to?: string
}

export interface UserSecurityLogItem {
  id: number
  user: LogUserSummary | null
  session_id: string | null
  request_id: string | null
  request_method: string | null
  request_path: string | null
  action: string
  result: string
  ip: string | null
  user_agent: string | null
  remark: string | null
  created_at: string
}

export interface UserBusinessLogItem {
  id: number
  user: LogUserSummary
  request_id: string | null
  request_method: string | null
  request_path: string | null
  module: string
  action: string
  object_type: string
  object_id: string | null
  summary: string | null
  ip: string | null
  user_agent: string | null
  created_at: string
}

export interface FrontendErrorLogItem {
  id: number
  source_app: string
  user_id: number | null
  admin_id: number | null
  request_id: string | null
  page_path: string
  error_type: string
  message: string
  stack: string | null
  api_path: string | null
  http_status: number | null
  business_code: number | null
  browser: string | null
  os: string | null
  app_version: string | null
  ip: string | null
  user_agent: string | null
  created_at: string
}

export interface BackendRuntimeLogItem {
  id: number
  level: string
  category: string
  request_id: string | null
  request_method: string | null
  request_path: string | null
  status: number | null
  latency_ms: number | null
  client_ip: string | null
  message: string
  detail: string | null
  created_at: string
}

export interface ClientErrorLogPayload {
  request_id?: string
  page_path: string
  error_type: string
  message: string
  stack?: string
  api_path?: string
  http_status?: number
  business_code?: number
  browser?: string
  os?: string
  app_version?: string
}

export async function getUserSecurityLogs(query?: LogListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<UserSecurityLogItem>>>('/logs/user-security', { params: query })
  return response.data.data
}

export async function getUserBusinessLogs(query?: LogListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<UserBusinessLogItem>>>('/logs/user-business', { params: query })
  return response.data.data
}

export async function getFrontendErrorLogs(query?: LogListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<FrontendErrorLogItem>>>('/logs/frontend-errors', { params: query })
  return response.data.data
}

export async function getBackendRuntimeLogs(query?: LogListQuery) {
  const response = await http.get<ApiEnvelope<PaginatedData<BackendRuntimeLogItem>>>('/logs/backend-runtime', { params: query })
  return response.data.data
}

export async function reportAdminClientError(payload: ClientErrorLogPayload) {
  await http.post<ApiEnvelope<Record<string, never>>>('/client-logs/errors', payload)
}
