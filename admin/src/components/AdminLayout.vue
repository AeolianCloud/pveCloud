<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  Bell,
  ChevronDown,
  ChevronLeft,
  CircleHelp,
  ClipboardList,
  Cloud,
  LayoutDashboard,
  LogOut,
  Megaphone,
  Menu,
  MessageSquare,
  Package,
  Palette,
  ReceiptText,
  Search,
  Server,
  Settings,
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
  package: Package,
  'receipt-text': ReceiptText,
  server: Server,
  settings: Settings,
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

const themeLabel = computed(() => (colorTheme.value === 'deep' ? '深蓝主题' : '默认主题'))

watch(colorTheme, (theme) => {
  localStorage.setItem(THEME_STORAGE_KEY, theme)
})

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
        <span class="brand-name">IDC云服务器销售系统</span>
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

      <button
        class="sidebar-collapse"
        type="button"
        :title="auth.sidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
        :aria-label="auth.sidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
        @click="auth.toggleSidebar()"
      >
        <ChevronLeft :size="17" aria-hidden="true" />
        <span>收起侧边栏</span>
      </button>
    </aside>

    <div class="admin-workspace">
      <header class="admin-topbar">
        <div class="topbar-breadcrumb">
          <button
            class="mobile-menu-button"
            type="button"
            :title="auth.sidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
            :aria-label="auth.sidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
            @click="auth.toggleSidebar()"
          >
            <Menu :size="17" aria-hidden="true" />
          </button>
          <span>控制台</span>
          <span class="breadcrumb-separator">/</span>
          <strong>{{ pageTitle }}</strong>
        </div>

        <label class="topbar-search" aria-label="全局搜索">
          <Search :size="17" aria-hidden="true" />
          <input type="search" placeholder="搜索订单、客户、实例、工单" />
          <kbd>Ctrl K</kbd>
        </label>

        <div class="topbar-actions">
          <button
            class="theme-switch-button"
            type="button"
            :title="`切换主题，当前为${themeLabel}`"
            :aria-label="`切换主题，当前为${themeLabel}`"
            @click="toggleTheme"
          >
            <Palette :size="16" aria-hidden="true" />
            <span>{{ themeLabel }}</span>
          </button>
          <button class="topbar-icon-button notification-button" type="button" title="通知" aria-label="通知">
            <Bell :size="18" aria-hidden="true" />
            <span>12</span>
          </button>
          <button class="topbar-icon-button" type="button" title="帮助" aria-label="帮助">
            <CircleHelp :size="18" aria-hidden="true" />
          </button>
          <div class="admin-profile">
            <span class="profile-avatar">
              <UserRound :size="18" aria-hidden="true" />
            </span>
            <strong>{{ auth.admin?.display_name || auth.admin?.username || 'admin' }}</strong>
            <button class="profile-logout" type="button" title="退出登录" aria-label="退出登录" @click="logout">
              <LogOut :size="15" aria-hidden="true" />
            </button>
          </div>
        </div>
      </header>

      <main class="admin-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
