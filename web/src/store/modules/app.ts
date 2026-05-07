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
    siteConfigLoading: false,
    siteConfigError: '',
    loginCaptchaEnabled: false,
    registerCaptchaEnabled: false,
    passwordResetRequestCaptchaEnabled: false,
    passwordResetConfirmCaptchaEnabled: false,
    realNameConfig: {
      enabled: false,
      required_for_order: true,
      allowed_providers: [] as string[],
      default_provider: '',
      resubmit_enabled: true,
      max_submit_attempts: 3,
      review_notice: '',
    },
  }),
  actions: {
    async loadSiteConfig(force = false) {
      if (this.siteConfigLoaded && !force) return
      if (siteConfigPromise && !force) return siteConfigPromise

      this.siteConfigLoading = true
      this.siteConfigError = ''
      siteConfigPromise = (async () => {
        try {
          const config = await getSiteConfig()
          this.siteName = config.site_name.trim() || 'pveCloud'
          this.logoUrl = config.logo_url.trim()
          this.loginCaptchaEnabled = config.login_captcha_enabled
          this.registerCaptchaEnabled = config.register_captcha_enabled
          this.passwordResetRequestCaptchaEnabled = config.password_reset_request_captcha_enabled
          this.passwordResetConfirmCaptchaEnabled = config.password_reset_confirm_captcha_enabled
          this.realNameConfig = config.real_name || this.realNameConfig
        } catch {
          this.siteName = 'pveCloud'
          this.logoUrl = ''
          this.loginCaptchaEnabled = false
          this.registerCaptchaEnabled = false
          this.passwordResetRequestCaptchaEnabled = false
          this.passwordResetConfirmCaptchaEnabled = false
          this.siteConfigError = '站点认证配置加载失败，当前已按默认关闭验证码处理。'
        } finally {
          this.siteConfigLoaded = true
          this.siteConfigLoading = false
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
