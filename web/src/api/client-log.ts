import { request, type WebApiEnvelope } from './request'

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

export async function reportWebClientError(payload: ClientErrorLogPayload) {
  await request.post<WebApiEnvelope<Record<string, never>>>('/client-logs/errors', payload)
}
