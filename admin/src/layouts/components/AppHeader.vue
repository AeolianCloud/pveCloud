<script setup lang="ts">
import { ArrowDown, Expand, Fold, SwitchButton, UserFilled } from '@element-plus/icons-vue'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ADMIN_ROUTE_PATH } from '../../router/constants'
import { useAppStore } from '../../store/modules/app'
import { useAuthStore } from '../../store/modules/auth'

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const pageTitle = computed(() => {
  for (let i = route.matched.length - 1; i >= 0; i -= 1) {
    const title = route.matched[i].meta.title
    if (typeof title === 'string' && title.length > 0) return title
  }
  return '管理后台'
})

async function logout() {
  await authStore.logoutRemote()
  await router.replace(ADMIN_ROUTE_PATH.login)
}
</script>

<template>
  <div class="app-header">
    <div class="app-header__left">
      <el-button circle text @click="appStore.toggleSidebar()">
        <el-icon :size="18">
          <Expand v-if="!appStore.sidebarOpened" />
          <Fold v-else />
        </el-icon>
      </el-button>
      <span class="app-header__title">{{ pageTitle }}</span>
    </div>

    <div class="app-header__right">
      <el-dropdown trigger="click">
        <span class="app-header__user">
          <el-avatar :size="30" :icon="UserFilled" />
          <span>{{ authStore.admin?.display_name || authStore.admin?.username || '管理员' }}</span>
          <el-icon><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="logout">
              <el-icon><SwitchButton /></el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<style scoped>
.app-header {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.app-header__left,
.app-header__right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.app-header__title {
  font-size: 16px;
  font-weight: 600;
}

.app-header__user {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--el-text-color-regular);
}

@media (max-width: 640px) {
  .app-header__user span:nth-child(2) {
    display: none;
  }
}
</style>
