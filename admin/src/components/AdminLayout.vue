<script setup lang="ts">
import { computed } from 'vue'
import {
  ClipboardList,
  LayoutDashboard,
  LogOut,
  Menu,
  MessageSquare,
  Package,
  ReceiptText,
  Server,
  Settings,
  ShieldCheck,
  Users,
  WalletCards,
} from 'lucide-vue-next'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

import { fallbackAdminMenus, visibleAdminMenus } from '../constants/adminMenu'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const iconMap = {
  'clipboard-list': ClipboardList,
  'layout-dashboard': LayoutDashboard,
  'message-square': MessageSquare,
  package: Package,
  'receipt-text': ReceiptText,
  server: Server,
  settings: Settings,
  'shield-check': ShieldCheck,
  users: Users,
  'wallet-cards': WalletCards,
}

const menus = computed(() =>
  visibleAdminMenus(auth.permissionCodes, auth.menuItems.length > 0 ? auth.menuItems : fallbackAdminMenus),
)

const pageTitle = computed(() => {
  const matchedMenu = menus.value.find((item) => item.path === route.path)
  return matchedMenu?.title || String(route.meta.title || '管理后台')
})

function iconFor(name: string | null) {
  if (!name) {
    return LayoutDashboard
  }
  return iconMap[name as keyof typeof iconMap] || LayoutDashboard
}

async function logout() {
  await auth.logoutRemote()
  await router.replace({ name: 'login' })
}
</script>

<template>
  <div class="admin-shell" :class="{ 'admin-shell--collapsed': auth.sidebarCollapsed }">
    <aside class="admin-sidebar">
      <div class="sidebar-brand">
        <ShieldCheck :size="24" aria-hidden="true" />
        <span>pveCloud</span>
      </div>

      <nav class="sidebar-nav" aria-label="管理菜单">
        <RouterLink
          v-for="item in menus"
          :key="item.key"
          class="sidebar-link"
          :to="item.path"
          :title="item.title"
        >
          <component :is="iconFor(item.icon)" :size="18" aria-hidden="true" />
          <span>{{ item.title }}</span>
        </RouterLink>
      </nav>
    </aside>

    <div class="admin-workspace">
      <header class="admin-topbar">
        <button
          class="icon-only-button"
          type="button"
          title="折叠侧边栏"
          @click="auth.toggleSidebar()"
        >
          <Menu :size="19" aria-hidden="true" />
        </button>

        <div class="topbar-title">
          <span class="eyebrow">pveCloud Admin</span>
          <h1>{{ pageTitle }}</h1>
        </div>

        <div class="topbar-actions">
          <div class="admin-identity">
            <strong>{{ auth.admin?.display_name || auth.admin?.username }}</strong>
            <span>{{ auth.admin?.username }}</span>
          </div>
          <button class="icon-button" type="button" title="退出登录" @click="logout">
            <LogOut :size="18" aria-hidden="true" />
            <span>退出</span>
          </button>
        </div>
      </header>

      <main class="admin-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>
