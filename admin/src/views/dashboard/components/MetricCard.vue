<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  title: string
  value: string
  icon: Component
}>()
</script>

<template>
  <el-card shadow="hover" class="metric-card">
    <div class="metric-card__icon">
      <el-icon :size="20">
        <component :is="icon" />
      </el-icon>
    </div>
    <div class="metric-card__body">
      <div class="metric-card__title">{{ title }}</div>
      <div class="metric-card__value">{{ value }}</div>
    </div>
  </el-card>
</template>

<style scoped>
.metric-card :deep(.el-card__body) {
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
  background: var(--el-color-primary);
  flex-shrink: 0;
}

.metric-card__title {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 2px;
}

.metric-card__value {
  font-size: 24px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  line-height: 1.2;
}
</style>
