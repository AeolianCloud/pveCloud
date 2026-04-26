import { http } from './http'
import type { ApiEnvelope } from '../types/auth'
import type { SystemConfigGroup, SystemConfigItem, SystemConfigUpdateRequest } from '../types/systemConfig'

export async function getSystemConfigs(groupName?: string): Promise<SystemConfigGroup[]> {
  const response = await http.get<ApiEnvelope<SystemConfigGroup[]>>('/system-configs', {
    params: { group_name: groupName || undefined },
  })
  return response.data.data
}

export async function updateSystemConfig(id: number, payload: SystemConfigUpdateRequest): Promise<SystemConfigItem> {
  const response = await http.patch<ApiEnvelope<SystemConfigItem>>(`/system-configs/${id}`, payload)
  return response.data.data
}
