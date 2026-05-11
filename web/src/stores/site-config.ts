import { defineStore } from 'pinia'

import { getSiteConfig, type SiteConfigResponse } from '../api/site-config'

interface SiteConfigState {
  siteName: string
  logoUrl: string
  loaded: boolean
}

export const useSiteConfigStore = defineStore('site-config', {
  state: (): SiteConfigState => ({
    siteName: 'PVECloud',
    logoUrl: '',
    loaded: false,
  }),
  actions: {
    async loadSiteConfig() {
      if (this.loaded) return
      try {
        const config = await getSiteConfig()
        this.applyConfig(config)
      } catch {
        this.loaded = true
      }
    },
    applyConfig(config: SiteConfigResponse) {
      this.siteName = config.site_name || 'PVECloud'
      this.logoUrl = config.logo_url || ''
      this.loaded = true
    },
  },
})
