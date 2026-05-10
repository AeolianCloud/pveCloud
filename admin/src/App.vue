<script setup lang="ts">
import { computed } from 'vue'
import { dateZhCN, zhCN, type GlobalThemeOverrides } from 'naive-ui'

import NaiveApiBridge from './components/NaiveApiBridge.vue'

const themeOverrides = computed<GlobalThemeOverrides>(() => ({
  common: {
    primaryColor: '#2563EB',
    primaryColorHover: '#1D4ED8',
    primaryColorPressed: '#1E40AF',
    primaryColorSuppl: '#1D4ED8',
    borderRadius: '6px',
  },
}))
</script>

<template>
  <n-config-provider :theme-overrides="themeOverrides" :locale="zhCN" :date-locale="dateZhCN">
    <n-loading-bar-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <n-message-provider>
            <NaiveApiBridge>
              <RouterView />
            </NaiveApiBridge>
          </n-message-provider>
        </n-notification-provider>
      </n-dialog-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>
