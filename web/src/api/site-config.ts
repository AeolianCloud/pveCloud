import { request, type WebApiEnvelope } from './request'

export interface SiteConfigResponse {
  site_name: string
  logo_url: string
  login_captcha_enabled: boolean
  register_captcha_enabled: boolean
  password_reset_request_captcha_enabled: boolean
  password_reset_confirm_captcha_enabled: boolean
  real_name: RealNameConfig
}

export interface RealNameConfig {
  enabled: boolean
  required_for_order: boolean
  allowed_providers: string[]
  default_provider: string
  resubmit_enabled: boolean
  max_submit_attempts: number
  review_notice: string
}

export async function getSiteConfig() {
  const response = await request.get<WebApiEnvelope<SiteConfigResponse>>('/site-config')
  return response.data.data
}
