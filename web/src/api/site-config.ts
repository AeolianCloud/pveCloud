import { request, type WebApiEnvelope } from './request'

export interface SiteConfigResponse {
  site_name: string
  logo_url: string
}

export async function getSiteConfig() {
  const response = await request.get<WebApiEnvelope<SiteConfigResponse>>('/site-config')
  return response.data.data
}
