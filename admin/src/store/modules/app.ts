import { defineStore } from 'pinia'

const APP_STORAGE_KEY = 'pvecloud_admin_app'
const MOBILE_BREAKPOINT = 992

type DeviceType = 'desktop' | 'mobile'

interface AppState {
  sidebarOpened: boolean
  device: DeviceType
}

function loadSidebarOpened() {
  const raw = localStorage.getItem(APP_STORAGE_KEY)
  if (!raw) {
    return true
  }

  try {
    const snapshot = JSON.parse(raw) as { sidebarOpened?: boolean }
    return snapshot.sidebarOpened ?? true
  } catch {
    localStorage.removeItem(APP_STORAGE_KEY)
    return true
  }
}

export const useAppStore = defineStore('admin-app', {
  state: (): AppState => ({
    sidebarOpened: loadSidebarOpened(),
    device: 'desktop',
  }),
  actions: {
    toggleSidebar() {
      this.sidebarOpened = !this.sidebarOpened
      this.persist()
    },
    closeSidebar() {
      this.sidebarOpened = false
      this.persist()
    },
    syncDevice(width: number) {
      this.device = width < MOBILE_BREAKPOINT ? 'mobile' : 'desktop'
      if (this.device === 'mobile') {
        this.sidebarOpened = false
      }
    },
    persist() {
      localStorage.setItem(
        APP_STORAGE_KEY,
        JSON.stringify({
          sidebarOpened: this.sidebarOpened,
        }),
      )
    },
  },
})
