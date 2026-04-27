<script setup lang="ts">
import { Compass, DataAnalysis, Odometer } from '@element-plus/icons-vue'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useAppStore } from '../../store/modules/app'
import { usePermissionStore } from '../../store/modules/permission'

const appStore = useAppStore()
const permissionStore = usePermissionStore()
const route = useRoute()

const iconMap = {
  Compass,
  DataAnalysis,
  Odometer,
}

const activeMenu = computed(() => route.path)
const menus = computed(() => permissionStore.sidebarMenus)
const showBrandText = computed(() => appStore.sidebarOpened || appStore.device === 'mobile')

function resolveIcon(name: string | null) {
  if (!name) {
    return Compass
  }

  return iconMap[name as keyof typeof iconMap] || Compass
}

function handleSelect() {
  if (appStore.device === 'mobile') {
    appStore.closeSidebar()
  }
}
</script>

<template>
  <div class="app-sidebar">
    <div class="app-sidebar__brand">
      <div class="app-sidebar__brand-mark">
        <el-icon :size="20"><DataAnalysis /></el-icon>
      </div>
      <div v-show="showBrandText" class="app-sidebar__brand-copy">
        <strong>pveCloud Admin</strong>
        <span>Minimal control shell</span>
      </div>
    </div>

    <div v-if="showBrandText" class="app-sidebar__section">
      <span>Navigation</span>
    </div>

    <el-scrollbar class="app-sidebar__scroll">
      <el-menu
        v-if="menus.length > 0"
        class="app-sidebar__menu"
        :default-active="activeMenu"
        :collapse="!appStore.sidebarOpened && appStore.device === 'desktop'"
        :collapse-transition="false"
        router
        @select="handleSelect"
      >
        <el-menu-item v-for="item in menus" :key="item.key" :index="item.path">
          <el-icon>
            <component :is="resolveIcon(item.icon)" />
          </el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>

      <div v-else class="app-sidebar__empty">
        <strong v-show="showBrandText">暂无可访问页面</strong>
        <span v-show="showBrandText">当前账号缺少可展示的后台页面权限。</span>
      </div>
    </el-scrollbar>

    <div v-if="showBrandText" class="app-sidebar__footer">
      <span>Current scope</span>
      <strong>Login / Dashboard / 403</strong>
    </div>
  </div>
</template>

<style scoped>
.app-sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
  color: rgba(241, 245, 249, 0.92);
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.26), transparent 30%),
    linear-gradient(180deg, #06131f 0%, #0a2033 52%, #0e3350 100%);
}

.app-sidebar__brand {
  min-height: 80px;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px;
}

.app-sidebar__brand-mark {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 14px;
  color: #f8fafc;
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.78), rgba(245, 158, 11, 0.78));
  box-shadow: 0 12px 28px rgba(14, 165, 233, 0.28);
}

.app-sidebar__brand-copy {
  display: grid;
  gap: 4px;
}

.app-sidebar__brand-copy strong {
  font-family: var(--pc-display-font);
  font-size: 15px;
}

.app-sidebar__brand-copy span,
.app-sidebar__section,
.app-sidebar__footer span {
  color: rgba(191, 219, 254, 0.72);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.app-sidebar__section {
  padding: 0 18px 10px;
}

.app-sidebar__scroll {
  flex: 1;
  padding: 0 12px;
}

.app-sidebar__menu {
  border-right: none;
  background: transparent;
}

.app-sidebar__menu :deep(.el-menu-item) {
  height: 46px;
  margin-bottom: 6px;
  border-radius: 14px;
  color: rgba(226, 232, 240, 0.84);
}

.app-sidebar__menu :deep(.el-menu-item:hover) {
  color: #ffffff;
  background: rgba(148, 163, 184, 0.12);
}

.app-sidebar__menu :deep(.el-menu-item.is-active) {
  color: #ffffff;
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.24), rgba(245, 158, 11, 0.24));
  box-shadow: inset 0 0 0 1px rgba(125, 211, 252, 0.2);
}

.app-sidebar__empty {
  padding: 18px 10px 18px 6px;
  display: grid;
  gap: 8px;
  color: rgba(203, 213, 225, 0.8);
}

.app-sidebar__empty strong {
  color: #f8fafc;
  font-size: 14px;
}

.app-sidebar__empty span {
  font-size: 12px;
  line-height: 1.6;
}

.app-sidebar__footer {
  padding: 18px;
  display: grid;
  gap: 6px;
  border-top: 1px solid rgba(148, 163, 184, 0.14);
}

.app-sidebar__footer strong {
  color: #f8fafc;
  font-size: 13px;
}
</style>
