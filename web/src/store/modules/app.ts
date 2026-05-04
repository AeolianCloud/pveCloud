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
let siteConfigPromise: Promise<void> | undefined

export const useWebAppStore = defineStore('web-app', {
  state: () => ({
    navigationOpen: false,
    theme: initialTheme,
    siteName: 'pveCloud',
    logoUrl: '',
    siteConfigLoaded: false,
    loginCaptchaEnabled: false,
    registerCaptchaEnabled: false,
    passwordResetRequestCaptchaEnabled: false,
    passwordResetConfirmCaptchaEnabled: false,
  }),
  actions: {
    async loadSiteConfig(force = false) {
      if (this.siteConfigLoaded && !force) return
      if (siteConfigPromise && !force) return siteConfigPromise

      siteConfigPromise = (async () => {
        try {
          const config = await getSiteConfig()
          this.siteName = config.site_name.trim() || 'pveCloud'
          this.logoUrl = config.logo_url.trim()
          this.loginCaptchaEnabled = config.login_captcha_enabled
          this.registerCaptchaEnabled = config.register_captcha_enabled
          this.passwordResetRequestCaptchaEnabled = config.password_reset_request_captcha_enabled
          this.passwordResetConfirmCaptchaEnabled = config.password_reset_confirm_captcha_enabled
        } catch {
          this.siteName = 'pveCloud'
          this.logoUrl = ''
          this.loginCaptchaEnabled = false
          this.registerCaptchaEnabled = false
          this.passwordResetRequestCaptchaEnabled = false
          this.passwordResetConfirmCaptchaEnabled = false
        } finally {
          this.siteConfigLoaded = true
          siteConfigPromise = undefined
        }
      })()

      return siteConfigPromise
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
