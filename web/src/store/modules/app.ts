import { defineStore } from 'pinia'

import { getSiteConfig } from '../../api/site-config'

type Theme = 'dark' | 'light'

function getInitialTheme(): Theme {
  if (typeof window === 'undefined') return 'dark'
  const saved = localStorage.getItem('pve-theme')
  if (saved === 'light' || saved === 'dark') return saved
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
}

function applyTheme(theme: Theme) {
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem('pve-theme', theme)
}

const initialTheme = getInitialTheme()
applyTheme(initialTheme)

export const useWebAppStore = defineStore('web-app', {
  state: () => ({
    navigationOpen: false,
    theme: initialTheme,
    siteName: 'pveCloud',
    logoUrl: '',
  }),
  actions: {
    async loadSiteConfig() {
      try {
        const config = await getSiteConfig()
        this.siteName = config.site_name.trim() || 'pveCloud'
        this.logoUrl = config.logo_url.trim()
      } catch {
        this.siteName = 'pveCloud'
        this.logoUrl = ''
      }
    },
    openNavigation() {
      this.navigationOpen = true
    },
    closeNavigation() {
      this.navigationOpen = false
    },
    toggleNavigation() {
      this.navigationOpen = !this.navigationOpen
    },
    toggleTheme() {
      this.theme = this.theme === 'dark' ? 'light' : 'dark'
      applyTheme(this.theme)
    },
  },
})
