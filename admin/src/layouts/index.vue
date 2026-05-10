<script setup lang="ts">
import { NLayout, NLayoutContent, NLayoutHeader, NLayoutSider } from 'naive-ui'
import { computed, onBeforeUnmount, onMounted } from 'vue'

import { useAppStore } from '../store/modules/app'
import AppHeader from './components/AppHeader.vue'
import AppMain from './components/AppMain.vue'
import AppSidebar from './components/AppSidebar.vue'

const appStore = useAppStore()
const isMobile = computed(() => appStore.device === 'mobile')
const collapsed = computed(() => !appStore.sidebarOpened)

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

    <NLayout has-sider class="layout-shell">
      <NLayoutSider
        bordered
        :collapsed="collapsed"
        :collapsed-width="64"
        :width="220"
        :native-scrollbar="false"
        :show-trigger="false"
        :inverted="true"
        class="layout-aside"
        :class="{
          'layout-aside--mobile': isMobile,
          'layout-aside--hidden': isMobile && !appStore.sidebarOpened,
        }"
      >
        <AppSidebar />
      </NLayoutSider>

      <NLayout>
        <NLayoutHeader bordered class="layout-header">
          <AppHeader />
        </NLayoutHeader>
        <NLayoutContent class="layout-main" :native-scrollbar="false">
          <AppMain />
        </NLayoutContent>
      </NLayout>
    </NLayout>
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

.layout-aside--mobile {
  position: fixed !important;
  inset: 0 auto 0 0;
  height: 100vh;
  z-index: 100;
  transition: transform 0.2s ease;
}

.layout-aside--hidden {
  transform: translateX(-100%);
}

.layout-header {
  height: 60px;
  padding: 0 20px;
}

.layout-main {
  padding: 20px;
  background: #f5f7fb;
  min-height: calc(100vh - 60px);
}

@media (max-width: 991px) {
  .layout-main {
    padding: 16px;
  }
}
</style>
