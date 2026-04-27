<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue'

import { useAppStore } from '../store/modules/app'
import AppHeader from './components/AppHeader.vue'
import AppMain from './components/AppMain.vue'
import AppSidebar from './components/AppSidebar.vue'

const appStore = useAppStore()
const isMobile = computed(() => appStore.device === 'mobile')
const sidebarWidth = computed(() => {
  if (isMobile.value) return '220px'
  return appStore.sidebarOpened ? '220px' : '64px'
})

function handleResize() {
  appStore.syncDevice(window.innerWidth)
}

onMounted(() => {
  handleResize()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="layout-root">
    <div
      v-if="isMobile && appStore.sidebarOpened"
      class="layout-mask"
      @click="appStore.closeSidebar()"
    ></div>

    <el-container class="layout-shell">
      <el-aside
        class="layout-aside"
        :class="{
          'layout-aside--mobile': isMobile,
          'layout-aside--hidden': isMobile && !appStore.sidebarOpened,
        }"
        :width="sidebarWidth"
      >
        <AppSidebar />
      </el-aside>

      <el-container>
        <el-header class="layout-header">
          <AppHeader />
        </el-header>
        <el-main class="layout-main">
          <AppMain />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<style scoped>
.layout-root,
.layout-shell {
  min-height: 100vh;
}

.layout-root {
  position: relative;
}

.layout-mask {
  position: fixed;
  inset: 0;
  z-index: 90;
  background: rgba(0, 0, 0, 0.5);
}

.layout-aside {
  position: relative;
  z-index: 100;
  transition: width 0.2s ease;
}

.layout-aside--mobile {
  position: fixed;
  inset: 0 auto 0 0;
  height: 100vh;
  transition: transform 0.2s ease;
}

.layout-aside--hidden {
  transform: translateX(-100%);
}

.layout-header {
  height: 60px;
  padding: 0 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: #fff;
}

.layout-main {
  padding: 20px;
  background: var(--el-bg-color-page);
}

@media (max-width: 991px) {
  .layout-aside--mobile {
    box-shadow: 4px 0 12px rgba(0, 0, 0, 0.15);
  }

  .layout-main {
    padding: 16px;
  }
}
</style>
