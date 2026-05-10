<script setup lang="ts">
import {
  ChevronDownOutline,
  LogOutOutline,
  MenuOutline,
  PersonCircleOutline,
} from '@vicons/ionicons5'
import { NAvatar, NButton, NDropdown, NIcon } from 'naive-ui'
import { computed, h } from 'vue'
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

const dropdownOptions = [
  {
    label: '退出登录',
    key: 'logout',
    icon: () => h(NIcon, null, { default: () => h(LogOutOutline) }),
  },
]

async function handleSelect(key: string) {
  if (key === 'logout') {
    await authStore.logoutRemote()
    await router.replace(ADMIN_ROUTE_PATH.login)
  }
}

const adminLabel = computed(
  () => authStore.admin?.display_name || authStore.admin?.username || '管理员',
)
</script>

<template>
  <div class="app-header">
    <div class="app-header__left">
      <NButton text @click="appStore.toggleSidebar()">
        <NIcon :size="20">
          <MenuOutline />
        </NIcon>
      </NButton>
      <span class="app-header__title">{{ pageTitle }}</span>
    </div>

    <div class="app-header__right">
      <NDropdown :options="dropdownOptions" trigger="click" @select="handleSelect">
        <span class="app-header__user">
          <NAvatar round :size="30">
            <NIcon><PersonCircleOutline /></NIcon>
          </NAvatar>
          <span class="app-header__user-name">{{ adminLabel }}</span>
          <NIcon :size="14"><ChevronDownOutline /></NIcon>
        </span>
      </NDropdown>
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
  color: rgba(15, 23, 42, 0.78);
}

@media (max-width: 640px) {
  .app-header__user-name {
    display: none;
  }
}
</style>
