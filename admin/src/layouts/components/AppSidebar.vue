<script setup lang="ts">
import {
  AnalyticsOutline,
  ChatbubblesOutline,
  CompassOutline,
  CubeOutline,
  DocumentTextOutline,
  FolderOpenOutline,
  PeopleOutline,
  PersonOutline,
  ReceiptOutline,
  ServerOutline,
  SettingsOutline,
  ShieldCheckmarkOutline,
  SpeedometerOutline,
} from '@vicons/ionicons5'
import { NIcon, NMenu, type MenuOption } from 'naive-ui'
import { computed, h } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { useAppStore } from '../../store/modules/app'
import { usePermissionStore } from '../../store/modules/permission'
import type { SidebarMenuItem } from '../../utils/permission'

const appStore = useAppStore()
const permissionStore = usePermissionStore()
const route = useRoute()

const iconMap: Record<string, any> = {
  Box: CubeOutline,
  Compass: CompassOutline,
  DataAnalysis: AnalyticsOutline,
  FolderOpened: FolderOpenOutline,
  Odometer: SpeedometerOutline,
  Setting: SettingsOutline,
  User: PersonOutline,
  UserFilled: PersonOutline,
  Users: PeopleOutline,
  Checked: ShieldCheckmarkOutline,
  Chatbubbles: ChatbubblesOutline,
  Tickets: ReceiptOutline,
  Server: ServerOutline,
  Document: DocumentTextOutline,
  DocumentText: DocumentTextOutline,
}

const activeMenu = computed(() => route.path)
const collapsed = computed(() => !appStore.sidebarOpened && appStore.device === 'desktop')

function resolveIcon(name: string | null) {
  if (!name) return CompassOutline
  return iconMap[name] || CompassOutline
}

function renderIcon(name: string | null) {
  return () => h(NIcon, null, { default: () => h(resolveIcon(name)) })
}

function renderLabel(item: SidebarMenuItem) {
  if (item.children?.length) {
    return item.title
  }
  return () => h(RouterLink, { to: item.path }, { default: () => item.title })
}

function toMenuOption(item: SidebarMenuItem): MenuOption {
  const opt: MenuOption = {
    label: renderLabel(item) as any,
    key: item.path,
    icon: renderIcon(item.icon),
  }
  if (item.children?.length) {
    opt.children = item.children.map((child) => ({
      label: () => h(RouterLink, { to: child.path }, { default: () => child.title }),
      key: child.path,
    }))
  }
  return opt
}

const menuOptions = computed<MenuOption[]>(() => permissionStore.sidebarMenus.map(toMenuOption))

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
        <NIcon :size="18"><AnalyticsOutline /></NIcon>
      </div>
      <span v-show="!collapsed" class="app-sidebar__name">PVE Cloud</span>
    </div>

    <NMenu
      v-if="menuOptions.length > 0"
      class="app-sidebar__menu"
      :options="menuOptions"
      :value="activeMenu"
      :collapsed="collapsed"
      :collapsed-width="64"
      :collapsed-icon-size="20"
      :indent="20"
      :inverted="true"
      :root-indent="20"
      @update:value="handleSelect"
    />

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
  background: #2563eb;
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
  min-height: 0;
  overflow: auto;
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
