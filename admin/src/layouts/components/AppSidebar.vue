<script setup lang="ts">
import { Compass, DataAnalysis, Odometer, Setting } from '@element-plus/icons-vue'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useAppStore } from '../../store/modules/app'
import { usePermissionStore } from '../../store/modules/permission'

const appStore = useAppStore()
const permissionStore = usePermissionStore()
const route = useRoute()

const iconMap = { Compass, DataAnalysis, Odometer, Setting }

const activeMenu = computed(() => route.path)
const menus = computed(() => permissionStore.sidebarMenus)
const collapsed = computed(() => !appStore.sidebarOpened && appStore.device === 'desktop')

function resolveIcon(name: string | null) {
  if (!name) return Compass
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
    <div class="app-sidebar__brand" :class="{ 'app-sidebar__brand--collapsed': collapsed }">
      <div class="app-sidebar__logo">
        <el-icon :size="18"><DataAnalysis /></el-icon>
      </div>
      <span v-show="!collapsed" class="app-sidebar__name">PVE Cloud</span>
    </div>

    <el-menu
      v-if="menus.length > 0"
      class="app-sidebar__menu"
      :default-active="activeMenu"
      :collapse="collapsed"
      :collapse-transition="false"
      background-color="#001529"
      text-color="rgba(255,255,255,0.7)"
      active-text-color="#ffffff"
      router
      @select="handleSelect"
    >
      <template v-for="item in menus" :key="item.key">
        <el-sub-menu v-if="item.children?.length" :index="item.path">
          <template #title>
            <el-icon>
              <component :is="resolveIcon(item.icon)" />
            </el-icon>
            <span>{{ item.title }}</span>
          </template>
          <el-menu-item v-for="child in item.children" :key="child.key" :index="child.path">
            <template #title>{{ child.title }}</template>
          </el-menu-item>
        </el-sub-menu>
        <el-menu-item v-else :index="item.path">
          <el-icon>
            <component :is="resolveIcon(item.icon)" />
          </el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </template>
    </el-menu>

    <div v-else class="app-sidebar__empty">
      <span>暂无菜单权限</span>
    </div>
  </div>
</template>

<style scoped>
.app-sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #001529;
  overflow: hidden;
}

.app-sidebar__brand {
  height: 60px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.app-sidebar__brand--collapsed {
  justify-content: center;
  padding: 0;
}

.app-sidebar__logo {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  color: #fff;
  background: var(--el-color-primary);
  flex-shrink: 0;
}

.app-sidebar__name {
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  white-space: nowrap;
}

.app-sidebar__menu {
  flex: 1;
  border-right: none;
}

.app-sidebar__empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.4);
  font-size: 13px;
}
</style>
