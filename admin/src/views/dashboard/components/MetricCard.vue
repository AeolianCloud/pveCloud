<script setup lang="ts">
import { NCard, NIcon } from 'naive-ui'
import type { Component } from 'vue'

defineProps<{
  title: string
  value: string
  icon: Component
}>()
</script>

<template>
  <NCard hoverable class="metric-card" :bordered="false">
    <div class="metric-card__row">
      <div class="metric-card__icon">
        <NIcon :size="20">
          <component :is="icon" />
        </NIcon>
      </div>
      <div class="metric-card__body">
        <div class="metric-card__title">{{ title }}</div>
        <div class="metric-card__value">{{ value }}</div>
      </div>
    </div>
  </NCard>
</template>

<style scoped>
.metric-card__row {
  display: flex;
  align-items: center;
  gap: 14px;
}

.metric-card__icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: #2563eb;
  flex-shrink: 0;
}

.metric-card__title {
  font-size: 13px;
  color: rgba(15, 23, 42, 0.6);
  margin-bottom: 2px;
}

.metric-card__value {
  font-size: 24px;
  font-weight: 700;
  color: rgba(15, 23, 42, 0.92);
  line-height: 1.2;
}
</style>
