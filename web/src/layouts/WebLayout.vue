<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { onMounted } from 'vue'

import { useWebAppStore } from '../store/modules/app'
import { useWebAuthStore } from '../store/modules/auth'

const appStore = useWebAppStore()
const authStore = useWebAuthStore()
const router = useRouter()
const { logoUrl, navigationOpen, siteName, theme } = storeToRefs(appStore)
const { isLoggedIn } = storeToRefs(authStore)

const navItems = [
  { label: '首页', to: '/' },
  { label: '产品', to: '/products' },
  { label: '价格', to: '/pricing' },
]

onMounted(() => {
  void appStore.loadSiteConfig()
})

async function handleLogout() {
  await authStore.logout()
  appStore.closeNavigation()
  await router.replace('/login')
}
</script>

<template>
  <div class="web-shell">
    <header class="web-header">
      <RouterLink to="/" class="brand">
        <img v-if="logoUrl" class="brand-logo" :src="logoUrl" :alt="`${siteName} Logo`">
        <div v-else class="brand-mark">p</div>
        <span class="brand-name">{{ siteName }}</span>
      </RouterLink>

      <button
        class="menu-button"
        type="button"
        :aria-expanded="navigationOpen"
        aria-controls="web-primary-nav"
        @click="appStore.toggleNavigation"
      >
        Menu
      </button>

      <nav
        id="web-primary-nav"
        class="nav-links"
        :class="{ 'is-open': navigationOpen }"
        aria-label="Primary"
      >
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-link"
          @click="appStore.closeNavigation"
        >
          {{ item.label }}
        </RouterLink>
        <RouterLink v-if="isLoggedIn" to="/user" class="nav-link nav-link--primary" @click="appStore.closeNavigation">
          控制台
        </RouterLink>
        <RouterLink v-if="isLoggedIn" to="/user/profile" class="nav-link" @click="appStore.closeNavigation">
          资料
        </RouterLink>
        <RouterLink v-if="!isLoggedIn" to="/login" class="nav-link nav-link--primary" @click="appStore.closeNavigation">
          登录
        </RouterLink>
        <button v-else class="nav-link nav-action nav-link--quiet" type="button" @click="handleLogout">
          退出
        </button>
      </nav>

      <button
        class="theme-toggle"
        type="button"
        :aria-label="theme === 'dark' ? '切换到浅色模式' : '切换到暗黑模式'"
        @click="appStore.toggleTheme"
      >
        <svg v-if="theme === 'dark'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="4"/>
          <path d="M12 2v2"/>
          <path d="M12 20v2"/>
          <path d="m4.93 4.93 1.41 1.41"/>
          <path d="m17.66 17.66 1.41 1.41"/>
          <path d="M2 12h2"/>
          <path d="M20 12h2"/>
          <path d="m6.34 17.66-1.41 1.41"/>
          <path d="m19.07 4.93-1.41 1.41"/>
        </svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>
        </svg>
      </button>
    </header>

    <main class="web-main">
      <RouterView v-slot="{ Component, route }">
        <Transition :name="typeof route.meta.transitionName === 'string' ? route.meta.transitionName : 'page-fade'" mode="out-in">
          <component :is="Component" :key="route.name || route.fullPath" />
        </Transition>
      </RouterView>
    </main>
  </div>
</template>
