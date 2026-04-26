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

export interface SystemConfigUpdateRequest {
  config_value: string
}
