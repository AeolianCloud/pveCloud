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
  for (let index = route.matched.length - 1; index >= 0; index -= 1) {
    const title = route.matched[index].meta.title
    if (typeof title === 'string' && title.length > 0) {
      return title
    }
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
      <el-button circle text class="app-header__menu" @click="appStore.toggleSidebar()">
        <el-icon :size="18">
          <Expand v-if="!appStore.sidebarOpened" />
          <Fold v-else />
        </el-icon>
      </el-button>

      <div class="app-header__title">
        <span class="app-header__eyebrow">pveCloud Admin</span>
        <strong>{{ pageTitle }}</strong>
      </div>
    </div>

    <div class="app-header__right">
      <div class="app-header__status">
        <span class="app-header__status-dot"></span>
        <span>基础后台已收口为三页模型</span>
      </div>

      <el-dropdown trigger="click">
        <span class="app-header__user">
          <el-avatar :icon="UserFilled" />
          <span class="app-header__user-copy">
            <strong>{{ authStore.admin?.display_name || authStore.admin?.username || '管理员' }}</strong>
            <small>后台会话</small>
          </span>
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
  gap: 20px;
}

.app-header__left,
.app-header__right {
  display: flex;
  align-items: center;
  gap: 14px;
}

.app-header__menu {
  color: var(--pc-title-text);
  background: rgba(255, 255, 255, 0.76);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.16);
}

.app-header__title {
  display: grid;
  gap: 4px;
}

.app-header__eyebrow {
  color: var(--pc-subtle-text);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.app-header__title strong {
  color: var(--pc-title-text);
  font-family: var(--pc-display-font);
  font-size: 20px;
  line-height: 1;
}

.app-header__status {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 999px;
  color: var(--pc-muted-text);
  background: rgba(255, 255, 255, 0.76);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.16);
  font-size: 12px;
}

.app-header__status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--pc-accent), var(--pc-accent-warm));
  box-shadow: 0 0 0 4px rgba(14, 165, 233, 0.12);
}

.app-header__user {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 6px 8px 6px 6px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.76);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.16);
  cursor: pointer;
}

.app-header__user-copy {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.app-header__user-copy strong {
  color: var(--pc-title-text);
  font-size: 13px;
  line-height: 1.2;
}

.app-header__user-copy small {
  color: var(--pc-subtle-text);
  font-size: 11px;
}

@media (max-width: 991px) {
  .app-header__status {
    display: none;
  }
}

@media (max-width: 640px) {
  .app-header__user-copy {
    display: none;
  }
}
</style>
