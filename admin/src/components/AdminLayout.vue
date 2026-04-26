<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  ChevronDown,
  ChevronLeft,
  ClipboardList,
  Cloud,
  LayoutDashboard,
  LogOut,
  Megaphone,
  Menu,
  MessageSquare,
  MonitorCheck,
  Package,
  Palette,
  ReceiptText,
  Server,
  Settings,
  ShieldAlert,
  ShieldCheck,
  UserRound,
  Users,
  WalletCards,
} from 'lucide-vue-next'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

import { fallbackAdminMenus, visibleAdminMenus } from '../constants/adminMenu'
import { useAuthStore } from '../stores/auth'

const THEME_STORAGE_KEY = 'pvecloud_admin_theme'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const colorTheme = ref(localStorage.getItem(THEME_STORAGE_KEY) === 'deep' ? 'deep' : 'default')

const iconMap = {
  'clipboard-list': ClipboardList,
  cloud: Cloud,
  'layout-dashboard': LayoutDashboard,
  megaphone: Megaphone,
  'message-square': MessageSquare,
  'monitor-check': MonitorCheck,
  package: Package,
  'receipt-text': ReceiptText,
  server: Server,
  settings: Settings,
  'shield-alert': ShieldAlert,
  'shield-check': ShieldCheck,
  users: Users,
  'wallet-cards': WalletCards,
}

const menus = computed(() =>
  visibleAdminMenus(auth.permissionCodes, auth.menuItems.length > 0 ? auth.menuItems : fallbackAdminMenus),
)

const pageTitle = computed(() => {
  const matchedMenu = menus.value.find((item) => item.path === route.path)
  return matchedMenu?.title || String(route.meta.title || '控制台')
})

const themeLabel = computed(() => (colorTheme.value === 'deep' ? '深色主题' : '浅色主题'))
const sidebarLabel = computed(() => (auth.sidebarCollapsed ? '展开侧边栏' : '收起侧边栏'))

watch(
  colorTheme,
  (theme) => {
    localStorage.setItem(THEME_STORAGE_KEY, theme)
    document.documentElement.classList.toggle('admin-theme-dark', theme === 'deep')
  },
  {
    immediate: true,
  },
)

function iconFor(name: string | null) {
  if (!name) {
    return LayoutDashboard
  }
  return iconMap[name as keyof typeof iconMap] || LayoutDashboard
}

function toggleTheme() {
  colorTheme.value = colorTheme.value === 'deep' ? 'default' : 'deep'
}

async function logout() {
  await auth.logoutRemote()
  await router.replace({ name: 'login' })
}
</script>

<template>
  <div
    class="admin-shell"
    :class="[
      { 'admin-shell--collapsed': auth.sidebarCollapsed },
      colorTheme === 'deep' ? 'admin-shell--deep' : 'admin-shell--default',
    ]"
  >
    <aside class="admin-sidebar">
      <div class="sidebar-brand">
        <span class="brand-mark">
          <Server :size="19" aria-hidden="true" />
        </span>
        <span class="brand-name">IDC 云服务器销售系统</span>
      </div>

      <nav class="sidebar-nav" aria-label="后台导航">
        <RouterLink
          v-for="item in menus"
          :key="item.key"
          class="sidebar-link"
          :to="item.path"
          :title="item.title"
        >
          <component :is="iconFor(item.icon)" :size="17" aria-hidden="true" />
          <span>{{ item.title }}</span>
          <ChevronDown class="sidebar-link-chevron" :size="13" aria-hidden="true" />
        </RouterLink>
      </nav>

      <Button class="sidebar-collapse" text severity="secondary" :title="sidebarLabel" :aria-label="sidebarLabel" @click="auth.toggleSidebar()">
        <ChevronLeft :size="17" aria-hidden="true" />
        <span>{{ auth.sidebarCollapsed ? '展开侧边栏' : '收起侧边栏' }}</span>
      </Button>
    </aside>

    <div class="admin-workspace">
      <header class="admin-topbar">
        <div class="topbar-breadcrumb">
          <Button class="mobile-menu-button" text severity="secondary" :title="sidebarLabel" :aria-label="sidebarLabel" @click="auth.toggleSidebar()">
            <Menu :size="17" aria-hidden="true" />
          </Button>
          <span>控制台</span>
          <span class="breadcrumb-separator">/</span>
          <strong>{{ pageTitle }}</strong>
        </div>

        <div class="topbar-actions">
          <Button class="theme-switch-button" text severity="secondary" :title="`切换主题，当前为${themeLabel}`" :aria-label="`切换主题，当前为${themeLabel}`" @click="toggleTheme">
            <Palette :size="16" aria-hidden="true" />
            <span>{{ themeLabel }}</span>
          </Button>
          <div class="admin-profile">
            <Avatar class="profile-avatar" shape="circle">
              <template #default>
                <UserRound :size="18" aria-hidden="true" />
              </template>
            </Avatar>
            <strong>{{ auth.admin?.display_name || auth.admin?.username || 'admin' }}</strong>
            <Button class="profile-logout" text severity="secondary" title="退出登录" aria-label="退出登录" @click="logout">
              <LogOut :size="15" aria-hidden="true" />
            </Button>
          </div>
        </div>
      </header>

      <main class="admin-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
