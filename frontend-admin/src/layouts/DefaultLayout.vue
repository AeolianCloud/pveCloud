<script setup lang="ts">
import { h, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { GridOutline, PersonOutline, ChevronDownOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()

// 当前页面标题，从路由 meta 取
const pageTitle = computed(() => (route.meta.title as string) || '')

// ── 菜单 ──────────────────────────────────────────────────

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// 系统管理子菜单定义，每项附带所需权限标识
const systemChildren = [
  { label: '管理员账号', key: 'admin-users', permission: 'admin:list' },
  { label: '角色管理',   key: 'roles',       permission: 'role:list'  },
  { label: '登录日志',   key: 'login-logs',  permission: 'log:list'   },
  { label: '操作日志',   key: 'op-logs',     permission: 'op:list'    },
]

// 动态菜单：根据当前用户权限过滤子菜单项
const menuOptions = computed<MenuOption[]>(() => {
  const filteredChildren = systemChildren.filter(item =>
    authStore.hasPermission(item.permission)
  )

  const menus: MenuOption[] = [
    {
      label: '控制台',
      key: 'dashboard',
      icon: renderIcon(GridOutline),
    },
  ]

  // 只有存在至少一个可见子菜单时才显示"系统管理"
  if (filteredChildren.length > 0) {
    menus.push({
      label: '系统管理',
      key: 'system',
      icon: renderIcon(PersonOutline),
      children: filteredChildren.map(({ label, key }) => ({ label, key })),
    })
  }

  return menus
})

// 路由路径 → 菜单 key 映射，用于高亮当前菜单项
const routeKeyMap: Record<string, string> = {
  '/dashboard': 'dashboard',
  '/system/admin-users': 'admin-users',
  '/system/roles': 'roles',
  '/system/login-logs': 'login-logs',
  '/system/op-logs': 'op-logs',
}

const activeMenuKey = computed(() => routeKeyMap[route.path] ?? '')

// 菜单点击 → 跳转路由
function handleMenuUpdate(key: string) {
  const pathMap: Record<string, string> = {
    dashboard: '/dashboard',
    'admin-users': '/system/admin-users',
    'roles': '/system/roles',
    'login-logs': '/system/login-logs',
    'op-logs': '/system/op-logs',
  }
  if (pathMap[key]) router.push(pathMap[key])
}

// ── 用户下拉菜单 ──────────────────────────────────────────

const userMenuOptions = [
  { label: '退出登录', key: 'logout' },
]

function handleUserMenu(key: string) {
  if (key === 'logout') {
    authStore.logout()
    message.success('已退出登录')
    router.push('/login')
  }
}
</script>

<template>
  <n-layout style="height: 100vh" has-sider>
    <!-- 侧边栏 -->
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="220"
      show-trigger="bar"
      :native-scrollbar="false"
    >
      <!-- Logo 区 -->
      <div class="sider-logo">
        <div class="sider-logo-icon">
          <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="4" y="4" width="18" height="18" rx="3" fill="#4fa8e8" />
            <rect x="26" y="4" width="18" height="18" rx="3" fill="#4fa8e8" opacity="0.6" />
            <rect x="4" y="26" width="18" height="18" rx="3" fill="#4fa8e8" opacity="0.6" />
            <rect x="26" y="26" width="18" height="18" rx="3" fill="#4fa8e8" opacity="0.85" />
          </svg>
        </div>
        <span class="sider-logo-text">pveCloud</span>
      </div>

      <!-- 导航菜单 -->
      <n-menu
        :options="menuOptions"
        :indent="18"
        :value="activeMenuKey"
        @update:value="handleMenuUpdate"
      />
    </n-layout-sider>

    <!-- 右侧主体 -->
    <n-layout>
      <!-- 顶部栏 -->
      <n-layout-header bordered class="header">
        <div class="header-left">
          <span class="header-title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <n-dropdown trigger="hover" :options="userMenuOptions" @select="handleUserMenu">
            <div class="user-area">
              <n-avatar
                round
                size="small"
                color="#4fa8e8"
                style="color: #fff; font-size: 13px; font-weight: 600;"
              >
                {{ authStore.user?.nickname?.charAt(0)?.toUpperCase() || 'A' }}
              </n-avatar>
              <span class="user-name">{{ authStore.user?.nickname || authStore.user?.username }}</span>
              <n-icon size="14" style="color: #c2c2c2;"><ChevronDownOutline /></n-icon>
            </div>
          </n-dropdown>
        </div>
      </n-layout-header>

      <!-- 内容区 -->
      <n-layout-content class="main-content">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<style scoped>
/* ========== 侧边栏 Logo ========== */
.sider-logo {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  border-bottom: 1px solid #efeff5;
  overflow: hidden;
  white-space: nowrap;
}

.sider-logo-icon {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}

.sider-logo-text {
  font-size: 16px;
  font-weight: 700;
  color: #18181c;
  letter-spacing: 0.5px;
}

/* ========== 顶部栏 ========== */
.header {
  height: 56px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-title {
  font-size: 15px;
  font-weight: 600;
  color: #18181c;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-area {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.15s;
}

.user-area:hover {
  background: #f5f5f5;
}

.user-name {
  font-size: 13px;
  color: #333;
}

/* ========== 内容区 ========== */
.main-content {
  padding: 24px;
  background: #f7f8fa;
  min-height: calc(100vh - 56px);
}
</style>
