<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

import { useWebAppStore } from '../store/modules/app'
import { useWebAuthStore } from '../store/modules/auth'

const route = useRoute()
const router = useRouter()
const appStore = useWebAppStore()
const authStore = useWebAuthStore()

const siteName = computed(() => appStore.siteName)
const logoUrl = computed(() => appStore.logoUrl)
const isLoggedIn = computed(() => authStore.isLoggedIn)
const displayName = computed(() => authStore.displayName)
const currentYear = new Date().getFullYear()

const navItems = [
  { path: '/', label: '首页' },
  { path: '/products', label: '产品' },
  { path: '/pricing', label: '价格' },
]

onMounted(() => {
  void appStore.loadSiteConfig()
})

async function handleLogout() {
  await authStore.logout()
  if (route.path.startsWith('/user')) {
    await router.replace('/login')
  }
}
</script>

<template>
  <div class="web-shell">
    <header class="site-header">
      <div class="container header-inner">
        <RouterLink to="/" class="brand">
          <span class="brand-mark">
            <img v-if="logoUrl" :src="logoUrl" :alt="siteName" />
            <span v-else>{{ siteName.slice(0, 1).toUpperCase() }}</span>
          </span>
          <span class="brand-text">{{ siteName }}</span>
        </RouterLink>

        <nav class="desktop-nav">
          <RouterLink v-for="item in navItems" :key="item.path" :to="item.path" class="nav-link" :class="{ active: route.path === item.path }">
            {{ item.label }}
          </RouterLink>
        </nav>

        <div class="header-actions">
          <template v-if="!isLoggedIn">
            <RouterLink to="/login" class="btn btn-text">登录</RouterLink>
            <RouterLink to="/register" class="btn btn-primary btn-sm">注册</RouterLink>
          </template>
          <template v-else>
            <span class="user-chip">{{ displayName }}</span>
            <RouterLink to="/user" class="btn btn-outline btn-sm">控制台</RouterLink>
            <button class="btn btn-text" type="button" @click="handleLogout">退出</button>
          </template>
        </div>
      </div>
    </header>

    <main class="site-main">
      <RouterView />
    </main>

    <footer class="site-footer">
      <div class="container footer-inner">
        <div>
          <p class="footer-title">{{ siteName }}</p>
          <p class="footer-text">公开产品展示、账号自助和个人实名入口。</p>
        </div>
        <p class="footer-copy">© {{ currentYear }} {{ siteName }}</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.web-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.site-header {
  position: sticky;
  top: 0;
  z-index: 30;
  border-bottom: 1px solid var(--c-border);
  background: var(--c-header-bg);
  backdrop-filter: blur(16px);
}

.header-inner {
  min-height: 72px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  font-weight: 800;
}

.brand-mark {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 12px;
  color: #fff;
  background: linear-gradient(135deg, var(--c-primary), var(--c-primary-strong));
  overflow: hidden;
}

.brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.brand-text {
  min-width: 0;
  font-size: 1rem;
  letter-spacing: -0.03em;
}

.desktop-nav {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-link {
  padding: 8px 12px;
  border-radius: 10px;
  color: var(--c-text-2);
  font-weight: 700;
}

.nav-link.active {
  color: var(--c-primary);
  background: var(--c-primary-soft);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-chip {
  max-width: 160px;
  padding: 8px 12px;
  border: 1px solid var(--c-border);
  border-radius: 999px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--c-text-2);
}

.site-main {
  flex: 1;
}

.site-footer {
  border-top: 1px solid var(--c-border);
  background: var(--c-surface);
}

.footer-inner {
  min-height: 72px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.footer-title {
  font-weight: 800;
}

.footer-text,
.footer-copy {
  color: var(--c-text-2);
}

@media (max-width: 860px) {
  .header-inner,
  .footer-inner {
    flex-direction: column;
    align-items: flex-start;
    padding-block: 14px;
  }

  .desktop-nav {
    flex-wrap: wrap;
  }

  .header-actions {
    flex-wrap: wrap;
  }
}
</style>
