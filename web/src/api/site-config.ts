import { request, type WebApiEnvelope } from './request'

export interface SiteConfigResponse {
  site_name: string
  logo_url: string
  login_captcha_enabled: boolean
  register_captcha_enabled: boolean
  password_reset_request_captcha_enabled: boolean
  password_reset_confirm_captcha_enabled: boolean
}

export async function getSiteConfig() {
  const response = await request.get<WebApiEnvelope<SiteConfigResponse>>('/site-config')
  return response.data.data
}
