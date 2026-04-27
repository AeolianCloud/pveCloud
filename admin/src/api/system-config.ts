import { http, type ApiEnvelope } from '../utils/request'

export interface SystemConfigItem {
  id: number
  config_key: string
  config_value: string | null
  value_type: string
  group_name: string
  is_secret: boolean
  has_value: boolean
  description: string | null
  updated_at: string
}

export interface SystemConfigGroup {
  group_name: string
  items: SystemConfigItem[]
}

export interface UpdateSystemConfigRequest {
  config_value: string
}

export async function getSystemConfigs(groupName?: string) {
  const params: Record<string, string> = {}
  if (groupName) {
    params.group_name = groupName
  }
  const response = await http.get<ApiEnvelope<SystemConfigGroup[]>>('/system-configs', { params })
  return response.data.data
}

export async function updateSystemConfig(id: number, payload: UpdateSystemConfigRequest) {
  const response = await http.patch<ApiEnvelope<{ id: number }>>(`/system-configs/${id}`, payload)
  return response.data.data
}
