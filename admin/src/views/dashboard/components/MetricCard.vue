<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  title: string
  value: string
  description: string
  icon: Component
}>()
</script>

<template>
  <el-card shadow="hover" class="metric-card">
    <div class="metric-card__glow"></div>
    <div class="metric-card__icon">
      <el-icon :size="22">
        <component :is="icon" />
      </el-icon>
    </div>
    <div class="metric-card__content">
      <span>{{ title }}</span>
      <strong>{{ value }}</strong>
      <p>{{ description }}</p>
    </div>
  </el-card>
</template>

<style scoped>
.metric-card {
  position: relative;
  height: 100%;
  overflow: hidden;
  border-radius: var(--pc-card-radius-sm);
  background: rgba(255, 255, 255, 0.84);
  box-shadow: var(--pc-card-shadow);
}

.metric-card__glow {
  position: absolute;
  inset: auto auto -36px -28px;
  width: 120px;
  height: 120px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.22), rgba(245, 158, 11, 0.18));
  filter: blur(4px);
}

.metric-card :deep(.el-card__body) {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 14px;
}

.metric-card__icon {
  width: 48px;
  height: 48px;
  display: grid;
  place-items: center;
  border-radius: 16px;
  color: #ffffff;
  background: linear-gradient(135deg, var(--pc-accent-strong), var(--pc-accent));
  box-shadow: 0 12px 26px rgba(14, 165, 233, 0.26);
}

.metric-card__content {
  min-width: 0;
  display: grid;
  gap: 6px;
}

.metric-card__content span {
  color: var(--pc-subtle-text);
  font-size: 13px;
  font-weight: 600;
}

.metric-card__content strong {
  color: var(--pc-title-text);
  font-family: var(--pc-display-font);
  font-size: 30px;
  line-height: 1;
}

.metric-card__content p {
  margin: 0;
  color: var(--pc-muted-text);
  line-height: 1.6;
}
</style>
