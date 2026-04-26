import { http } from './http'
import type { ApiEnvelope } from '../types/auth'
import type { AuditLogItem, AuditLogQuery, PageResponse, RiskLogItem, RiskLogQuery } from '../types/audit'

export async function getAuditLogs(params: AuditLogQuery): Promise<PageResponse<AuditLogItem>> {
  const response = await http.get<ApiEnvelope<PageResponse<AuditLogItem>>>('/audit-logs', { params })
  return response.data.data
}

export async function getRiskLogs(params: RiskLogQuery): Promise<PageResponse<RiskLogItem>> {
  const response = await http.get<ApiEnvelope<PageResponse<RiskLogItem>>>('/risk-logs', { params })
  return response.data.data
}
