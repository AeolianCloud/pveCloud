<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterView, useRoute, useRouter } from 'vue-router'

import { useWebAppStore } from '../store/modules/app'
import { useWebAuthStore } from '../store/modules/auth'

const route = useRoute()
const router = useRouter()
const appStore = useWebAppStore()
const authStore = useWebAuthStore()

const isLoggedIn = computed(() => authStore.isLoggedIn)
const displayName = computed(() => authStore.displayName)
const siteName = computed(() => appStore.siteName)
const logoUrl = computed(() => appStore.logoUrl)
const theme = computed(() => appStore.theme)

const isScrolled = ref(false)
const isMobileMenuOpen = ref(false)

function handleScroll() {
  isScrolled.value = window.scrollY > 20
}

onMounted(() => {
  void appStore.loadSiteConfig()
  window.addEventListener('scroll', handleScroll, { passive: true })
  handleScroll()
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})

const navItems = [
  { path: '/', label: '首页' },
  { path: '/products', label: '产品服务' },
  { path: '/pricing', label: '价格方案' },
]

async function handleLogout() {
  await authStore.logout()
  if (route.path.startsWith('/user')) {
    router.push('/login')
  }
}
</script>

<template>
  <div class="web-layout">
    <div class="site-bg">
      <div class="bg-gradient-top"></div>
      <div class="bg-gradient-bottom"></div>
      <div class="bg-grid"></div>
    </div>

    <!-- Header Navigation -->
    <header class="header" :class="{ 'header-scrolled': isScrolled }">
      <div class="container header-inner">
        <RouterLink to="/" class="logo" @click="isMobileMenuOpen = false">
          <div class="logo-mark">
            <img v-if="logoUrl" :src="logoUrl" :alt="siteName" />
            <svg v-else viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" />
              <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
              <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </div>
          <span class="logo-text">{{ siteName }}</span>
        </RouterLink>

        <!-- Desktop Nav -->
        <nav class="desktop-nav">
          <RouterLink 
            v-for="item in navItems" 
            :key="item.path" 
            :to="item.path" 
            class="nav-link"
            :class="{ active: route.path === item.path }"
          >
            {{ item.label }}
          </RouterLink>
        </nav>

        <!-- Auth Actions -->
        <div class="auth-actions desktop-only">
          <button class="theme-toggle" type="button" :aria-label="theme === 'dark' ? '切换浅色主题' : '切换深色主题'" @click="appStore.toggleTheme()">
            <svg v-if="theme === 'dark'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="4"></circle><path d="M12 2v2"></path><path d="M12 20v2"></path><path d="m4.93 4.93 1.41 1.41"></path><path d="m17.66 17.66 1.41 1.41"></path><path d="M2 12h2"></path><path d="M20 12h2"></path><path d="m6.34 17.66-1.41 1.41"></path><path d="m19.07 4.93-1.41 1.41"></path></svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 3a6 6 0 0 0 9 7.5A9 9 0 1 1 12 3Z"></path></svg>
          </button>
          <template v-if="!isLoggedIn">
            <RouterLink to="/login" class="btn btn-text">登录</RouterLink>
            <RouterLink to="/register" class="btn btn-primary btn-sm" style="border-radius: 99px;">注册</RouterLink>
          </template>
          <template v-else>
            <div class="user-dropdown">
              <button class="user-btn">
                <div class="avatar">{{ displayName.charAt(0).toUpperCase() }}</div>
                <span>{{ displayName }}</span>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"></polyline></svg>
              </button>
              <div class="dropdown-menu">
                <RouterLink to="/user" class="dropdown-item">控制台</RouterLink>
                <div class="dropdown-divider"></div>
                <button class="dropdown-item text-error" @click="handleLogout">退出登录</button>
              </div>
            </div>
            <RouterLink to="/user" class="btn btn-primary btn-sm" style="border-radius: 99px; margin-left: 12px;">控制台</RouterLink>
          </template>
        </div>

        <!-- Mobile Menu Toggle -->
        <button class="mobile-toggle" @click="isMobileMenuOpen = !isMobileMenuOpen">
          <svg v-if="!isMobileMenuOpen" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="6" x2="21" y2="6"></line><line x1="3" y1="18" x2="21" y2="18"></line></svg>
          <svg v-else width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        </button>
      </div>
    </header>

    <!-- Mobile Menu Overlay -->
    <div class="mobile-menu" :class="{ 'is-open': isMobileMenuOpen }">
      <nav class="mobile-nav-links">
        <RouterLink 
          v-for="item in navItems" 
          :key="item.path" 
          :to="item.path" 
          class="mobile-nav-link"
          @click="isMobileMenuOpen = false"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
      <div class="mobile-auth">
        <button class="btn btn-block btn-outline" type="button" @click="appStore.toggleTheme()">
          {{ theme === 'dark' ? '切换浅色主题' : '切换深色主题' }}
        </button>
        <template v-if="!isLoggedIn">
          <RouterLink to="/login" class="btn btn-block btn-outline" style="margin-top: 12px;" @click="isMobileMenuOpen = false">登录</RouterLink>
          <RouterLink to="/register" class="btn btn-block btn-primary" style="margin-top: 12px;" @click="isMobileMenuOpen = false">注册</RouterLink>
        </template>
        <template v-else>
          <RouterLink to="/user" class="btn btn-block btn-primary" style="margin-top: 12px;" @click="isMobileMenuOpen = false">进入控制台</RouterLink>
          <button class="btn btn-block btn-outline text-error" style="margin-top: 12px;" @click="handleLogout(); isMobileMenuOpen = false">退出登录</button>
        </template>
      </div>
    </div>

    <!-- Main Content -->
    <main class="main-content">
      <RouterView />
    </main>

    <!-- Fat Footer -->
    <footer class="footer">
      <div class="container">
        <div class="footer-grid grid gap-8 md:grid-cols-2 lg:grid-cols-4">
          <div class="footer-brand">
            <RouterLink to="/" class="logo">
              <div class="logo-mark">
                <img v-if="logoUrl" :src="logoUrl" :alt="siteName" />
                <svg v-else viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor"/></svg>
              </div>
              <span class="logo-text">{{ siteName }}</span>
            </RouterLink>
            <p class="footer-desc">展示当前开放的服务器产品目录、价格、销售地域和系统模板，购买能力后续开放。</p>
            <div class="social-links">
              <a href="#" class="social-link">GH</a>
              <a href="#" class="social-link">TW</a>
              <a href="#" class="social-link">DC</a>
            </div>
          </div>
          
          <div class="footer-links">
            <h3>云产品</h3>
            <RouterLink to="/products">弹性云服务器</RouterLink>
            <RouterLink to="/pricing">套餐价格</RouterLink>
            <RouterLink to="/products">销售地域</RouterLink>
            <RouterLink to="/products">系统模板</RouterLink>
          </div>
          
          <div class="footer-links">
            <h3>用户入口</h3>
            <RouterLink to="/login">登录</RouterLink>
            <RouterLink to="/register">注册</RouterLink>
            <RouterLink to="/forgot-password">找回密码</RouterLink>
            <RouterLink to="/user">用户中心</RouterLink>
          </div>
          
          <div class="footer-links">
            <h3>当前范围</h3>
            <RouterLink to="/products">产品展示</RouterLink>
            <RouterLink to="/pricing">价格展示</RouterLink>
            <RouterLink to="/user/profile">账号资料</RouterLink>
            <RouterLink to="/">返回首页</RouterLink>
          </div>
        </div>
        
        <div class="footer-bottom">
          <p>&copy; 2026 PVECloud Inc. All rights reserved.</p>
          <div class="status-indicator">
            <span class="status-dot"></span> 所有系统运行正常
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.web-layout {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Background Effects */
.site-bg {
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  background-color: var(--c-bg);
}
.bg-gradient-top {
  position: absolute;
  top: -20%; left: -10%; right: -10%; height: 60vh;
  background: radial-gradient(circle at 50% 0%, rgba(59, 130, 246, 0.15) 0%, transparent 60%);
}
.bg-gradient-bottom {
  position: absolute;
  bottom: -20%; left: -10%; right: -10%; height: 60vh;
  background: radial-gradient(circle at 80% 100%, rgba(139, 92, 246, 0.1) 0%, transparent 50%);
}
.bg-grid {
  position: absolute;
  inset: 0;
  background-image: linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px),
                    linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: radial-gradient(ellipse at center, black 20%, transparent 80%);
  -webkit-mask-image: radial-gradient(ellipse at center, black 20%, transparent 80%);
}
[data-theme="light"] .bg-grid {
  background-image: linear-gradient(rgba(0,0,0,0.03) 1px, transparent 1px),
                    linear-gradient(90deg, rgba(0,0,0,0.03) 1px, transparent 1px);
}

/* Header */
.header {
  position: fixed;
  top: 0; left: 0; right: 0;
  height: 72px;
  z-index: 100;
  transition: all 0.3s ease;
  border-bottom: 1px solid transparent;
}
.header-scrolled {
  background: var(--c-header-bg);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-bottom: 1px solid var(--c-border);
  box-shadow: var(--shadow-sm);
}
.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

/* Logo */
.logo {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  z-index: 101; /* Above mobile menu */
}
.logo-mark {
  width: 32px; height: 32px;
  color: var(--c-primary);
}
.logo-mark img { width: 100%; height: 100%; object-fit: contain; }
.logo-text {
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--c-text);
}

/* Desktop Nav */
.desktop-nav {
  display: none;
  align-items: center;
  gap: 8px;
}
@media (min-width: 768px) {
  .desktop-nav { display: flex; }
}
.nav-link {
  padding: 8px 16px;
  border-radius: 99px;
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--c-text-2);
  transition: all var(--transition-fast);
}
.nav-link:hover { color: var(--c-text); background: var(--c-surface-dim); }
.nav-link.active { color: var(--c-primary); background: var(--c-primary-soft); font-weight: 600; }

/* Auth Actions */
.auth-actions { display: none; align-items: center; gap: 8px; }
@media (min-width: 768px) { .auth-actions { display: flex; } }

.theme-toggle {
  width: 38px;
  height: 38px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--c-border);
  border-radius: 999px;
  color: var(--c-text-2);
  background: var(--c-surface-dim);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.theme-toggle:hover {
  color: var(--c-primary);
  border-color: var(--c-primary);
  background: var(--c-primary-soft);
}
.theme-toggle svg { width: 18px; height: 18px; }

/* User Dropdown */
.user-dropdown { position: relative; }
.user-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px 6px 6px;
  border-radius: 99px;
  border: 1px solid var(--c-border);
  background: var(--c-surface);
  color: var(--c-text);
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.user-btn:hover { border-color: var(--c-primary); }
.avatar {
  width: 24px; height: 24px;
  border-radius: 50%;
  background: var(--c-primary);
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.75rem; font-weight: 700;
}
.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px); right: 0;
  width: 200px;
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: var(--radius);
  padding: 8px;
  box-shadow: var(--shadow-lg);
  opacity: 0; visibility: hidden;
  transform: translateY(10px);
  transition: all var(--transition-fast);
}
.user-dropdown:hover .dropdown-menu,
.user-dropdown:focus-within .dropdown-menu {
  opacity: 1; visibility: visible; transform: translateY(0);
}
.dropdown-item {
  display: block; width: 100%;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  text-align: left;
  font-size: 0.95rem;
  color: var(--c-text-2);
  cursor: pointer;
}
.dropdown-item:hover { background: var(--c-surface-dim); color: var(--c-text); }
.dropdown-divider { height: 1px; background: var(--c-border-light); margin: 4px 0; }
.text-error { color: var(--c-error) !important; }

/* Mobile Menu */
.mobile-toggle {
  display: block;
  color: var(--c-text);
  cursor: pointer;
  z-index: 101;
}
@media (min-width: 768px) { .mobile-toggle { display: none; } }

.mobile-menu {
  position: fixed;
  inset: 0;
  background: var(--c-bg);
  z-index: 100;
  padding: 80px 24px 24px;
  display: flex;
  flex-direction: column;
  opacity: 0;
  visibility: hidden;
  transform: translateY(-20px);
  transition: all var(--transition);
}
.mobile-menu.is-open {
  opacity: 1; visibility: visible; transform: translateY(0);
}
.mobile-nav-links { display: flex; flex-direction: column; gap: 16px; margin-bottom: 40px; }
.mobile-nav-link {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--c-text);
  padding: 12px 0;
  border-bottom: 1px solid var(--c-border-light);
}

/* Main */
.main-content {
  flex: 1;
  padding-top: 72px; /* Header height */
}

/* Footer */
.footer {
  border-top: 1px solid var(--c-border);
  background: var(--c-bg-alt);
  padding: 80px 0 40px;
  margin-top: auto;
}
.footer-desc {
  color: var(--c-text-2);
  font-size: 0.95rem;
  line-height: 1.6;
  margin: 20px 0;
}
.social-links { display: flex; gap: 12px; }
.social-link {
  width: 40px; height: 40px;
  border-radius: 50%;
  background: var(--c-surface-dim);
  display: flex; align-items: center; justify-content: center;
  color: var(--c-text-2); font-weight: 600; font-size: 0.8rem;
  transition: all var(--transition-fast);
}
.social-link:hover { background: var(--c-primary); color: #fff; transform: translateY(-2px); }

.footer-links h3 {
  font-size: 1.1rem; font-weight: 700; color: var(--c-text); margin-bottom: 24px;
}
.footer-links a {
  display: block; color: var(--c-text-2); font-size: 0.95rem; margin-bottom: 12px;
  transition: color var(--transition-fast);
}
.footer-links a:hover { color: var(--c-primary); }

.footer-bottom {
  margin-top: 64px;
  padding-top: 24px;
  border-top: 1px solid var(--c-border-light);
  display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 16px;
  color: var(--c-text-3); font-size: 0.9rem;
}
.status-indicator { display: flex; align-items: center; gap: 8px; color: var(--c-text-2); }
.status-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--c-success); box-shadow: 0 0 8px var(--c-success-soft); }
</style>
